package app

import (
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/command"
	"github.com/docker/libcompose/project"
	"github.com/rancherio/rancher-compose/rancher"
)

type ProjectFactory struct {
}

func (p *ProjectFactory) Create(c *cli.Context) (*project.Project, error) {
	context := &rancher.Context{
		RancherComposeFile: c.GlobalString("rancher-file"),
		Url:                c.GlobalString("url"),
		AccessKey:          c.GlobalString("access-key"),
		SecretKey:          c.GlobalString("secret-key"),
	}
	command.Populate(&context.Context, c)

	return rancher.NewProject(context)
}
