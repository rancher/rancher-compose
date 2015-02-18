package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudnautique/rancher-composer/project"
	"github.com/codegangsta/cli"
	"github.com/rancherio/go-rancher/client"
)

func main() {
	app := cli.NewApp()

	app.Name = "composer"
	app.Usage = "FIG 2 Rancher"
	app.Version = "0.0.1"
	app.Author = "Bill Maxwell"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "api-url",
			Usage: fmt.Sprintf(
				"Specify the Rancher API Endpoint URL",
			),
		},
		cli.StringFlag{
			Name: "access-key",
			Usage: fmt.Sprintf(
				"Specify api access key",
			),
		},
		cli.StringFlag{
			Name: "secret-key",
			Usage: fmt.Sprintf(
				"Specify api secret key",
			),
			EnvVar: "RANCHER_SECRET_KEY",
		},
		cli.StringFlag{
			Name:  "f",
			Usage: "docker-compose yml file to use",
			Value: "docker-compose.yml",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "up",
			Usage: "Bring all services up",
			Action: func(c *cli.Context) {
				prj := getProject(c)
				err := prj.StartAllServices()
				if err != nil {
					log.Fatalf("Died trying to create Servers")
				}
			},
		},
		{
			Name:  "rm",
			Usage: "Remove all containers and services",
			Action: func(c *cli.Context) {
				prj := getProject(c)
				err := prj.RmAllServices()
				if err != nil {
					log.Fatal("Could not remove all services. %s", err)
				}
			},
		},
		{
			Name:  "test",
			Usage: "testing...",
			Action: func(c *cli.Context) {
				prj := getProject(c)
				prj.Print()
			},
		},
	}

	app.Run(os.Args)
}

func getRancherClient(c *cli.Context) *client.RancherClient {
	url := c.GlobalString("api-url")
	accessKey := c.GlobalString("access-key")
	secretKey := c.GlobalString("secret-key")

	rClient, err := GetRancherClient(url, accessKey, secretKey)
	if err != nil {
		log.Fatalf("Unable to get Rancher client: %s", err)
	}

	return rClient
}

func getProject(c *cli.Context) *project.Project {
	filename := c.GlobalString("f")
	rClient := getRancherClient(c)

	prj, err := project.NewProject("rc", filename, rClient)
	if err != nil {
		log.Fatalf("Could not create project from file. %v", filename, err)
	}

	return prj
}
