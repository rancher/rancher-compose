package handlers

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project/options"
	"github.com/docker/libcompose/utils"
	"github.com/rancher/event-subscriber/events"
	"github.com/rancher/go-rancher/v2"
)

func UpgradeStack(event *events.Event, apiClient *client.RancherClient) error {
	logger := logrus.WithFields(logrus.Fields{
		"resourceId": event.ResourceID,
		"eventId":    event.ID,
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

func FinishUpgradeStack(event *events.Event, apiClient *client.RancherClient) error {
	logger := logrus.WithFields(logrus.Fields{
		"resourceId": event.ResourceID,
		"eventId":    event.ID,
	})

	logger.Info("Finish Upgrade Stack Event Received")

	stack, err := apiClient.Stack.ById(event.ResourceID)
	if err != nil {
		return err
	}

	if stack == nil {
		return errors.New("Failed to find stack")
	}

	var services client.ServiceCollection
	if err := apiClient.GetLink(stack.Resource, "services", &services); err != nil {
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
		"previousExternalId":  nil,
		"previousEnvironment": nil,
	})
}

func RollbackStack(event *events.Event, apiClient *client.RancherClient) error {
	logger := logrus.WithFields(logrus.Fields{
		"resourceId": event.ResourceID,
		"eventId":    event.ID,
	})

	logger.Info("Rollback Stack Event Received")

	stack, err := apiClient.Stack.ById(event.ResourceID)
	if err != nil {
		return err
	}

	if stack == nil {
		return errors.New("Failed to find stack")
	}

	var services client.ServiceCollection
	if err := apiClient.GetLink(stack.Resource, "services", &services); err != nil {
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
	newId := stack.PreviousExternalId
	if newId == "" {
		newId = stack.ExternalId
	}
	newEnv := stack.PreviousEnvironment
	if len(newEnv) == 0 {
		newEnv = stack.Environment
	}

	return reply(event, apiClient, map[string]interface{}{
		"previousExternalId":  nil,
		"previousEnvironment": nil,
		"externalId":          newId,
		"environment":         newEnv,
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
		time.Sleep(500 * time.Millisecond)
	}

	switch service.Transitioning {
	case "yes":
		logrus.Infof("Timeout waiting for %s to finish", service.Name)
		return ErrTimeout
	case "no":
		return nil
	default:
		return fmt.Errorf("Waiting for %s failed: %s", service.Name, service.TransitioningMessage)
	}
}

func upgradeEnvironment(logger *logrus.Entry, event *events.Event, apiClient *client.RancherClient) error {
	var upgradeOpts client.StackUpgrade

	if err := utils.ConvertByJSON(event.Data, &upgradeOpts); err != nil {
		return err
	}

	stack, err := apiClient.Stack.ById(event.ResourceID)
	if err != nil {
		return err
	}

	if stack == nil {
		return errors.New("Failed to find stack")
	}

	if stack.DockerCompose == "" {
		return emptyReply(event, apiClient)
	}

	project, newEnv, err := constructProjectUpgrade(logger, stack, upgradeOpts, apiClient.Opts.Url, apiClient.Opts.AccessKey, apiClient.Opts.SecretKey)
	if err != nil {
		return err
	}

	publishTransitioningReply("Upgrading stack", event, apiClient)

	if err := project.Up(context.Background(), options.Up{}); err != nil {
		return err
	}

	previous := stack.PreviousExternalId
	if previous == "" {
		previous = stack.ExternalId
	}

	previousEnv := stack.PreviousEnvironment
	if len(previousEnv) == 0 {
		previousEnv = stack.Environment
	}

	return reply(event, apiClient, map[string]interface{}{
		"externalId":          upgradeOpts.ExternalId,
		"environment":         newEnv,
		"previousExternalId":  previous,
		"previousEnvironment": previousEnv,
	})
}
