package docker

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"
)

type Service struct {
	project.EmptyService

	name          string
	serviceConfig *project.ServiceConfig
	context       *Context
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Config() *project.ServiceConfig {
	return s.serviceConfig
}

func (s *Service) DependentServices() []project.ServiceRelationship {
	return project.DefaultDependentServices(s.context.Project, s)
}

func (s *Service) Create() error {
	_, err := s.doCreate(true)
	return err
}

func (s *Service) collectContainers() ([]*Container, error) {
	containers, err := GetContainerByFilter(s.context.Client, SERVICE.Eq(s.name), PROJECT.Eq(s.context.Project.Name))
	if err != nil {
		return nil, err
	}

	result := []*Container{}

	for _, container := range containers {
		result = append(result, NewContainer(container.Labels[NAME.Str()], s))
	}

	return result, nil
}

func (s *Service) doCreate(create bool) (*Container, error) {
	containers, err := s.collectContainers()
	if err != nil {
		return nil, err
	} else if len(containers) > 0 {
		return containers[0], err
	}

	containerName, err := OneName(s.context.Client, s.context.Project.Name, s.name)
	if err != nil {
		return nil, err
	}

	c := NewContainer(containerName, s)

	if create {
		_, err = c.Create()
	}
	return c, err
}

func (s *Service) Up() error {
	return s.up(true)
}

func (s *Service) Start() error {
	return s.up(false)
}

func (s *Service) up(create bool) error {
	containers, err := s.collectContainers()
	if err != nil {
		return err
	}

	logrus.Debugf("Found %d existing containers for service %s", len(containers), s.name)

	//TODO: Replace here if needed

	if len(containers) == 0 && create {
		c, err := s.doCreate(true)
		if err != nil {
			return err
		}
		containers = []*Container{c}
	}

	return s.eachContainer(func(c *Container) error {
		return c.Up()
	})
}

func (s *Service) eachContainer(action func(*Container) error) error {
	containers, err := s.collectContainers()
	if err != nil {
		return err
	}

	tasks := utils.InParallel{}
	for _, container := range containers {
		task := func(container *Container) func() error {
			return func() error {
				return action(container)
			}
		}(container)

		tasks.Add(task)
	}

	return tasks.Wait()
}

func (s *Service) Down() error {
	return s.eachContainer(func(c *Container) error {
		return c.Down()
	})
}

func (s *Service) Restart() error {
	return s.eachContainer(func(c *Container) error {
		return c.Restart()
	})
}

func (s *Service) Delete() error {
	return s.eachContainer(func(c *Container) error {
		return c.Delete()
	})
}

func (s *Service) Log() error {
	return s.eachContainer(func(c *Container) error {
		return c.Log()
	})
}

func (s *Service) Pull() error {
	c, err := s.doCreate(false)
	if err != nil {
		return err
	}

	return c.Pull()
}

func (s *Service) Containers() ([]project.Container, error) {
	result := []project.Container{}
	containers, err := s.collectContainers()
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		result = append(result, c)
	}

	return result, nil
}
