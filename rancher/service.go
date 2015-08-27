package rancher

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/runconfig"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"
	"github.com/gorilla/websocket"
	rancherClient "github.com/rancher/go-rancher/client"
	"github.com/rancher/go-rancher/hostaccess"
)

const (
	LB_IMAGE       = "rancher/load-balancer-service"
	DNS_IMAGE      = "rancher/dns-service"
	EXTERNAL_IMAGE = "rancher/external-service"
)

type Link struct {
	ServiceName, Alias string
}

type IsDone func(*rancherClient.Resource) (bool, error)

type ContainerInspect struct {
	Name       string
	Config     *runconfig.Config
	HostConfig *runconfig.HostConfig
}

type RancherService struct {
	name          string
	serviceConfig *project.ServiceConfig
	context       *Context
}

func (r *RancherService) Name() string {
	return r.name
}

func (r *RancherService) Config() *project.ServiceConfig {
	return r.serviceConfig
}

func NewService(name string, config *project.ServiceConfig, context *Context) *RancherService {
	return &RancherService{
		name:          name,
		serviceConfig: config,
		context:       context,
	}
}

func (r *RancherService) RancherService() (*rancherClient.Service, error) {
	return r.findExisting(r.name)
}

func (r *RancherService) Create() error {
	service, err := r.findExisting(r.name)

	if err == nil && service == nil {
		service, err = r.createService()
	}

	return err
}

func (r *RancherService) Start() error {
	return r.up(false)
}

func (r *RancherService) Up() error {
	return r.up(true)
}

func (r *RancherService) Build() error {
	return nil
}

func (r *RancherService) up(create bool) error {
	var err error

	defer func() {
		if err == nil && r.context.Log {
			go r.Log()
		}
	}()

	service, err := r.findExisting(r.name)
	if err != nil {
		return err
	}

	if service == nil && !create {
		return nil
	}

	if service == nil {
		service, err = r.createService()
	} else {
		err = r.setupLinks(service)
	}

	if err != nil {
		return err
	}

	if service.State == "active" {
		return nil
	}

	if service.Actions["activate"] != "" {
		service, err = r.context.Client.Service.ActionActivate(service)
		err = r.Wait(service)
	}

	return err
}

func (r *RancherService) Down() error {
	service, err := r.findExisting(r.name)

	if err == nil && service == nil {
		return nil
	}

	if err != nil {
		return err
	}

	if service.State == "inactive" {
		return nil
	}

	service, err = r.context.Client.Service.ActionDeactivate(service)
	return r.Wait(service)
}

func (r *RancherService) Delete() error {
	service, err := r.findExisting(r.name)

	if err == nil && service == nil {
		return nil
	}

	if err != nil {
		return err
	}

	if service.Removed != "" || service.State == "removing" || service.State == "removed" {
		return nil
	}

	err = r.context.Client.Service.Delete(service)
	if err != nil {
		return err
	}

	return r.Wait(service)
}

func (r *RancherService) resolveServiceAndEnvironmentId(name string) (string, string, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 1 {
		return name, r.context.Environment.Id, nil
	}

	envs, err := r.context.Client.Environment.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"name":         parts[0],
			"removed_null": nil,
		},
	})

	if err != nil {
		return "", "", err
	}

	if len(envs.Data) == 0 {
		return "", "", fmt.Errorf("Failed to find stack: %s", parts[0])
	}

	return parts[1], envs.Data[0].Id, nil
}

func (r *RancherService) findExisting(name string) (*rancherClient.Service, error) {
	logrus.Debugf("Finding service %s", name)

	name, environmentId, err := r.resolveServiceAndEnvironmentId(name)
	if err != nil {
		return nil, err
	}

	services, err := r.context.Client.Service.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"environmentId": environmentId,
			"name":          name,
			"removed_null":  nil,
		},
	})

	if err != nil {
		return nil, err
	}

	if len(services.Data) == 0 {
		return nil, nil
	}

	logrus.Debugf("Found service %s", name)
	return &services.Data[0], nil
}

func (r *RancherService) createExternalService() (*rancherClient.Service, error) {
	config, _ := r.context.RancherConfig[r.name]

	_, err := r.context.Client.ExternalService.Create(&rancherClient.ExternalService{
		Name:                r.name,
		ExternalIpAddresses: config.ExternalIps,
		Hostname:            config.Hostname,
		EnvironmentId:       r.context.Environment.Id,
	})

	if err != nil {
		return nil, err
	}

	return r.findExisting(r.name)
}

