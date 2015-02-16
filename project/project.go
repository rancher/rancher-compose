package project

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudnautique/go-rancher/client"
	"github.com/cloudnautique/rancher-composer/parser"
	"github.com/cloudnautique/rancher-composer/service"
)

type Project struct {
	ProjectName string
	Client      *client.RancherClient
	Print       func()
}

var (
	ProjectServices map[string]*service.Service
)

func NewProject(name string, filename string, client *client.RancherClient) (*Project, error) {
	ProjectServices = make(map[string]*service.Service)
	project := &Project{
		ProjectName: name,
		Client:      client,
		Print:       printProjectServices,
	}

	m, err := parser.YamlUnmarshal(filename)
	if err != nil {
		log.Fatalf("Could not parse %s file. %v", filename, err)
	}

	for service, config := range m {
		log.Infof("Project has service: %s", service)
		addService(service.(string), config.(map[interface{}]interface{}))
	}

	return project, nil
}

func printProjectServices() {
	for service := range ProjectServices {
		log.Infof("Service: %s has been parsed", service)
	}
}

func addService(serviceName string, containerConfig map[interface{}]interface{}) error {
	//todo: remove the "_" character
	service := service.New("rc_", serviceName, containerConfig)
	if _, exists := ProjectServices[serviceName]; exists {
		return fmt.Errorf("Service: %s already exists", serviceName)
	}

	ProjectServices[serviceName] = service
	return nil
}

func (p *Project) StartAllServices() error {
	for name, service := range ProjectServices {
		log.Infof("Bringing up service: %s", name)
		err := service.Create(p.Client)
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}
	}
	return nil
}

func (p *Project) RmAllServices() error {
	for name, service := range ProjectServices {
		log.Infof("Removing service: %s", name)
		err := service.Delete(p.Client)
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}
	}
	return nil
}
