package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	cliApp "github.com/docker/libcompose/cli/app"
	"github.com/docker/libcompose/cli/command"
	rancherApp "github.com/rancher/rancher-compose/app"
	"github.com/rancher/rancher-compose/version"
)

func main() {
	factory := &rancherApp.ProjectFactory{}

	app := cli.NewApp()
	app.Name = "rancher-compose"
	app.Usage = "Docker-compose to Rancher"
	app.Version = version.VERSION
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Before = cliApp.BeforeApp
	app.Flags = append(command.CommonFlags(),
		cli.StringFlag{
			Name: "url",
			Usage: fmt.Sprintf(
				"Specify the Rancher API endpoint URL",
			),
			EnvVar: "RANCHER_URL",
		},
		cli.StringFlag{
			Name: "access-key",
			Usage: fmt.Sprintf(
				"Specify Rancher API access key",
			),
			EnvVar: "RANCHER_ACCESS_KEY",
		},
		cli.StringFlag{
			Name: "secret-key",
			Usage: fmt.Sprintf(
				"Specify Rancher API secret key",
			),
			EnvVar: "RANCHER_SECRET_KEY",
		},
		cli.StringFlag{
			Name:  "rancher-file,r",
			Usage: "Specify an alternate Rancher compose file (default: rancher-compose.yml)",
		},
	)
	app.Commands = []cli.Command{
		command.CreateCommand(factory),
		command.UpCommand(factory),
		command.StartCommand(factory),
		command.LogsCommand(factory),
		command.RestartCommand(factory),
		command.StopCommand(factory),
		command.ScaleCommand(factory),
		command.RmCommand(factory),
		rancherApp.UpgradeCommand(factory),
	}

	app.Run(os.Args)
}
