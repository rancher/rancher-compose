package handlers

import (
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/event-subscriber/events"
	"github.com/rancher/go-rancher/client"
)

func CreateEnvironment(event *events.Event, apiClient *client.RancherClient) error {
	logger := logrus.WithFields(logrus.Fields{
		"resourceId": event.ResourceID,
		"eventId":    event.ID,
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
	env, err := apiClient.Environment.ById(event.ResourceID)
	if err != nil {
		return err
	}

	if env == nil {
		return errors.New("Failed to find stack")
	}

	if env.DockerCompose == "" {
		return emptyReply(event, apiClient)
	}

	_, project, err := constructProject(logger, env, apiClient.Opts.Url, apiClient.Opts.AccessKey, apiClient.Opts.SecretKey)
	if err != nil {
		return err
	}

	publishTransitioningReply("Creating stack", event, apiClient)

	if err := project.Create(); err != nil {
		return err
	}

	startOnCreate := false
	if fields, ok := env.Data["fields"].(map[string]interface{}); ok {
		if on, ok := fields["startOnCreate"].(bool); ok {
			startOnCreate = on
		}
	}

	if startOnCreate {
		if err := project.Create(); err != nil {
			return err
		}

		if err := project.Up(); err != nil {
			return err
		}
	} else {
		// This is to make sure circular links work
		if err := project.Create(); err != nil {
			return err
		}
	}

	return emptyReply(event, apiClient)
}
