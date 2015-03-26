package project

import (
	"errors"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ServiceState string

var (
	EXECUTED   ServiceState = ServiceState("executed")
	UNKNOWN    ServiceState = ServiceState("unknown")
	ErrRestart error        = errors.New("Restart execution")
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

	log.Infof("Adding service: %s", name)

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
	wrappers := make(map[string]*ServiceWrapper)

	for name, _ := range p.Services {
		wrappers[name] = NewServiceWrapper(name, p)
	}

	log.Infof("Starting project: %s, services: %d", p.Name, len(wrappers))

	return p.startAll(wrappers)
}

func (p *Project) startAll(wrappers map[string]*ServiceWrapper) error {
	restart := false

	for _, wrapper := range wrappers {
		wrapper.Reset()
	}

	for _, wrapper := range wrappers {
		go wrapper.Start(wrappers)
	}

	var firstError error

	for _, wrapper := range wrappers {
		err := wrapper.Wait()
		if err == ErrRestart {
			restart = true
		} else if err != nil {
			log.Errorf("Failed to start: %s : %v", wrapper.name, err)
			if firstError == nil {
				firstError = err
			}
		}
	}

	if restart {
		if p.ReloadCallback != nil {
			if err := p.ReloadCallback(); err != nil {
				log.Errorf("Failed calling callback: %v", err)
			}
		}
		return p.startAll(wrappers)
	} else {
		return firstError
	}
}

type ServiceWrapper struct {
	name     string
	services map[string]Service
	service  Service
	done     sync.WaitGroup
	state    ServiceState
	err      error
}

func NewServiceWrapper(name string, p *Project) *ServiceWrapper {
	wrapper := &ServiceWrapper{
		name:     name,
		services: make(map[string]Service),
		service:  p.Services[name],
		state:    UNKNOWN,
	}
	return wrapper
}

func (s *ServiceWrapper) Reset() {
	if s.err == ErrRestart {
		s.err = nil
	}
	s.done.Add(1)
}

func (s *ServiceWrapper) Start(wrappers map[string]*ServiceWrapper) {
	defer s.done.Done()

	if s.state == EXECUTED {
		return
	}

	for _, link := range append(s.service.Config().Links, s.service.Config().VolumesFrom...) {
		name := strings.Split(link, ":")[0]
		if wrapper, ok := wrappers[name]; ok {
			if wrapper.Wait() == ErrRestart {
				log.Infof("Restart from dependency: %s of %s", wrapper.name, s.name)
				s.err = ErrRestart
				return
			}
		} else {
			log.Errorf("Failed to find %s", name)
		}
	}

	s.state = EXECUTED

	log.Infof("Starting service %s", s.name)

	s.err = s.service.Up()
	if s.err == ErrRestart {
		log.Infof("Restart from service %s", s.name)
	} else if s.err != nil {
		log.Errorf("Failed to start %s : %v", s.name, s.err)
	} else {
		log.Infof("Started service %s", s.name)
	}
}

func (s *ServiceWrapper) Wait() error {
	s.done.Wait()
	return s.err
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
