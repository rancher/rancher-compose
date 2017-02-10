package convert

import (
	"github.com/docker/libcompose/utils"
	"github.com/rancher/go-rancher/v2"
)

func LaunchConfigToContainer(launchConfig client.LaunchConfig) (client.Container, error) {
	var container client.Container
	return container, utils.Convert(launchConfig, &container)
}
