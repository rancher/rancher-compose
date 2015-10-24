package executor

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-machine-service/events"
	"github.com/rancher/go-rancher/client"
	"github.com/rancher/rancher-compose/executor/handlers"
	"github.com/rancher/rancher-compose/version"
)

func Main() {
	logger := logrus.WithFields(logrus.Fields{
		"version": version.VERSION,
	})

	logger.Info("Starting rancher-compose-executor")

	eventHandlers := map[string]events.EventHandler{
		"environment.create":        handlers.CreateEnvironment,
		"environment.upgrade":       handlers.UpgradeEnvironment,
		"environment.finishupgrade": handlers.FinishUpgradeEnvironment,
		"environment.rollback":      handlers.RollbackEnvironment,
		"ping": func(event *events.Event, apiClient *client.RancherClient) error {
			return nil
		},
	}

	router, err := events.NewEventRouter("rancher-compose-executor", 2000,
		os.Getenv("CATTLE_URL"),
		os.Getenv("CATTLE_ACCESS_KEY"),
		os.Getenv("CATTLE_SECRET_KEY"),
		nil, eventHandlers, "environment", 10)
	if err != nil {
		logrus.WithField("error", err).Fatal("Unable to create event router")
	}

	if err := router.Start(nil); err != nil {
		logrus.WithField("error", err).Fatal("Unable to start event router")
	}

	logger.Info("Exiting rancher-compose-executor")
}
