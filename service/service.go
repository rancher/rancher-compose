package service

import (
	log "github.com/Sirupsen/logrus"
	"github.com/cloudnautique/go-rancher/client"
)

type Service struct {
	ServiceName string
	Config      map[interface{}]interface{}
}

func New(serviceName string, containerConfig map[interface{}]interface{}) *Service {
	service := &Service{
		ServiceName: serviceName,
		Config:      containerConfig,
	}
	return service
}

func (s *Service) Create(rClient *client.RancherClient) error {
	network := []string{"1n2"}
	container := client.Container{
		Name:       "rc_" + s.ServiceName + "_1",
		ImageUuid:  "docker:" + s.Config["image"].(string),
		NetworkIds: network,
	}
	log.Infof("Starting container: %s", container.Name)

	_, err := rClient.Container.Create(&container)
	if err != nil {
		log.Fatalf("Could not create container: %s", err)
	}

	return nil
}
