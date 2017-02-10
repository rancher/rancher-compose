package convert

import (
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/runconfig/opts"
	"github.com/docker/go-connections/nat"
	"github.com/docker/libcompose/config"
	"github.com/rancher/go-rancher/v2"
)

const (
	ImagePrefix = "docker:"
)

var (
	ImageKindPattern = regexp.MustCompile("^(sim|docker):.*")
)

func ComposeToLaunchConfig(c config.ServiceConfig) (client.LaunchConfig, error) {
	/*restartPolicy, err := restartPolicy(c)
	if err != nil {
		return client.LaunchConfig{}, err
	}

	exposedPorts, portBindings, err := ports(c)
	if err != nil {
		return client.LaunchConfig{}, err
	}

	deviceMappings, err := parseDevices(c.Devices)
	if err != nil {
		return client.LaunchConfig{}, err
	}*/

	// TODO
	/*var volumesFrom []string
	if c.VolumesFrom != nil {
		volumesFrom, err = getVolumesFrom(c.VolumesFrom, ctx.Project.ServiceConfigs, ctx.ProjectName)
		if err != nil {
			return nil, nil, err
		}
	}*/

	var environment map[string]interface{}
	if len(c.Environment) > 0 {
		environment = make(map[string]interface{})
		for _, env := range c.Environment {
			split := strings.SplitN(env, "=", 2)
			if len(split) == 0 {
				environment[split[0]] = ""
			} else if len(split) == 2 {
				environment[split[0]] = split[1]
			}
		}
	}

	image := c.Image
	if c.Image != "" && !ImageKindPattern.MatchString(image) {
		image = ImagePrefix + image
	}

	var logConfig *client.LogConfig
	if c.Logging.Driver != "" {
		logConfig = &client.LogConfig{
			Driver: c.Logging.Driver,
			Config: mapStringStringToMapStringInterface(c.Logging.Options),
		}
	}

	networkMode := c.NetworkMode
	var networkLaunchConfig string
	/*if networkMode == "" {
		networkMode = "managed"
	} else if container.IpcMode(networkMode).IsContainer() {
		networkMode = "container"
		networkLaunchConfig = strings.TrimPrefix(networkMode, "container:")
	}*/

	var tmpfs map[string]interface{}
	if len(c.Tmpfs) > 0 {
		tmpfs = make(map[string]interface{})
		for _, path := range c.Tmpfs {
			split := strings.SplitN(path, ":", 2)
			if len(split) == 1 {
				tmpfs[split[0]] = ""
			} else if len(split) == 2 {
				tmpfs[split[0]] = split[1]
			}
		}
	}

	var ulimits []client.Ulimit
	if c.Ulimits.Elements != nil {
		for _, ulimit := range c.Ulimits.Elements {
			ulimits = append(ulimits, client.Ulimit{
				Name: ulimit.Name,
				Soft: ulimit.Soft,
				Hard: ulimit.Hard,
			})
		}
	}

	// TODO: hostname, volumes, data volumes, binds, mac adress, network mode, restart
	return client.LaunchConfig{
		CapAdd:       c.CapAdd,
		CapDrop:      c.CapDrop,
		CgroupParent: c.CgroupParent,
		Command:      c.Command,
		CpuShares:    int64(c.CPUShares),
		CpuQuota:     int64(c.CPUQuota),
		CpuSet:       c.CPUSet,
		DataVolumesFromLaunchConfigs: c.VolumesFrom,
		Devices:             c.Devices, // TODO: verify
		Dns:                 c.DNS,
		DnsOpt:              c.DNSOpt,
		DnsSearch:           c.DNSSearch,
		DomainName:          c.DomainName,
		EntryPoint:          c.Entrypoint,
		Environment:         environment,
		Expose:              c.Expose, // TODO: verify
		ExtraHosts:          c.ExtraHosts,
		GroupAdd:            c.GroupAdd,
		ImageUuid:           image,
		IpcMode:             c.Ipc,
		Isolation:           c.Isolation,
		Labels:              mapStringStringToMapStringInterface(c.Labels),
		LogConfig:           logConfig,
		Memory:              int64(c.MemLimit),
		MemorySwap:          int64(c.MemSwapLimit),
		MemorySwappiness:    int64(c.MemSwappiness),
		NetworkLaunchConfig: networkLaunchConfig,
		NetworkMode:         networkMode,
		OomKillDisable:      c.OomKillDisable,
		OomScoreAdj:         int64(c.OomScoreAdj),
		PidMode:             c.Pid,
		Ports:               c.Ports, // TODO: verify
		Privileged:          c.Privileged,
		ReadOnly:            c.ReadOnly,
		SecurityOpt:         c.SecurityOpt,
		ShmSize:             int64(c.ShmSize),
		StdinOpen:           c.StdinOpen,
		StopSignal:          c.StopSignal,
		Tmpfs:               tmpfs,
		Tty:                 c.Tty,
		Ulimits:             ulimits,
		User:                c.User,
		Uts:                 c.Uts,
		VolumeDriver:        c.VolumeDriver,
		WorkingDir:          c.WorkingDir,
	}, nil
}

func isBind(s string) bool {
	return strings.ContainsRune(s, ':')
}

func isVolume(s string) bool {
	return !isBind(s)
}

func restartPolicy(c *config.ServiceConfig) (*container.RestartPolicy, error) {
	restart, err := opts.ParseRestartPolicy(c.Restart)
	if err != nil {
		return nil, err
	}
	return &container.RestartPolicy{Name: restart.Name, MaximumRetryCount: restart.MaximumRetryCount}, nil
}

func ports(c *config.ServiceConfig) (map[nat.Port]struct{}, nat.PortMap, error) {
	ports, binding, err := nat.ParsePortSpecs(c.Ports)
	if err != nil {
		return nil, nil, err
	}

	exPorts, _, err := nat.ParsePortSpecs(c.Expose)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range exPorts {
		ports[k] = v
	}

	exposedPorts := map[nat.Port]struct{}{}
	for k, v := range ports {
		exposedPorts[nat.Port(k)] = v
	}

	portBindings := nat.PortMap{}
	for k, bv := range binding {
		dcbs := make([]nat.PortBinding, len(bv))
		for k, v := range bv {
			dcbs[k] = nat.PortBinding{HostIP: v.HostIP, HostPort: v.HostPort}
		}
		portBindings[nat.Port(k)] = dcbs
	}
	return exposedPorts, portBindings, nil
}

/*func getVolumesFrom(volumesFrom []string, serviceConfigs *config.ServiceConfigs, projectName string) ([]string, error) {
	volumes := []string{}
	for _, volumeFrom := range volumesFrom {
		if serviceConfig, ok := serviceConfigs.Get(volumeFrom); ok {
			// It's a service - Use the first one
			name := fmt.Sprintf("%s_%s_1", projectName, volumeFrom)
			// If a container name is specified, use that instead
			if serviceConfig.ContainerName != "" {
				name = serviceConfig.ContainerName
			}
			volumes = append(volumes, name)
		} else {
			volumes = append(volumes, volumeFrom)
		}
	}
	return volumes, nil
}*/

func parseDevices(devices []string) ([]container.DeviceMapping, error) {
	deviceMappings := []container.DeviceMapping{}
	for _, device := range devices {
		v, err := opts.ParseDevice(device)
		if err != nil {
			return nil, err
		}
		deviceMappings = append(deviceMappings, container.DeviceMapping{
			PathOnHost:        v.PathOnHost,
			PathInContainer:   v.PathInContainer,
			CgroupPermissions: v.CgroupPermissions,
		})
	}

	return deviceMappings, nil
}
