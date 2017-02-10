package convert

import (
	"fmt"
	"testing"

	"github.com/rancher/go-rancher/v2"
)

func TestLaunchConfigToContainer(t *testing.T) {
	container, err := LaunchConfigToContainer(client.LaunchConfig{
		ImageUuid: "docker:nginx",
	})

	fmt.Println(err, container)
}
