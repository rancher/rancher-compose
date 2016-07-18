package rancher

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/rancher/go-rancher/client"
)

type RancherVolumesFactory struct {
	Context *Context
}

func (f *RancherVolumesFactory) Create(projectName string, volumeConfigs map[string]*config.VolumeConfig, serviceConfigs *config.ServiceConfigs, volumeEnabled bool) (project.Volumes, error) {
	volumes := make([]*Volume, 0, len(volumeConfigs))
	for name, config := range volumeConfigs {
		volume := NewVolume(projectName, name, config, f.Context)
		volumes = append(volumes, volume)
	}
	return &Volumes{
		volumes:       volumes,
		volumeEnabled: volumeEnabled,
		Context:       f.Context,
	}, nil
}

type Volumes struct {
	volumes       []*Volume
	volumeEnabled bool
	Context       *Context
}

func (v *Volumes) Initialize(ctx context.Context) error {
	if !v.volumeEnabled {
		return nil
	}
	for _, volume := range v.volumes {
		if err := volume.EnsureItExists(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (v *Volumes) Remove(ctx context.Context) error {
	if !v.volumeEnabled {
		return nil
	}
	for _, volume := range v.volumes {
		if err := volume.Remove(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Volume holds attributes and method for a volume definition in compose
type Volume struct {
	context       *Context
	name          string
	projectName   string
	driver        string
	driverOptions map[string]string
	external      bool
}

func (v *Volume) fullName() string {
	name := v.projectName + "_" + v.name
	if v.external {
		name = v.name
	}
	return name
}

func (v *Volume) Inspect(ctx context.Context) (*client.Volume, error) {
	volumes, err := v.context.Client.Volume.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name": v.fullName(),
		},
	})
	if err != nil {
		return nil, err
	}

	if len(volumes.Data) > 0 {
		return &volumes.Data[0], nil
	}

	return nil, nil
}

func (v *Volume) Remove(ctx context.Context) error {
	volumeResource, err := v.Inspect(ctx)
	if err != nil {
		return err
	}
	err = v.context.Client.Volume.Delete(volumeResource)
	return err
}

func (v *Volume) EnsureItExists(ctx context.Context) error {
	volumeResource, err := v.Inspect(ctx)
	if err != nil {
		return err
	}
	if volumeResource == nil {
		return v.create(ctx)
	}
	if v.driver != "" && volumeResource.Driver != v.driver {
		return fmt.Errorf("Volume %q needs to be recreated - driver has changed", v.name)
	}
	return nil
}

func (v *Volume) create(ctx context.Context) error {
	driverOptions := map[string]interface{}{}
	for k, v := range v.driverOptions {
		driverOptions[k] = v
	}
	_, err := v.context.Client.Volume.Create(&client.Volume{
		Name:       v.fullName(),
		Driver:     v.driver,
		DriverOpts: driverOptions,
	})
	return err
}

func NewVolume(projectName, name string, config *config.VolumeConfig, context *Context) *Volume {
	return &Volume{
		context:       context,
		name:          name,
		projectName:   projectName,
		driver:        config.Driver,
		driverOptions: config.DriverOpts,
		external:      config.External.External,
	}
}
