package executor

import (
	"net/url"
	"os"
	"path"

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
		"stack.create":        handlers.WithLock(handlers.CreateStack),
		"stack.upgrade":       handlers.WithLock(handlers.UpgradeStack),
		"stack.finishupgrade": handlers.WithLock(handlers.FinishUpgradeStack),
		"stack.rollback":      handlers.WithLock(handlers.RollbackStack),
		"ping": func(event *events.Event, apiClient *client.RancherClient) error {
			return nil
		},
	}

	url, err := url.Parse(os.Getenv("CATTLE_URL"))
	if err != nil {
		logrus.Fatal(err)
	}

	if path.Base(url.Path) == "v1" {
		dir, _ := path.Split(url.Path)
		url.Path = path.Join(dir, "v2-beta")
	}

	router, err := events.NewEventRouter("rancher-compose-executor", 2000,
		url.String(),
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
