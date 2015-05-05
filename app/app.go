package app

import (
	"log"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancherio/rancher-compose/librcompose/project"
	"github.com/rancherio/rancher-compose/rancher"
)

func newContext(c *cli.Context) *rancher.Context {
	return &rancher.Context{
		ComposeFile: c.GlobalString("file"),
		ProjectName: c.GlobalString("project-name"),
		Url:         c.GlobalString("url"),
		AccessKey:   c.GlobalString("access-key"),
		SecretKey:   c.GlobalString("secret-key"),
	}
}

func ProjectUp(c *cli.Context) {
	err := requireProject(c).Up()
	if err != nil {
		logrus.Fatal(err)
	}
}

func requireProject(c *cli.Context) *project.Project {
	context := newContext(c)
	project, err := rancher.NewProject(context)
	if err != nil {
		log.Fatal(err)
	}

	return project
}
