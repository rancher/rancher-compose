package executor

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/event-subscriber/events"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose/executor/handlers"
	"github.com/rancher/rancher-compose/version"
)

func Main() {
	logger := logrus.WithFields(logrus.Fields{
		"version": version.VERSION,
	})

	logger.Info("Starting rancher-compose-executor")

	eventHandlers := map[string]events.EventHandler{
		"stack.create":        handlers.WithTimeout(handlers.CreateStack),
		"stack.upgrade":       handlers.WithTimeout(handlers.UpgradeStack),
		"stack.finishupgrade": handlers.WithTimeout(handlers.FinishUpgradeStack),
		"stack.rollback":      handlers.WithTimeout(handlers.RollbackStack),
		"ping": func(event *events.Event, apiClient *client.RancherClient) error {
			return nil
		},
	}

	router, err := events.NewEventRouter("rancher-compose-executor", 2000,
		os.Getenv("CATTLE_URL"),
		os.Getenv("CATTLE_ACCESS_KEY"),
		os.Getenv("CATTLE_SECRET_KEY"),
		nil, eventHandlers, "stack", 10, events.DefaultPingConfig)
	if err != nil {
		logrus.WithField("error", err).Fatal("Unable to create event router")
	}

	if err := router.Start(nil); err != nil {
		logrus.WithField("error", err).Fatal("Unable to start event router")
	}

	logger.Info("Exiting rancher-compose-executor")
}
