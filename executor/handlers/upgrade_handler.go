package handlers

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/utils"
	"github.com/rancher/go-machine-service/events"
	"github.com/rancher/go-rancher/client"
)

func UpgradeEnvironment(event *events.Event, apiClient *client.RancherClient) error {
	logger := logrus.WithFields(logrus.Fields{
		"resourceId": event.ResourceId,
		"eventId":    event.Id,
	})

	logger.Info("Upgrade Stack Event Received")

	if err := upgradeEnvironment(logger, event, apiClient); err != nil {
		logger.Errorf("Stack Upgrade Event Failed: %v", err)
		publishTransitioningReply(err.Error(), event, apiClient)
		return err
	}

	logger.Info("Stack Upgrade Event Done")
	return nil
}

func FinishUpgradeEnvironment(event *events.Event, apiClient *client.RancherClient) error {
	logger := logrus.WithFields(logrus.Fields{
		"resourceId": event.ResourceId,
		"eventId":    event.Id,
	})

	logger.Info("Finish Upgrade Stack Event Received")

	env, err := apiClient.Environment.ById(event.ResourceId)
	if err != nil {
		return err
	}

	if env == nil {
		return errors.New("Failed to find stack")
	}

	var services client.ServiceCollection
	if err := apiClient.GetLink(env.Resource, "services", &services); err != nil {
		return err
	}

	for _, service := range services.Data {
		if err := wait(apiClient, &service); err != nil {
			return err
		}
		if service.State == "upgraded" {
			service, err := apiClient.Service.ActionFinishupgrade(&service)
			if err != nil {
				return err
			}
			if err := wait(apiClient, service); err != nil {
				return err
			}
		}
	}

	logger.Info("Finish Stack Upgrade Event Done")
	return reply(event, apiClient, map[string]interface{}{
		"previousExternalId": nil,
	})
}

func RollbackEnvironment(event *events.Event, apiClient *client.RancherClient) error {
	logger := logrus.WithFields(logrus.Fields{
		"resourceId": event.ResourceId,
		"eventId":    event.Id,
	})

	logger.Info("Rollback Stack Event Received")

	env, err := apiClient.Environment.ById(event.ResourceId)
	if err != nil {
		return err
	}

	if env == nil {
		return errors.New("Failed to find stack")
	}

	var services client.ServiceCollection
	if err := apiClient.GetLink(env.Resource, "services", &services); err != nil {
		return err
	}

	for _, service := range services.Data {
		if err := wait(apiClient, &service); err != nil {
			return err
		}
		if service.State == "upgraded" || service.State == "cancel" {
			service, err := apiClient.Service.ActionRollback(&service)
			if err != nil {
				return err
			}
			if err := wait(apiClient, service); err != nil {
				return err
			}
		}
	}

	logger.Info("Rollback Stack Event Done")
	newId := env.PreviousExternalId
	if newId == "" {
		newId = env.ExternalId
	}

	return reply(event, apiClient, map[string]interface{}{
		"previousExternalId": nil,
		"externalId":         newId,
	})
}

func wait(apiClient *client.RancherClient, service *client.Service) error {
	for i := 0; i < 6; i++ {
		if err := apiClient.Reload(&service.Resource, service); err != nil {
			return err
		}
		if service.Transitioning != "yes" {
			break
		}
	}

	switch service.Transitioning {
	case "yes":
		return fmt.Errorf("Timeout waiting for %s to finish", service.Name)
	case "no":
		return nil
	default:
		return fmt.Errorf("Waiting for %s failed: %s", service.TransitioningMessage)
	}
}

func upgradeEnvironment(logger *logrus.Entry, event *events.Event, apiClient *client.RancherClient) error {
	var upgradeOpts client.EnvironmentUpgrade

	if err := utils.ConvertByJSON(event.Data, &upgradeOpts); err != nil {
		return err
	}

	env, err := apiClient.Environment.ById(event.ResourceId)
	if err != nil {
		return err
	}

	if env == nil {
		return errors.New("Failed to find stack")
	}

	if env.DockerCompose == "" {
		return emptyReply(event, apiClient)
	}

	project, err := constructProjectUpgrade(logger, env, upgradeOpts, apiClient.Opts.Url, apiClient.Opts.AccessKey, apiClient.Opts.SecretKey)
	if err != nil {
		return err
	}

	publishTransitioningReply("Upgrading stack", event, apiClient)

	if err := project.Up(); err != nil {
		return err
	}

	previous := env.PreviousExternalId
	if previous == "" {
		previous = env.ExternalId
	}

	return reply(event, apiClient, map[string]interface{}{
		"externalId":         upgradeOpts.ExternalId,
		"previousExternalId": previous,
	})
}
