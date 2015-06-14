package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	cliApp "github.com/rancherio/rancher-compose/app"
	"github.com/rancherio/rancher-compose/librcompose"
)

var VERSION = "0.0.0-dev"

func main() {
	app := cli.NewApp()

	app.Name = "rancher-compose"
	app.Usage = "Docker-compose to Rancher"
	app.Version = librcompose.VERSION
	app.Author = "Rancher"
	app.Email = ""
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "debug",
		},
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
			Name:  "rancher-file,r",
			Usage: "Specify an alternate Rancher compose file (default: rancher-compose.yml)",
		},
		cli.StringFlag{
			Name:  "project-name,p",
			Usage: "Specify an alternate project name (default: directory name)",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "create",
			Usage:  "Create all services but do not start",
			Action: cliApp.ProjectCreate,
		},
		{
			Name:   "up",
			Usage:  "Bring all services up",
			Action: cliApp.ProjectUp,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "d",
					Usage: "Do not block and log",
				},
			},
		},
		{
			Name:   "start",
			Usage:  "Start services",
			Action: cliApp.ProjectUp,
		},
		{
			Name:   "logs",
			Usage:  "Get service logs",
			Action: cliApp.ProjectLog,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "lines",
					Usage: "number of lines to tail",
					Value: 100,
				},
			},
		},
		{
			Name:   "restart",
			Usage:  "Restart services",
			Action: cliApp.ProjectRestart,
		},
		{
			Name:      "stop",
			ShortName: "down",
			Usage:     "Stop services",
			Action:    cliApp.ProjectDown,
		},
		{
			Name:   "scale",
			Usage:  "Scale services",
			Action: cliApp.Scale,
		},
		{
			Name:   "rm",
			Usage:  "Delete services",
			Action: cliApp.ProjectDelete,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force,f",
					Usage: "Allow deletion of all services",
				},
			},
		},
	}

	app.Run(os.Args)
}
