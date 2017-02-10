package convert

import (
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/rancher/go-rancher/v2"
)

// TODO: ports, volumes, hostname, restart, stdin_open, logconfig, devices, labels, expose, extra_hosts, volume driver, device*bps/iops
func DockerToLaunchConfig(config *container.Config, hostConfig *container.HostConfig) (client.LaunchConfig, error) {
	var launchConfig client.LaunchConfig

	if config != nil {
		//launchConfig.ImageUuid = "docker:" + config.Image // TODO
		launchConfig.Command = config.Cmd
		launchConfig.DomainName = config.Domainname
		launchConfig.EntryPoint = config.Entrypoint
		launchConfig.Environment = dockerToLaunchEnvironment(config.Env)
		launchConfig.StdinOpen = config.OpenStdin
		launchConfig.StopSignal = config.StopSignal
		launchConfig.Tty = config.Tty
		launchConfig.User = config.User
		launchConfig.WorkingDir = config.WorkingDir
	}
	if hostConfig != nil {
		launchConfig.BlkioWeight = int64(hostConfig.BlkioWeight)
		launchConfig.CapAdd = hostConfig.CapAdd
		launchConfig.CapDrop = hostConfig.CapDrop
		launchConfig.CgroupParent = hostConfig.CgroupParent
		launchConfig.CpuPeriod = hostConfig.CPUPeriod
		launchConfig.CpuQuota = hostConfig.CPUQuota
		launchConfig.CpuSetMems = hostConfig.CpusetMems
		launchConfig.CpuShares = hostConfig.CPUShares
		launchConfig.Dns = hostConfig.DNS
		launchConfig.DnsOpt = hostConfig.DNSOptions
		launchConfig.DnsSearch = hostConfig.DNSSearch
		launchConfig.ExtraHosts = hostConfig.ExtraHosts
		launchConfig.GroupAdd = hostConfig.GroupAdd
		launchConfig.IpcMode = string(hostConfig.IpcMode)
		launchConfig.Isolation = string(hostConfig.Isolation)
		launchConfig.KernelMemory = hostConfig.KernelMemory
		launchConfig.Memory = hostConfig.Memory
		launchConfig.MemorySwap = hostConfig.MemorySwap
		if hostConfig.MemorySwappiness != nil {
			launchConfig.MemorySwappiness = *hostConfig.MemorySwappiness
		}
		launchConfig.NetworkMode = string(hostConfig.NetworkMode)
		if hostConfig.OomKillDisable != nil {
			launchConfig.OomKillDisable = *hostConfig.OomKillDisable
		}
		launchConfig.OomScoreAdj = int64(hostConfig.OomScoreAdj)
		launchConfig.PidMode = string(hostConfig.PidMode)
		launchConfig.Privileged = hostConfig.Privileged
		launchConfig.ReadOnly = hostConfig.ReadonlyRootfs
		launchConfig.SecurityOpt = hostConfig.SecurityOpt
		launchConfig.ShmSize = hostConfig.ShmSize
		launchConfig.Sysctls = mapStringStringToMapStringInterface(hostConfig.Sysctls)
		launchConfig.Tmpfs = mapStringStringToMapStringInterface(hostConfig.Tmpfs)
		launchConfig.Uts = string(hostConfig.UTSMode)
		launchConfig.VolumeDriver = hostConfig.VolumeDriver
	}
	if config != nil && hostConfig != nil {
	}

	return launchConfig, nil
}

// TODO: doesn't handle all cases
func dockerToLaunchEnvironment(env []string) map[string]interface{} {
	if len(env) == 0 {
		return nil
	}
	envMap := map[string]interface{}{}
	for _, v := range env {
		split := strings.Split(v, "=")
		envMap[split[0]] = split[1]
	}
	return envMap
}