func (r *RancherService) createDnsService() (*rancherClient.Service, error) {
	_, err := r.context.Client.DnsService.Create(&rancherClient.DnsService{
		Name:          r.name,
		EnvironmentId: r.context.Environment.Id,
	})

	if err != nil {
		return nil, err
	}

	links, err := r.getServiceLinks()
	if err != nil {
		return nil, err
	}

	service, err := r.findExisting(r.name)
	if err != nil {
		return nil, err
	}

	if len(links) > 0 {
		_, err = r.context.Client.Service.ActionSetservicelinks(service, &rancherClient.SetServiceLinksInput{
			ServiceLinks: links,
		})
	}

	return service, nil
}

func (r *RancherService) createLbService() (*rancherClient.Service, error) {
	var lbConfig *rancherClient.LoadBalancerConfig

	if config, ok := r.context.RancherConfig[r.name]; ok {
		lbConfig = config.LoadBalancerConfig
	}

	launchConfig, err := r.createLaunchConfig(r.serviceConfig)
	if err != nil {
		return nil, err
	}

	launchConfig.ImageUuid = ""
	// Write back to the ports passed in because the Docker parsing logic changes then
	launchConfig.Ports = r.serviceConfig.Ports
	launchConfig.Expose = r.serviceConfig.Expose

	_, err = r.context.Client.LoadBalancerService.Create(&rancherClient.LoadBalancerService{
		Name:               r.name,
		LoadBalancerConfig: lbConfig,
		LaunchConfig:       launchConfig,
		Scale:              int64(r.getConfiguredScale()),
		EnvironmentId:      r.context.Environment.Id,
	})

	if err != nil {
		return nil, err
	}

	return r.findExisting(r.name)
}

func (r *RancherService) createNormalService() (*rancherClient.Service, error) {
	secondaryLaunchConfigs := []interface{}{}

	launchConfig, err := r.createLaunchConfig(r.serviceConfig)
	if err != nil {
		return nil, err
	}

	if secondaries, ok := r.context.SidekickInfo.primariesToSidekicks[r.name]; ok {
		for _, secondaryName := range secondaries {
			serviceConfig, ok := r.context.Project.Configs[secondaryName]
			if !ok {
				return nil, fmt.Errorf("Failed to find sidekick: %s", secondaryName)
			}

			launchConfig, err := r.createLaunchConfig(serviceConfig)
			if err != nil {
				return nil, err
			}

			var secondaryLaunchConfig rancherClient.SecondaryLaunchConfig
			utils.Convert(launchConfig, &secondaryLaunchConfig)
			secondaryLaunchConfig.Name = secondaryName

			secondaryLaunchConfigs = append(secondaryLaunchConfigs, secondaryLaunchConfig)
		}
	}

	return r.context.Client.Service.Create(&rancherClient.Service{
		Name:                   r.name,
		LaunchConfig:           launchConfig,
		SecondaryLaunchConfigs: secondaryLaunchConfigs,
		Scale:         int64(r.getConfiguredScale()),
		EnvironmentId: r.context.Environment.Id,
	})
}

func (r *RancherService) getHealthCheck() *rancherClient.InstanceHealthCheck {
	if config, ok := r.context.RancherConfig[r.name]; ok {
		return config.HealthCheck
	}

	return nil
}

func (r *RancherService) getConfiguredScale() int {
	scale := 1
	if config, ok := r.context.RancherConfig[r.name]; ok {
		if config.Scale > 0 {
			scale = config.Scale
		}
	}

	return scale
}

func (r *RancherService) createService() (*rancherClient.Service, error) {
	logrus.Infof("Creating service %s", r.name)

	rancherConfig, _ := r.context.RancherConfig[r.name]
	var service *rancherClient.Service
	var err error

	if len(rancherConfig.ExternalIps) > 0 || rancherConfig.Hostname != "" {
		service, err = r.createExternalService()
	} else if r.serviceConfig.Image == LB_IMAGE {
		service, err = r.createLbService()
	} else if r.serviceConfig.Image == DNS_IMAGE {
		service, err = r.createDnsService()
	} else {
		service, err = r.createNormalService()
	}

	if err != nil {
		return nil, err
	}

	if err := r.setupLinks(service); err != nil {
		return nil, err
	}

	err = r.Wait(service)
	return service, err
}

func (r *RancherService) setupLinks(service *rancherClient.Service) error {
	var err error
	var links []interface{}

	if service.Type == rancherClient.LOAD_BALANCER_SERVICE_TYPE {
		links, err = r.getLbLinks()
	} else {
		links, err = r.getServiceLinks()
	}

	if err == nil && len(links) > 0 {
		_, err = r.context.Client.Service.ActionSetservicelinks(service, &rancherClient.SetServiceLinksInput{
			ServiceLinks: links,
		})
	}
	return err
}

