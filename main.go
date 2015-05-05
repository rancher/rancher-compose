package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	cliApp "github.com/rancherio/rancher-compose/app"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	app := cli.NewApp()

	app.Name = "rancher-compose"
	app.Usage = "Docker-compose to Rancher"
	app.Version = "0.1.0"
	app.Author = "Rancher"
	app.Email = ""
	app.Flags = []cli.Flag{
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
			Name:  "file,f",
			Usage: "Specify an alternate compose file (default: docker-compose.yml)",
			Value: "docker-compose.yml",
		},
		cli.StringFlag{
			Name:  "project-name,p",
			Usage: "Specify an alternate project name (default: directory name)",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "up",
			Usage:  "Bring all services up",
			Action: cliApp.ProjectUp,
		},
	}

	app.Run(os.Args)
}
