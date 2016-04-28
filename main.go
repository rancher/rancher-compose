package main

import (
	"fmt"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/cli/command"
	rancherApp "github.com/rancher/rancher-compose/app"
	"github.com/rancher/rancher-compose/executor"
	"github.com/rancher/rancher-compose/version"
)

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

func main() {
	if path.Base(os.Args[0]) == "rancher-compose-executor" {
		executor.Main()
	} else {
		cliMain()
	}
}

func cliMain() {
	factory := &rancherApp.ProjectFactory{}

	app := cli.NewApp()
	app.Name = "rancher-compose"
	app.Usage = "Docker-compose to Rancher"
	app.Version = version.VERSION
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Before = beforeApp
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
		cli.StringFlag{
			Name:  "env-file,e",
			Usage: "Specify a file from which to read environment variables",
		},
	)
	app.Commands = []cli.Command{
		rancherApp.CreateCommand(factory),
		rancherApp.UpCommand(factory),
		command.StartCommand(factory),
		command.LogsCommand(factory),
		rancherApp.RestartCommand(factory),
		command.StopCommand(factory),
		command.ScaleCommand(factory),
		command.RmCommand(factory),
		rancherApp.PullCommand(factory),
		rancherApp.UpgradeCommand(factory),
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
