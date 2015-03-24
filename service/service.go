package service

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/go-rancher/client"
)

func New(projectName string, serviceName string, containerConfig map[interface{}]interface{}, rClient *client.RancherClient) *Service {
	service := &Service{
		ServiceName:   serviceName,
		Config:        containerConfig,
		ProjectName:   projectName,
		RancherClient: rClient,
	}
	return service
}

func (s *Service) Create() error {
	container := s.createContainerConfig()
	log.Infof("Starting container: %s", container.Name)

	_, err := s.RancherClient.Container.Create(container)
	if err != nil {
		log.Fatalf("Could not create container: %s", err)
	}

	return nil
}

func (s *Service) Delete() error {
	listOpts := client.NewListOpts()
	listOpts.Filters["name_prefix"] = s.containerNamePrefix()

	containers, err := s.RancherClient.Container.List(listOpts)
	if err != nil {
		log.Fatalf("Could not get list of containers")
	}

	for _, container := range containers.Data {
		log.Infof("Stopping Container: %s", container.Name)
		s.RancherClient.Container.Delete(&container)
	}
	return nil
}

func (s *Service) GetLinkDefs() map[string]string {
	links := make(map[string]string)
	_, exists := s.Config["links"]
	if exists {
		for _, link := range s.Config["links"].([]interface{}) {
			spLink := strings.Split(link.(string), ":")
			if len(spLink) == 2 {
				links[spLink[0]] = spLink[1]
			} else {
				links[spLink[0]] = spLink[0]
			}
		}
	}
	return links
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
	addEnvironmentVariables(container, s.Config)
	addCommands(container, s.Config)
	setPrivileged(container, s.Config)

	return container
}

func (s *Service) addLinks(container *client.Container) *client.Container {
	return container
}

func addEnvironmentVariables(container *client.Container, config map[interface{}]interface{}) *client.Container {
	envVars, exists := config["environment"]
	if exists {
		vars := make(map[string]interface{})
		for key, value := range envVars.(map[interface{}]interface{}) {
			vars[key.(string)] = value.(string)
		}
		container.Environment = vars
	}
	return container
}

func addCommands(container *client.Container, config map[interface{}]interface{}) *client.Container {
	cmd, exists := config["command"]
	if exists {
		container.Command = cmd.(string)
	}
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

func setPrivileged(container *client.Container, config map[interface{}]interface{}) *client.Container {
	priv, exists := config["privileged"]
	if exists {
		container.Privileged = priv.(bool)
	}
	return container
}
