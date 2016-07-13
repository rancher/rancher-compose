package executor

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/event-subscriber/events"
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
		"environment.create":        handlers.WithLock(handlers.CreateEnvironment),
		"environment.upgrade":       handlers.WithLock(handlers.UpgradeEnvironment),
		"environment.finishupgrade": handlers.WithLock(handlers.FinishUpgradeEnvironment),
		"environment.rollback":      handlers.WithLock(handlers.RollbackEnvironment),
		"ping": func(event *events.Event, apiClient *client.RancherClient) error {
			return nil
		},
	}

	router, err := events.NewEventRouter("rancher-compose-executor", 2000,
		os.Getenv("CATTLE_URL"),
		os.Getenv("CATTLE_ACCESS_KEY"),
		os.Getenv("CATTLE_SECRET_KEY"),
		nil, eventHandlers, "environment", 10, events.DefaultPingConfig)
	if err != nil {
		logrus.WithField("error", err).Fatal("Unable to create event router")
	}

	if err := router.Start(nil); err != nil {
		logrus.WithField("error", err).Fatal("Unable to start event router")
	}

	logger.Info("Exiting rancher-compose-executor")
}
