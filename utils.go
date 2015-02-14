package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/cloudnautique/go-rancher/client"
)

func GetRancherClient(url string, apiKey string, apiSecret string) (*client.RancherClient, error) {
	client_opts := client.ClientOpts{
		Url:       url,
		AccessKey: apiKey,
		SecretKey: apiSecret,
	}

	client, err := client.NewRancherClient(&client_opts)
	if err != nil {
		log.Fatalf("Could not get Rancher client: %s", err)
	}

	return client, nil
}
