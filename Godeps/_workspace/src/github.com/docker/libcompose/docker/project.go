package docker

import (
	"github.com/Sirupsen/logrus"

	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"
)

func NewProject(context *Context) (*project.Project, error) {
	context.ConfigLookup = &lookup.FileConfigLookup{}
	context.EnvironmentLookup = &lookup.OsEnvLookup{}
	context.ServiceFactory = &ServiceFactory{
		context: context,
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

	return p, err
}
