package convert

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/rancher/go-rancher/v2"
	"github.com/stretchr/testify/assert"
)

var (
	fullConfig = container.Config{
		Cmd:        []string{"cmd"},
		Domainname: "domainname",
		Entrypoint: []string{"entrypoint"},
		Env:        []string{"A=B"},
		OpenStdin:  true,
		StopSignal: "stop",
		Tty:        true,
		User:       "user",
		WorkingDir: "workingdir",
	}
	fullHostConfig = container.HostConfig{
		Resources: container.Resources{
			BlkioWeight:      uint16(1),
			CPUPeriod:        int64(2),
			CPUQuota:         int64(3),
			CpusetMems:       "cpusetmems",
			CPUShares:        int64(4),
			KernelMemory:     int64(5),
			Memory:           int64(6),
			MemorySwap:       int64(7),
			MemorySwappiness: &[]int64{int64(8)}[0],
			OomKillDisable:   &[]bool{true}[0],
		},
		CapAdd:         []string{"capadd"},
		CapDrop:        []string{"capdrop"},
		DNSOptions:     []string{"dnsoptions"},
		DNSSearch:      []string{"dnssearch"},
		ExtraHosts:     []string{"extrahosts"},
		GroupAdd:       []string{"groupadd"},
		IpcMode:        "ipcmode",
		NetworkMode:    "networkmode",
		OomScoreAdj:    9,
		PidMode:        "pidmode",
		Privileged:     true,
		ReadonlyRootfs: true,
		SecurityOpt:    []string{"securityopt"},
		ShmSize:        10,
		Sysctls:        map[string]string{"sysctlk": "sysctlv"},
		Tmpfs:          map[string]string{"tmpfsk": "tmpfsv"},
		UTSMode:        "utsmode",
		VolumeDriver:   "volumedriver",
	}
	fullLaunchConfig = client.LaunchConfig{
		BlkioWeight:      int64(1),
		Command:          []string{"cmd"},
		CapAdd:           []string{"capadd"},
		CapDrop:          []string{"capdrop"},
		CpuPeriod:        int64(2),
		CpuQuota:         int64(3),
		CpuSetMems:       "cpusetmems",
		CpuShares:        int64(4),
		DnsOpt:           []string{"dnsoptions"},
		DnsSearch:        []string{"dnssearch"},
		DomainName:       "domainname",
		EntryPoint:       []string{"entrypoint"},
		Environment:      map[string]interface{}{"A": "B"},
		ExtraHosts:       []string{"extrahosts"},
		GroupAdd:         []string{"groupadd"},
		IpcMode:          "ipcmode",
		KernelMemory:     int64(5),
		Memory:           int64(6),
		MemorySwap:       int64(7),
		MemorySwappiness: int64(8),
		NetworkMode:      "networkmode",
		OomKillDisable:   true,
		OomScoreAdj:      9,
		PidMode:          "pidmode",
		Privileged:       true,
		ReadOnly:         true,
		SecurityOpt:      []string{"securityopt"},
		ShmSize:          10,
		StdinOpen:        true,
		StopSignal:       "stop",
		Sysctls:          map[string]interface{}{"sysctlk": "sysctlv"},
		Tmpfs:            map[string]interface{}{"tmpfsk": "tmpfsv"},
		Tty:              true,
		User:             "user",
		Uts:              "utsmode",
		WorkingDir:       "workingdir",
		VolumeDriver:     "volumedriver",
	}
)

func testDockerToLaunchConfig(t *testing.T, config container.Config, hostConfig container.HostConfig, expectedLaunchConfig client.LaunchConfig) {
	launchConfig, err := DockerToLaunchConfig(&config, &hostConfig)
	assert.Nil(t, err)
	assert.Equal(t, expectedLaunchConfig, launchConfig)
}

func TestDockerToLaunchConfig(t *testing.T) {
	testDockerToLaunchConfig(t, container.Config{}, container.HostConfig{}, client.LaunchConfig{})
	testDockerToLaunchConfig(t, fullConfig, fullHostConfig, fullLaunchConfig)
}
