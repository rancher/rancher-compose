package convert

import (
	"testing"

	"github.com/docker/libcompose/config"
	"github.com/rancher/go-rancher/v2"
	"github.com/stretchr/testify/assert"
)

func testComposeToLaunchConfig(t *testing.T, service config.ServiceConfig, expectedLaunchConfig client.LaunchConfig) {
	launchConfig, err := ComposeToLaunchConfig(service)
	assert.Nil(t, err)
	assert.Equal(t, expectedLaunchConfig, launchConfig)
}

func TestComposeToLaunchConfig(t *testing.T) {
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
