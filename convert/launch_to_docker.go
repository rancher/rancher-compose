package convert

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/rancher/go-rancher/v2"
)

func LaunchConfigToDocker(launchConfig client.LaunchConfig) (container.Config, container.HostConfig, error) {
	var config container.Config
	config.Cmd = launchConfig.Command
	config.Domainname = launchConfig.DomainName
	config.Entrypoint = launchConfig.EntryPoint
	config.Env = launchToDockerEnvironment(launchConfig.Environment)
	config.OpenStdin = launchConfig.StdinOpen
	config.StopSignal = launchConfig.StopSignal
	config.Tty = launchConfig.Tty
	config.User = launchConfig.User
	config.WorkingDir = launchConfig.WorkingDir

	var hostConfig container.HostConfig
	hostConfig.BlkioWeight = uint16(launchConfig.BlkioWeight)
	hostConfig.CapAdd = launchConfig.CapAdd
	hostConfig.CapDrop = launchConfig.CapDrop
	hostConfig.CgroupParent = launchConfig.CgroupParent
	hostConfig.CPUPeriod = launchConfig.CpuPeriod
	hostConfig.CPUQuota = launchConfig.CpuQuota
	hostConfig.CpusetMems = launchConfig.CpuSetMems
	hostConfig.CPUShares = launchConfig.CpuShares
	hostConfig.DNS = launchConfig.Dns
	hostConfig.DNSOptions = launchConfig.DnsOpt
	hostConfig.DNSSearch = launchConfig.DnsSearch
	hostConfig.ExtraHosts = launchConfig.ExtraHosts
	hostConfig.GroupAdd = launchConfig.GroupAdd
	hostConfig.IpcMode = container.IpcMode(launchConfig.IpcMode)
	hostConfig.Isolation = container.Isolation(launchConfig.Isolation)
	hostConfig.KernelMemory = launchConfig.KernelMemory
	hostConfig.Memory = launchConfig.Memory
	hostConfig.MemorySwap = launchConfig.MemorySwap
	if launchConfig.MemorySwappiness != 0 {
		hostConfig.MemorySwappiness = &launchConfig.MemorySwappiness
	}
	hostConfig.NetworkMode = container.NetworkMode(launchConfig.NetworkMode)
	if launchConfig.OomKillDisable {
		hostConfig.OomKillDisable = &launchConfig.OomKillDisable
	}
	hostConfig.OomScoreAdj = int(launchConfig.OomScoreAdj)
	hostConfig.PidMode = container.PidMode(launchConfig.PidMode)
	hostConfig.Privileged = launchConfig.Privileged
	hostConfig.ReadonlyRootfs = launchConfig.ReadOnly
	hostConfig.SecurityOpt = launchConfig.SecurityOpt
	hostConfig.ShmSize = launchConfig.ShmSize
	hostConfig.Sysctls = mapStringInterfaceToStringString(launchConfig.Sysctls)
	hostConfig.Tmpfs = mapStringInterfaceToStringString(launchConfig.Tmpfs)
	hostConfig.UTSMode = container.UTSMode(launchConfig.Uts)
	hostConfig.VolumeDriver = launchConfig.VolumeDriver

	return config, hostConfig, nil
}

// TODO: verify this is correct and see if logic exists elsewhere
func launchToDockerEnvironment(env map[string]interface{}) []string {
	var envSlice []string
	for k, v := range env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}
	return envSlice
}
