package rancher

import "github.com/Sirupsen/logrus"

import "github.com/rancherio/rancher-compose/librcompose/project"

type RancherServiceFactory struct {
	context *Context
}

func (r *RancherServiceFactory) Create(project *project.Project, name string, serviceConfig *project.ServiceConfig) (project.Service, error) {
	return NewService(name, serviceConfig, r.context), nil
}

func NewProject(c *Context) (*project.Project, error) {
	if err := c.open(); err != nil {
		logrus.Errorf("Failed to open project %s: %v", c.ProjectName, err)
		return nil, err
	}

	factory := &RancherServiceFactory{
		context: c,
	}

	p := project.NewProject(c.ProjectName, factory)
	err := p.Load(c.ComposeBytes)

	c.Project = p
	return p, err
}
