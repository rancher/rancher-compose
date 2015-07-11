package rancher

import "github.com/docker/libcompose/project"

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
