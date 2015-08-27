package app

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/cli/app"
	"github.com/docker/libcompose/cli/command"
	"github.com/docker/libcompose/project"
	"github.com/rancher/rancher-compose/rancher"
	"github.com/rancher/rancher-compose/upgrade"
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

func UpgradeCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "upgrade",
		Usage:  "Perform rolling upgrade between services",
		Action: app.WithProject(factory, Upgrade),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "batch-size",
				Usage: "Number of containers to upgrade at once",
				Value: 2,
			},
			cli.IntFlag{
				Name:  "scale",
				Usage: "Final number of running containers",
				Value: -1,
			},
			cli.IntFlag{
				Name:  "interval",
				Usage: "Update interval in milliseconds",
				Value: 2000,
			},
			cli.BoolTFlag{
				Name:  "update-links",
				Usage: "Update inbound links on target service",
			},
			cli.BoolFlag{
				Name:  "wait,w",
				Usage: "Wait for upgrade to complete",
			},
			cli.BoolFlag{
				Name:  "cleanup, c",
				Usage: "Remove the original service definition once upgraded, implies --wait",
			},
		},
	}
}

func Upgrade(p *project.Project, c *cli.Context) {
	args := c.Args()
	if len(args) != 2 {
		logrus.Fatalf("Please pass arguments in the form: [from service] [to service]")
	}

	err := upgrade.Upgrade(p, args[0], args[1], upgrade.UpgradeOpts{
		BatchSize:      c.Int("batch-size"),
		IntervalMillis: c.Int("interval"),
		FinalScale:     c.Int("scale"),
		UpdateLinks:    c.Bool("update-links"),
		Wait:           c.Bool("wait"),
		CleanUp:        c.Bool("cleanup"),
	})

	if err != nil {
		logrus.Fatal(err)
	}
}
