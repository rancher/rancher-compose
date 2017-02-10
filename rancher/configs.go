package rancher

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/utils"
	"github.com/docker/libcompose/yaml"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose/convert"
)

func createLaunchConfigs(r *RancherService) (client.LaunchConfig, []client.SecondaryLaunchConfig, error) {
	secondaryLaunchConfigs := []client.SecondaryLaunchConfig{}
	launchConfig, err := createLaunchConfig(r, r.Name(), r.serviceConfig)
	if err != nil {
		return launchConfig, nil, err
	}
	launchConfig.HealthCheck = r.HealthCheck("")

	secondaries, ok := r.Context().SidekickInfo.primariesToSidekicks[r.Name()]
	if !ok {
		return launchConfig, []client.SecondaryLaunchConfig{}, nil
	}

	for _, secondaryName := range secondaries {
		serviceConfig, ok := r.Context().Project.ServiceConfigs.Get(secondaryName)
		if !ok {
			return launchConfig, nil, fmt.Errorf("Failed to find sidekick: %s", secondaryName)
		}

		launchConfig, err := createLaunchConfig(r, secondaryName, serviceConfig)
		if err != nil {
			return launchConfig, nil, err
		}
		launchConfig.HealthCheck = r.HealthCheck(secondaryName)

		var secondaryLaunchConfig client.SecondaryLaunchConfig
		if err = utils.Convert(launchConfig, &secondaryLaunchConfig); err != nil {
			return client.LaunchConfig{}, nil, err
		}
		secondaryLaunchConfig.Name = secondaryName

		if secondaryLaunchConfig.Labels == nil {
			secondaryLaunchConfig.Labels = map[string]interface{}{}
		}
		secondaryLaunchConfigs = append(secondaryLaunchConfigs, secondaryLaunchConfig)
	}

	return launchConfig, secondaryLaunchConfigs, nil
}

// TODO: handle legacy load balancers
func createLaunchConfig(r *RancherService, name string, serviceConfig *config.ServiceConfig) (client.LaunchConfig, error) {
	tempImage := serviceConfig.Image
	tempLabels := serviceConfig.Labels

	newLabels := yaml.SliceorMap{}
	if serviceConfig.Image == "rancher/load-balancer-service" {
		// Lookup default load balancer image
		lbImageSetting, err := r.Client().Setting.ById("lb.instance.image")
		if err != nil {
			return client.LaunchConfig{}, err
		}
		serviceConfig.Image = lbImageSetting.Value

		// Strip off legacy load balancer labels
		for k, v := range serviceConfig.Labels {
			if !strings.HasPrefix(k, "io.rancher.loadbalancer") && !strings.HasPrefix(k, "io.rancher.service.selector") {
				newLabels[k] = v
			}
		}
		serviceConfig.Labels = newLabels
	}

	launchConfig, err := convert.ComposeToLaunchConfig(*serviceConfig)
	if err != nil {
		return client.LaunchConfig{}, err
	}

	serviceConfig.Image = tempImage
	serviceConfig.Labels = tempLabels

	if err = setupBuild(r, name, &launchConfig, serviceConfig); err != nil {
		return client.LaunchConfig{}, nil
	}

	// TODO: should this be done in ComposeToLaunchConfig?
	if launchConfig.Labels == nil {
		launchConfig.Labels = map[string]interface{}{}
	}

	// TODO: should this be done in ComposeToLaunchConfig?
	if launchConfig.LogConfig != nil && launchConfig.LogConfig.Config == nil {
		launchConfig.LogConfig.Config = map[string]interface{}{}
	}

	rancherConfig := r.RancherConfig()

	launchConfig.Kind = rancherConfig.Type
	launchConfig.Vcpu = int64(rancherConfig.Vcpu)
	launchConfig.Userdata = rancherConfig.Userdata
	launchConfig.MemoryMb = int64(rancherConfig.Memory)
	launchConfig.Disks = rancherConfig.Disks

	if strings.EqualFold(launchConfig.Kind, "virtual_machine") || strings.EqualFold(launchConfig.Kind, "virtualmachine") {
		launchConfig.Kind = "virtualMachine"
	}

	return launchConfig, nil
}

// TODO: should this be done in ComposeToLaunchConfig?
func setupBuild(r *RancherService, name string, result *client.LaunchConfig, serviceConfig *config.ServiceConfig) error {
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
}
