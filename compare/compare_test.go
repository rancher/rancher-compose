package compare

import (
	"testing"

	"github.com/docker/libcompose/config"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose/convert"
	"github.com/stretchr/testify/assert"
)

func testComposeToLaunchConfig(t *testing.T, service config.ServiceConfig, launchConfig client.LaunchConfig) {
	same, err := ComposeToLaunchConfig(service, launchConfig)
	assert.Nil(t, err)
	assert.True(t, same)
}

func TestComposeToLaunchConfig(t *testing.T) {
	convert.ComposeToLaunchConfig
	testComposeToLaunchConfig(t, config.ServiceConfig{}, cli)
}
