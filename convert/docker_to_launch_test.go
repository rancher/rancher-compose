package convert

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/rancher/go-rancher/v2"
	"github.com/stretchr/testify/assert"
)

func testLaunchConfigToDocker(t *testing.T, launchConfig client.LaunchConfig, expectedConfig container.Config, expectedHostConfig container.HostConfig) {
	config, hostConfig, err := LaunchConfigToDocker(launchConfig)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, config)
	assert.Equal(t, expectedHostConfig, hostConfig)
}

func TestLaunchConfigToDocker(t *testing.T) {
	testLaunchConfigToDocker(t, client.LaunchConfig{}, container.Config{}, container.HostConfig{})
	testLaunchConfigToDocker(t, fullLaunchConfig, fullConfig, fullHostConfig)
}
