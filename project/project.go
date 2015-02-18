package project

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudnautique/rancher-composer/parser"
	"github.com/cloudnautique/rancher-composer/service"
	"github.com/rancherio/go-rancher/client"
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
		project.addService(service.(string), config.(map[interface{}]interface{}))
	}

	return project, nil
}

func printProjectServices() {
	s := sortServices()
	for i, service := range s {
		log.Infof("Service: %v %s has been parsed", i, service)
	}
}

func (p *Project) addService(serviceName string, containerConfig map[interface{}]interface{}) error {
	service := service.New(p.ProjectName, serviceName, containerConfig, p.Client)
	if _, exists := ProjectServices[serviceName]; exists {
		return fmt.Errorf("Service: %s already exists", serviceName)
	}

	ProjectServices[serviceName] = service
	return nil
}

func (p *Project) StartAllServices() error {
	serviceStartOrder := sortServices()
	for _, service := range serviceStartOrder {
		log.Infof("Bringing up service: %s", service)
		err := ProjectServices[service].Create()
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}
	}
	return nil
}

func (p *Project) RmAllServices() error {
	for name, service := range ProjectServices {
		log.Infof("Removing service: %s", name)
		err := service.Delete()
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}
	}
	return nil
}

func sortServices() []string {
	// ported from Fig
	unmarkedServices := make([]string, 0, len(ProjectServices))
	sortedServices := make([]string, 0)
	temporaryMarked := make([]string, 0, len(unmarkedServices))

	for service, _ := range ProjectServices {
		unmarkedServices = append(unmarkedServices, service)
	}

	var visit func(service *service.Service)

	visit = func(service *service.Service) {
		if stringInArray(service.ServiceName, temporaryMarked) {
			log.Fatalf("Service %s has either a circular dep or is linked to itself", service.ServiceName)
		}
		if stringInArray(service.ServiceName, unmarkedServices) {
			temporaryMarked = append(temporaryMarked, service.ServiceName)
			links := service.GetLinkDefs()
			for svc, _ := range links {
				visit(ProjectServices[svc])
			}
			temporaryMarked = removeFromArray(service.ServiceName, temporaryMarked)
			unmarkedServices = removeFromArray(service.ServiceName, unmarkedServices)
			sortedServices = append(sortedServices, service.ServiceName)
		}

	}

	for _, service := range unmarkedServices {
		visit(ProjectServices[service])
	}

	return sortedServices
}

func removeFromArray(item string, array []string) []string {
	dumb := make([]string, 0, len(array)-1)
	for _, str := range array {
		if str != item {
			dumb = append(dumb, str)
		}
	}
	return dumb
}

func stringInArray(item string, array []string) bool {
	for _, str := range array {
		if str == item {
			return true
		}
	}
	return false
}
