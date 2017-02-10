package container

import (
	"fmt"
	"os"

	"github.com/docker/libcompose/config"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose/convert"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Containers map[string]config.ServiceConfig
}

func Up(compose string, apiClient *client.RancherClient) error {
	var config Config
	if err := yaml.Unmarshal([]byte(compose), &config); err != nil {
		return err
	}

	for name, container := range config.Containers {
		launchConfig, err := convert.ComposeToLaunchConfig(container)
		if err != nil {
			return err
		}
		containerConfig, err := convert.LaunchConfigToContainer(launchConfig)
		if err != nil {
			return err
		}
		containerConfig.Name = name
		c, err := apiClient.Container.Create(&containerConfig)
		if err != nil {
			return err
		}
		fmt.Println(name, "&&", c)
	}

	fmt.Println(config.Containers)

	os.Exit(0)

	return nil
}
