package service

import (
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudnautique/go-rancher/client"
)

type Service struct {
	ServiceName    string
	Config         map[interface{}]interface{}
	ProjectName    string
	ContainerCount int
}

func New(projectName string, serviceName string, containerConfig map[interface{}]interface{}) *Service {
	service := &Service{
		ServiceName:    serviceName,
		Config:         containerConfig,
		ProjectName:    projectName,
		ContainerCount: 0,
	}
	return service
}

func (s *Service) Create(rClient *client.RancherClient) error {
	container := s.createContainerConfig()
	log.Infof("Starting container: %s", container.Name)

	_, err := rClient.Container.Create(container)
	if err != nil {
		log.Fatalf("Could not create container: %s", err)
	}

	return nil
}

func (s *Service) Delete(rClient *client.RancherClient) error {
	listOpts := client.NewListOpts()
	listOpts.Filters["name_prefix"] = s.containerNamePrefix()

	containers, err := rClient.Container.List(listOpts)
	if err != nil {
		log.Fatalf("Could not get list of containers")
	}

	for _, container := range containers.Data {
		log.Infof("Stopping Container: %s", container.Name)
		rClient.Container.Delete(&container)
	}
	return nil
}

func (s *Service) containerNamePrefix() string {
	return s.ProjectName + "_" + s.ServiceName + "_"
}

func (s *Service) createContainerConfig() *client.Container {
	network := []string{"1n2"}
	log.Infof("%v", s.Config)
	container_num := s.ContainerCount + 1
	container := &client.Container{
		Name:       s.containerNamePrefix() + strconv.Itoa(container_num),
		ImageUuid:  "docker:" + s.Config["image"].(string),
		NetworkIds: network,
	}

	addPorts(container, s.Config)

	return container
}

func addPorts(container *client.Container, config map[interface{}]interface{}) *client.Container {
	keys := make(map[string]int)
	length := 0

	_, portsExists := config["ports"]
	if portsExists {
		keys["ports"] = len(config["ports"].([]interface{}))
		length = length + keys["ports"]
	}
	_, exposedPortsExist := config["expose"]
	if exposedPortsExist {
		keys["expose"] = len(config["expose"].([]interface{}))
		length = length + keys["expose"]
	}

	// if nothing to add... leave container alone.
	if length == 0 {
		return container
	}

	portList := make([]string, length)
	offset := 0
	for key, off := range keys {
		for i, port := range config[key].([]interface{}) {
			portList[offset+i] = port.(string)
		}
		offset = off
		container.Ports = portList
	}

	return container
}