func (r *RancherService) getLbLinks() ([]interface{}, error) {
	links, err := r.getLinks()
	if err != nil {
		return nil, err
	}

	result := []interface{}{}
	for link, id := range links {
		ports, err := r.getLbLinkPorts(link.ServiceName)
		if err != nil {
			return nil, err
		}

		result = append(result, rancherClient.LoadBalancerServiceLink{
			Ports:     ports,
			ServiceId: id,
		})
	}

	return result, nil
}

func (r *RancherService) getLbLinkPorts(name string) ([]string, error) {
	labelName := "io.rancher.loadbalancer.target." + name
	v := r.serviceConfig.Labels.MapParts()[labelName]
	if v == "" {
		return []string{}, nil
	}

	return TrimSplit(v, ",", -1), nil
}

func (r *RancherService) getServiceLinks() ([]interface{}, error) {
	links, err := r.getLinks()
	if err != nil {
		return nil, err
	}

	result := []interface{}{}
	for link, id := range links {
		result = append(result, rancherClient.ServiceLink{
			Name:      link.Alias,
			ServiceId: id,
		})
	}

	return result, nil
}

func (r *RancherService) getLinks() (map[Link]string, error) {
	result := map[Link]string{}

	for _, link := range append(r.serviceConfig.Links.Slice(), r.serviceConfig.ExternalLinks...) {
		parts := strings.SplitN(link, ":", 2)
		name := parts[0]
		alias := name
		if len(parts) == 2 {
			alias = parts[1]
		}

		name = strings.TrimSpace(name)
		alias = strings.TrimSpace(alias)

		linkedService, err := r.findExisting(name)
		if err != nil {
			return nil, err
		}

		if linkedService == nil {
			logrus.Warnf("Failed to find service %s to link to", name)
		} else {
			result[Link{
				ServiceName: name,
				Alias:       alias,
			}] = linkedService.Id
		}
	}

	return result, nil
}

func setupNetworking(netMode string, launchConfig *rancherClient.LaunchConfig) {
	if netMode == "" {
		launchConfig.NetworkMode = "managed"
	} else if runconfig.IpcMode(netMode).IsContainer() {
		// For some reason NetworkMode object is gone runconfig, but IpcMode works the same for this
		launchConfig.NetworkMode = "container"
		launchConfig.NetworkLaunchConfig = strings.TrimPrefix(netMode, "container:")
	} else {
		launchConfig.NetworkMode = netMode
	}
}

func setupVolumesFrom(volumesFrom []string, launchConfig *rancherClient.LaunchConfig) {
	launchConfig.DataVolumesFromLaunchConfigs = volumesFrom
}

func (r *RancherService) setupBuild(result *rancherClient.LaunchConfig, serviceConfig *project.ServiceConfig) error {
	if serviceConfig.Build != "" {
		result.Build = &rancherClient.DockerBuild{
			Remote:     serviceConfig.Build,
			Dockerfile: serviceConfig.Dockerfile,
		}

		needBuild := true
		for _, remote := range project.ValidRemotes {
			if strings.HasPrefix(serviceConfig.Build, remote) {
				needBuild = false
				break
			}
		}

		if needBuild {
			image, url, err := Upload(r.context.Project, r.name)
			if err != nil {
				return err
			}
			logrus.Infof("Build for %s available at %s", r.name, url)
			serviceConfig.Build = url

			if serviceConfig.Image == "" {
				serviceConfig.Image = image
			}

			result.Build = &rancherClient.DockerBuild{
				Context:    url,
				Dockerfile: serviceConfig.Dockerfile,
			}
			result.ImageUuid = "docker:" + image
		} else if result.ImageUuid == "" {
			result.ImageUuid = fmt.Sprintf("docker:%s_%s_%d", r.context.ProjectName, r.name, time.Now().UnixNano()/int64(time.Millisecond))
		}
	}

	return nil
}

func (r *RancherService) createLaunchConfig(serviceConfig *project.ServiceConfig) (rancherClient.LaunchConfig, error) {
	var result rancherClient.LaunchConfig

	schemasUrl := strings.SplitN(r.context.Client.Schemas.Links["self"], "/schemas", 2)[0]
	scriptsUrl := schemasUrl + "/scripts/transform"

	config, hostConfig, err := docker.Convert(serviceConfig)
	if err != nil {
		return result, err
	}

	dockerContainer := &ContainerInspect{
		Config:     config,
		HostConfig: hostConfig,
	}

	dockerContainer.HostConfig.NetworkMode = runconfig.NetworkMode("")

	if serviceConfig.Name != "" {
		dockerContainer.Name = "/" + serviceConfig.Name
	} else {
		dockerContainer.Name = "/" + r.name
	}

	err = r.context.Client.Post(scriptsUrl, dockerContainer, &result)
	if err != nil {
		return result, err
	}

	result.HealthCheck = r.getHealthCheck()

	setupNetworking(serviceConfig.Net, &result)
	setupVolumesFrom(serviceConfig.VolumesFrom, &result)

	err = r.setupBuild(&result, serviceConfig)
	return result, err
}

