package docker

import (
	"os"

	"github.com/docker/libcompose/project"
	"golang.org/x/crypto/ssh/terminal"
)

type ServiceFactory struct {
	context *Context
}

func (s *ServiceFactory) Create(project *project.Project, name string, serviceConfig *project.ServiceConfig) (project.Service, error) {
	return &Service{
		tty:           terminal.IsTerminal(int(os.Stdout.Fd())),
		name:          name,
		serviceConfig: serviceConfig,
		context:       s.context,
	}, nil
}
