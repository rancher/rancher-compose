package handlers

import (
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-machine-service/events"
	"github.com/rancher/go-rancher/client"
)

func CreateEnvironment(event *events.Event, apiClient *client.RancherClient) error {
	logger := logrus.WithFields(logrus.Fields{
		"resourceId": event.ResourceId,
		"eventId":    event.Id,
	})

	logger.Info("Stack Create Event Received")

	if err := createEnvironment(logger, event, apiClient); err != nil {
		logger.Errorf("Stack Create Event Failed: %v", err)
		publishTransitioningReply(err.Error(), event, apiClient)
		return err
	}

	logger.Info("Stack Create Event Done")
	return nil
}

func createEnvironment(logger *logrus.Entry, event *events.Event, apiClient *client.RancherClient) error {
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

	context, project, err := constructProject(logger, env, apiClient.Opts.Url, apiClient.Opts.AccessKey, apiClient.Opts.SecretKey)
	if err != nil {
		return err
	}

	publishTransitioningReply("Creating stack", event, apiClient)

	if err := project.Create(); err != nil {
		return err
	}

	// This is to make sure circular links work
	if err := project.Create(); err != nil {
		return err
	}

	uuid := context.Uuid()
	if uuid == "" {
		return emptyReply(event, apiClient)
	}

	return reply(event, apiClient, map[string]interface{}{
		"externalId": uuid,
	})
}
