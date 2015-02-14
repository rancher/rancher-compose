package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudnautique/go-rancher/client"
	"github.com/cloudnautique/rancher-composer/service"
	"github.com/codegangsta/cli"
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
			Usage: "Bring a yml config or service up",
			Action: func(c *cli.Context) {
				url := c.GlobalString("api-url")
				accessKey := c.GlobalString("access-key")
				secretKey := c.GlobalString("secret-key")
				fileName := c.GlobalString("f")

				log.Infof("Attempting to use file: %s", fileName)

				m, err := service.YamlUnmarshall(fileName)
				if err != nil {
					log.Fatalf("Error parsing Yaml: %s", err)
				}
				log.Infof("Setting up Rancher Client")
				rClient, err := GetRancherClient(url, accessKey, secretKey)
				if err != nil {
					log.Fatalf("Unable to get Client: %s", err)
				}

				for service, config := range m {
					log.Infof("Service: %s found", service)
					container := client.Container{
						Name:      service.(string),
						ImageUuid: "docker:" + config.(map[interface{}]interface{})["image"].(string),
					}
					log.Infof("Image: %s", container.ImageUuid)
					log.Infof("Name: %s", container.Name)
					_, err := rClient.Container.Create(&container)
					if err != nil {
						log.Fatal("Could not create container: ", err)
					}
				}

			},
		},
	}

	app.Run(os.Args)
}
