package project

import (
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func NewProject(name string, factory ServiceFactory) *Project {
	return &Project{
		Name:     name,
		configs:  make(map[string]*ServiceConfig),
		Services: make(map[string]Service),
		factory:  factory,
	}
}

func (p *Project) AddConfig(name string, config *ServiceConfig) error {
	service, err := p.factory.Create(p, name, config)
	if err != nil {
		log.Errorf("Failed to create service for %s : %v", name, err)
		return err
	}

	p.configs[name] = config
	p.Services[name] = service

	return nil
}

func (p *Project) Load(bytes []byte) error {
	configs := make(map[string]*ServiceConfig)
	err := yaml.Unmarshal(bytes, configs)
	if err != nil {
		log.Fatalf("Could not parse config for project %s : %v", p.Name, err)
	}

	for name, config := range configs {
		err := p.AddConfig(name, config)
		if err != nil {
			return err
		}
	}

	return nil
}

//func printProjectServices() {
//	s := sortServices()
//	for i, service := range s {
//		log.Infof("Service: %v %s has been parsed", i, service)
//	}
//}
//
//func (p *Project) addService(serviceName string, containerConfig map[interface{}]interface{}) error {
//	service := service.New(p.ProjectName, serviceName, containerConfig, p.Client)
//	if _, exists := ProjectServices[serviceName]; exists {
//		return fmt.Errorf("Service: %s already exists", serviceName)
//	}
//
//	ProjectServices[serviceName] = service
//	return nil
//}

func (p *Project) Up() error {
	log.Infof("Starting project %s", p.Name)
	wrappers := make(map[string]*ServiceWrapper)

	for name, _ := range p.Services {
		wrappers[name] = NewServiceWrapper(name, p)
	}

	for _, wrapper := range wrappers {
		go wrapper.Start(wrappers)
	}

	for _, wrapper := range wrappers {
		wrapper.Wait()
	}

	return nil
}

type ServiceWrapper struct {
	name     string
	services map[string]Service
	service  Service
	done     sync.WaitGroup
}

func NewServiceWrapper(name string, p *Project) *ServiceWrapper {
	wrapper := &ServiceWrapper{
		name:     name,
		services: make(map[string]Service),
		service:  p.Services[name],
	}
	wrapper.done.Add(1)
	return wrapper
}

func (s *ServiceWrapper) Start(wrappers map[string]*ServiceWrapper) {
	defer s.done.Done()

	for _, link := range append(s.service.Config().Links, s.service.Config().VolumesFrom...) {
		name := strings.Split(link, ":")[0]
		if wrapper, ok := wrappers[name]; ok {
			wrapper.Wait()
		} else {
			log.Errorf("Failed to find %s", name)
		}
	}

	log.Infof("Starting service %s", s.name)
	err := s.service.Up()
	if err != nil {
		log.Errorf("Failed to start %s : %v", s.name, err)
	}
	log.Infof("Started service %s", s.name)
}

func (s *ServiceWrapper) Wait() {
	s.done.Wait()
}

//
//func (p *Project) RmAllServices() error {
//	for name, service := range ProjectServices {
//		log.Infof("Removing service: %s", name)
//		err := service.Delete()
//		if err != nil {
//			return fmt.Errorf("Error: %v", err)
//		}
//	}
//	return nil
//}

//func removeFromArray(item string, array []string) []string {
//	dumb := make([]string, 0, len(array)-1)
//	for _, str := range array {
//		if str != item {
//			dumb = append(dumb, str)
//		}
//	}
//	return dumb
//}
//
//func stringInArray(item string, array []string) bool {
//	for _, str := range array {
//		if str == item {
//			return true
//		}
//	}
//	return false
//}
