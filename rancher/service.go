package rancher

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/runconfig"
	rancherClient "github.com/rancherio/go-rancher/client"
	"github.com/rancherio/rancher-compose/librcompose/docker"
	"github.com/rancherio/rancher-compose/librcompose/project"
)

const (
	LB_IMAGE = "rancher/load-balancer"
)

type IsDone func(*rancherClient.Resource) (bool, error)

type ContainerInspect struct {
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

func (r *RancherService) Up() error {
	service, err := r.findExisting(r.name)

	if err == nil && service == nil {
		service, err = r.createService()
	}

	if err != nil {
		return err
	}

	if service.Actions["activate"] != "" {
		service, err = r.context.Client.Service.ActionActivate(service)
		err = r.wait(service)
	}

	return err
}

func (r *RancherService) findExisting(name string) (*rancherClient.Service, error) {
	logrus.Debugf("Finding service %s", name)
	services, err := r.context.Client.Service.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"environmentId": r.context.Environment.Id,
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

func (r *RancherService) createLbService() (*rancherClient.Service, error) {
	_, err := r.context.Client.LoadBalancerService.Create(&rancherClient.LoadBalancerService{
		Name: r.name,
		LaunchConfig: rancherClient.Container{
			Ports: r.serviceConfig.Ports,
		},
		Scale:         1,
		EnvironmentId: r.context.Environment.Id,
	})

	if err != nil {
		return nil, err
	}

	return r.findExisting(r.name)
}

func (r *RancherService) createNormalService() (*rancherClient.Service, error) {
	launchConfig, err := r.createLaunchConfig()
	if err != nil {
		return nil, err
	}

	launchConfig.Ports = r.serviceConfig.Ports

	return r.context.Client.Service.Create(&rancherClient.Service{
		Name:          r.name,
		LaunchConfig:  launchConfig,
		Scale:         1,
		EnvironmentId: r.context.Environment.Id,
	})
}

func (r *RancherService) createService() (*rancherClient.Service, error) {
	logrus.Infof("Creating service %s", r.name)

	var service *rancherClient.Service
	var err error

	if r.serviceConfig.Image == LB_IMAGE {
		service, err = r.createLbService()
	} else {
		service, err = r.createNormalService()
	}

	if err != nil {
		return nil, err
	}

	links, err := r.getLinks()
	if err == nil && len(links) > 0 {
		_, err = r.context.Client.Service.ActionSetservicelinks(service, &rancherClient.SetServiceLinksInput{
			ServiceIds: links,
		})
	}

	err = r.wait(service)
	return service, err
}

func (r *RancherService) getLinks() ([]string, error) {
	result := []string{}

	for _, link := range r.serviceConfig.Links {
		name := strings.Split(link, ":")[0]
		linkedService, err := r.findExisting(name)
		if err != nil {
			return nil, err
		}

		if linkedService == nil {
			logrus.Warnf("Failed to find service %s to link to", name)
		} else {
			result = append(result, linkedService.Id)
		}
	}

	return result, nil
}

func (r *RancherService) createLaunchConfig() (rancherClient.Container, error) {
	var result rancherClient.Container

	schemasUrl := strings.SplitN(r.context.Client.Schemas.Links["self"], "/schemas", 2)[0]
	scriptsUrl := schemasUrl + "/scripts/transform"

	config, hostConfig, err := docker.Convert(r.serviceConfig)
	if err != nil {
		return result, err
	}

	dockerContainer := &ContainerInspect{
		Config:     config,
		HostConfig: hostConfig,
	}

	err = r.context.Client.Post(scriptsUrl, dockerContainer, &result)
	return result, err
}

func (r *RancherService) wait(service *rancherClient.Service) error {
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