func (r *RancherService) Wait(service *rancherClient.Service) error {
	for {
		if service.Transitioning != "yes" {
			return nil
		}

		time.Sleep(150 * time.Millisecond)

		err := r.context.Client.Reload(&service.Resource, service)
		if err != nil {
			return err
		}
	}
}

func (r *RancherService) waitInstance(service *rancherClient.Instance) error {
	for {
		if service.Transitioning != "yes" {
			return nil
		}

		time.Sleep(150 * time.Millisecond)

		err := r.context.Client.Reload(&service.Resource, service)
		if err != nil {
			return err
		}
	}
}

func (r *RancherService) Scale(count int) error {
	service, err := r.findExisting(r.name)
	if err != nil {
		return err
	}

	if service == nil {
		return fmt.Errorf("Failed to find %s to scale", r.name)
	}

	service, err = r.context.Client.Service.Update(service, map[string]interface{}{
		"scale": count,
	})
	if err != nil {
		return err
	}

	return r.Wait(service)
}

func (r *RancherService) Containers() ([]project.Container, error) {
	result := []project.Container{}

	containers, err := r.containers()
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		name := c.Name
		if name == "" {
			name = c.Uuid
		}
		result = append(result, NewContainer(c.Id, name))
	}

	return result, nil
}

func (r *RancherService) containers() ([]rancherClient.Container, error) {
	service, err := r.findExisting(r.name)
	if err != nil {
		return nil, err
	}

	var instances rancherClient.ContainerCollection

	err = r.context.Client.GetLink(service.Resource, "instances", &instances)
	if err != nil {
		return nil, err
	}

	return instances.Data, nil
}

func (r *RancherService) Restart() error {
	containers, err := r.containers()
	if err != nil {
		return err
	}

	for _, container := range containers {
		logrus.Infof("Restarting container: %s", container.Name)
		instance, err := r.context.Client.Container.ActionRestart(&container)
		if err != nil {
			return err
		}

		r.waitInstance(instance)
		if instance.State != "running" {
			return fmt.Errorf("Failed to restart %s, in state: %s", instance.Name, instance.State)
		}
	}

	return nil
}

func (r *RancherService) Log() error {
	service, err := r.findExisting(r.name)
	if err != nil {
		return err
	}

	if service.Type != "service" {
		return nil
	}

	containers, err := r.containers()
	if err != nil {
		logrus.Errorf("Failed to list containers to log: %v", err)
		return err
	}

	for _, container := range containers {
		conn, err := (*hostaccess.RancherWebsocketClient)(r.context.Client).GetHostAccess(container.Resource, "logs", nil)
		if err != nil {
			logrus.Errorf("Failed to get logs for %s: %v", container.Name, err)
			continue
		}

		go r.pipeLogs(&container, conn)
	}

	return nil
}

func (r *RancherService) pipeLogs(container *rancherClient.Container, conn *websocket.Conn) {
	defer conn.Close()

	log_name := strings.TrimPrefix(container.Name, r.context.ProjectName+"_")
	logger := r.context.LoggerFactory.Create(log_name)

	for {
		messageType, bytes, err := conn.ReadMessage()
		if messageType != websocket.TextMessage {
			continue
		}

		if err == io.EOF {
			return
		} else if err != nil {
			logrus.Errorf("Failed to read log: %v", err)
			return
		}

		if len(bytes) <= 3 {
			continue
		}

		if bytes[len(bytes)-1] != '\n' {
			bytes = append(bytes, '\n')
		}
		message := bytes[3:]

		if "01" == string(bytes[:2]) {
			logger.Out(message)
		} else {
			logger.Err(message)
		}
	}
}

func (r *RancherService) DependentServices() []project.ServiceRelationship {
	result := []project.ServiceRelationship{}

	for _, rel := range project.DefaultDependentServices(r.context.Project, r) {
		if rel.Type == project.REL_TYPE_LINK {
			rel.Optional = true
			result = append(result, rel)
		}
	}

	return result
}

func (r *RancherService) Client() *rancherClient.RancherClient {
	return r.context.Client
}

func (r *RancherService) Kill() error {
	return project.ErrUnsupported
}

func (r *RancherService) Info() (project.InfoSet, error) {
	return project.InfoSet{}, nil
}

func (r *RancherService) Pull() error {
	return project.ErrUnsupported
}

func TrimSplit(str, sep string, count int) []string {
	result := []string{}
	for _, i := range strings.SplitN(strings.TrimSpace(str), sep, count) {
		result = append(result, strings.TrimSpace(i))
	}

	return result
}
