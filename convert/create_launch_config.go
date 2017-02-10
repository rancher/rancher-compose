package convert

import (
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/rancher/go-rancher/v2"
)

// TODO: handle legacy load balancers
func CreateLaunchConfig(context *project.Context, name string, serviceConfig *config.ServiceConfig) (client.LaunchConfig, error) {
	launchConfig, err := ComposeToLaunchConfig(*serviceConfig)
	if err != nil {
		return client.LaunchConfig{}, err
	}

	/*if err = setupBuild(r, name, &result, serviceConfig); err != nil {
		return client.LaunchConfig{}, nil
	}*/

	if launchConfig.Labels == nil {
		launchConfig.Labels = map[string]interface{}{}
	}

	/*rancherConfig := r.RancherConfig()

	launchConfig.Kind = rancherConfig.Type
	launchConfig.Vcpu = int64(rancherConfig.Vcpu)
	launchConfig.Userdata = rancherConfig.Userdata
	launchConfig.MemoryMb = int64(rancherConfig.Memory)
	launchConfig.Disks = rancherConfig.Disks

	if strings.EqualFold(launchConfig.Kind, "virtual_machine") || strings.EqualFold(launchConfig.Kind, "virtualmachine") {
		launchConfig.Kind = "virtualMachine"
	}*/

	/*if launchConfig.LogConfig.Config == nil {
		launchConfig.LogConfig.Config = map[string]interface{}{}
	}*/

	return launchConfig, nil
}

// TODO: should this be done in ComposeToLaunchConfig?
/*func setupBuild(r *RancherService, name string, result *client.LaunchConfig, serviceConfig *config.ServiceConfig) error {
	if serviceConfig.Build.Context != "" {
		result.Build = &client.DockerBuild{
			Remote:     serviceConfig.Build.Context,
			Dockerfile: serviceConfig.Build.Dockerfile,
		}

		needBuild := true
		if config.IsValidRemote(serviceConfig.Build.Context) {
			needBuild = false
		}

		if needBuild {
			image, url, err := Upload(r.Context(), name)
			if err != nil {
				return err
			}
			logrus.Infof("Build for %s available at %s", name, url)
			serviceConfig.Build.Context = url

			if serviceConfig.Image == "" {
				serviceConfig.Image = image
			}

			result.Build = &client.DockerBuild{
				Context:    url,
				Dockerfile: serviceConfig.Build.Dockerfile,
			}
			result.ImageUuid = "docker:" + image
		} else if result.ImageUuid == "" {
			result.ImageUuid = fmt.Sprintf("docker:%s_%s_%d", r.Context().ProjectName, name, time.Now().UnixNano()/int64(time.Millisecond))
		}
	}

	return nil
}*/
