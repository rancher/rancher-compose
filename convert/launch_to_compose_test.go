package convert

import (
	"testing"

	"github.com/docker/libcompose/config"
	"github.com/rancher/go-rancher/v2"
	"github.com/stretchr/testify/assert"
)

func testLaunchConfigToCompose(t *testing.T, launchConfig client.LaunchConfig, expectedService config.ServiceConfig) {
	service, err := LaunchConfigToCompose(launchConfig)
	assert.Nil(t, err)
	assert.Equal(t, expectedService, service)
}

func TestLaunchConfigToCompose(t *testing.T) {
	testComposeToLaunchConfig(t, config.ServiceConfig{}, client.LaunchConfig{})
	/*testComposeToLaunchConfig(t, config.ServiceConfig{
		Environment: []string{"A=B", "C=", "D"},
	}, client.LaunchConfig{
		Environment: map[string]interface{}{
			"A": "B",
			"C": "",
			"D": "",
		},
	})*/
}
