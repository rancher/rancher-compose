package rancher

import (
	"github.com/Sirupsen/logrus"

	"github.com/docker/libcompose/project"
)

func NewProject(context *Context) (*project.Project, error) {
	context.ServiceFactory = &RancherServiceFactory{
		Context: context,
	}

	p := project.NewProject(&context.Context)

	err := p.Parse()
	if err != nil {
		return nil, err
	}

	if err = context.open(); err != nil {
		logrus.Errorf("Failed to open project %s: %v", p.Name, err)
		return nil, err
	}

	p.Name = context.ProjectName

	context.SidekickInfo = NewSidekickInfo(p)

	return p, err
}
