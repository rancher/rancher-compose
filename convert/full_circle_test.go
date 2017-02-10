package convert

import (
	"testing"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/yaml"
	"github.com/stretchr/testify/assert"
)

func testFullCircle(t *testing.T, service config.ServiceConfig) {
	launchConfig, err := ComposeToLaunchConfig(service)
	assert.Nil(t, err)

	config, hostConfig, err := LaunchConfigToDocker(launchConfig)
	assert.Nil(t, err)

	launchConfig, err = DockerToLaunchConfig(&config, &hostConfig)
	assert.Nil(t, err)

	generatedService, err := LaunchConfigToCompose(launchConfig)
	assert.Nil(t, err)

	assert.Equal(t, service, generatedService)
}

func TestFullCircle(t *testing.T) {
	testFullCircle(t, config.ServiceConfig{})

	testFullCircle(t, config.ServiceConfig{
		CapAdd:       []string{"capadd"},
		CapDrop:      []string{"capdrop"},
		CgroupParent: "cgroupparent",
		Command:      []string{"command"},
		CPUShares:    yaml.StringorInt(1),
		CPUQuota:     yaml.StringorInt(2),
		DNS:          []string{"dns"},
		DNSOpt:       []string{"dnsopt"},
		DNSSearch:    []string{"dnssearch"},
		DomainName:   "domainname",
		Entrypoint:   []string{"entrypoint"},
		ExtraHosts:   []string{"extrahosts"},
		GroupAdd:     []string{"groupadd"},
		ReadOnly:     true,
		StopSignal:   "stopsignal",
		Tty:          true,
		User:         "user",
		Uts:          "uts",
		VolumeDriver: "volumedriver",
		WorkingDir:   "workingdir",
	})
}
