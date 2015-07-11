package docker

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"
)

type Service struct {
	project.EmptyService

	tty           bool
	name          string
	serviceConfig *project.ServiceConfig
	context       *Context
}

func (s *Service) Create() error {
	_, err := s.doCreate()
	return err
}

func (s *Service) collectContainers() ([]*Container, error) {
	containers, err := GetContainerByFilter(s.context.Client, SERVICE.Eq(s.name), PROJECT.Eq(s.context.Project.Name))
	if err != nil {
		return nil, err
	}

	result := []*Container{}

	for _, container := range containers {
		result = append(result, NewContainer(container.Names[0][1:], s))
	}

	return result, nil
}

func (s *Service) doCreate() (*Container, error) {
	containerName, err := OneName(s.context.Client, s.context.Project.Name, s.name)
	if err != nil {
		return nil, err
	}

	c := NewContainer(containerName, s)

	_, err = c.Create()
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
		c, err := s.doCreate()
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

func (s *Service) Log() error {
	return s.eachContainer(func(c *Container) error {
		return c.Log()
	})
}
