package rancher

import (
	"github.com/Sirupsen/logrus"

	"github.com/rancherio/rancher-compose/librcompose/project"
)

type RancherServiceFactory struct {
	context *Context
}

func (r *RancherServiceFactory) Create(project *project.Project, name string, serviceConfig *project.ServiceConfig) (project.Service, error) {
	if len(r.context.SidekickInfo.sidekickToPrimaries[name]) > 0 {
		return NewSidekick(name, serviceConfig, r.context), nil
	} else {
		return NewService(name, serviceConfig, r.context), nil
	}
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
	p.EnvironmentLookup = c
	if c.ComposeFile == "-" {
		p.File = "."
	} else {
		p.File = c.ComposeFile
	}
	p.ConfigLookup = project.FileConfigLookup{}

	err := p.Load(c.ComposeBytes)

	c.Project = p
	c.SidekickInfo = NewSidekickInfo(p)

	return p, err
}
