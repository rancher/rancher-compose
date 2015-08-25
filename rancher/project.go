package rancher

import (
	"github.com/Sirupsen/logrus"

	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"
)

func NewProject(context *Context) (*project.Project, error) {
	context.ConfigLookup = &lookup.FileConfigLookup{}
	context.EnvironmentLookup = &lookup.OsEnvLookup{}
	context.LoggerFactory = logger.NewColorLoggerFactory()
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

	context.SidekickInfo = NewSidekickInfo(p)

	return p, err
}
