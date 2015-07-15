package rancher

import "github.com/docker/libcompose/project"

type Sidekick struct {
	project.EmptyService

	name          string
	serviceConfig *project.ServiceConfig
	context       *Context
}

func NewSidekick(name string, serviceConfig *project.ServiceConfig, context *Context) *Sidekick {
	return &Sidekick{
		name:          name,
		serviceConfig: serviceConfig,
		context:       context,
	}
}

func (s *Sidekick) Name() string {
	return s.name
}

func (s *Sidekick) primaries() []string {
	return s.context.SidekickInfo.sidekickToPrimaries[s.name]
}

func (s *Sidekick) Config() *project.ServiceConfig {
	links := []string{}

	for _, primary := range s.primaries() {
		links = append(links, primary)
	}

	config := *s.serviceConfig
	config.Links = project.NewMaporColonSlice(links)
	config.VolumesFrom = []string{}

	return &config
}

func (s *Sidekick) DependentServices() []project.ServiceRelationship {
	return project.DefaultDependentServices(s.context.Project, s)
}

func (s *Sidekick) Log() error {
	return nil
}
