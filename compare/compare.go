package compare

import (
	"fmt"
	"reflect"

	"github.com/docker/libcompose/config"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose/convert"
)

func ComposeToLaunchConfig(serviceConfig config.ServiceConfig, launchConfig client.LaunchConfig) (bool, error) {
	convertedLaunchConfig, err := convert.ComposeToLaunchConfig(serviceConfig)
	fmt.Println(launchConfig)
	fmt.Println(convertedLaunchConfig)
	return reflect.DeepEqual(launchConfig, convertedLaunchConfig), err
}
