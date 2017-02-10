package convert

import (
	"fmt"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/yaml"
	"github.com/rancher/go-rancher/v2"
)

func LaunchConfigToCompose(launchConfig client.LaunchConfig) (config.ServiceConfig, error) {
	var logging config.Log
	if launchConfig.LogConfig != nil {
		logging = config.Log{
			Driver:  launchConfig.LogConfig.Driver,
			Options: mapStringInterfaceToStringString(launchConfig.LogConfig.Config),
		}
	}

	var tmpfs []string
	for device, options := range launchConfig.Tmpfs {
		tmpfs = append(tmpfs, device, fmt.Sprint(options))
	}

	var ulimits yaml.Ulimits
	for _, ulimit := range launchConfig.Ulimits {
		ulimits.Elements = append(ulimits.Elements, yaml.NewUlimit(ulimit.Name, ulimit.Soft, ulimit.Hard))
	}

	return config.ServiceConfig{
		CapAdd:         launchConfig.CapAdd,
		CapDrop:        launchConfig.CapDrop,
		CgroupParent:   launchConfig.CgroupParent,
		Command:        launchConfig.Command,
		CPUShares:      yaml.StringorInt(launchConfig.CpuShares),
		CPUQuota:       yaml.StringorInt(launchConfig.CpuQuota),
		CPUSet:         launchConfig.CpuSet,
		Devices:        launchConfig.Devices,
		DNS:            launchConfig.Dns,
		DNSOpt:         launchConfig.DnsOpt,
		DNSSearch:      launchConfig.DnsSearch,
		DomainName:     launchConfig.DomainName,
		Entrypoint:     launchConfig.EntryPoint,
		Expose:         launchConfig.Expose,
		ExtraHosts:     launchConfig.ExtraHosts,
		GroupAdd:       launchConfig.GroupAdd,
		Ipc:            launchConfig.IpcMode,
		Isolation:      launchConfig.Isolation,
		Labels:         mapStringInterfaceToStringString(launchConfig.Labels),
		Logging:        logging,
		MemLimit:       yaml.StringorInt(launchConfig.Memory),
		MemSwapLimit:   yaml.StringorInt(launchConfig.MemorySwap),
		MemSwappiness:  yaml.StringorInt(launchConfig.MemorySwappiness),
		NetworkMode:    launchConfig.NetworkMode,
		OomKillDisable: launchConfig.OomKillDisable,
		OomScoreAdj:    yaml.StringorInt(launchConfig.OomScoreAdj),
		Pid:            launchConfig.PidMode,
		Ports:          launchConfig.Ports, // TODO: verify
		Privileged:     launchConfig.Privileged,
		ReadOnly:       launchConfig.ReadOnly,
		SecurityOpt:    launchConfig.SecurityOpt,
		ShmSize:        yaml.StringorInt(launchConfig.ShmSize),
		StdinOpen:      launchConfig.StdinOpen,
		StopSignal:     launchConfig.StopSignal,
		Tmpfs:          tmpfs,
		Tty:            launchConfig.Tty,
		Ulimits:        ulimits,
		User:           launchConfig.User,
		Uts:            launchConfig.Uts,
		VolumeDriver:   launchConfig.VolumeDriver,
		WorkingDir:     launchConfig.WorkingDir,
	}, nil
}
