package events

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/client"
)

type ProcessConfig struct {
	Name    string `json:"name"`
	OnError string `json:"onError"`
}

func (router *EventRouter) createExternalHandler() error {
	// If it exists, delete it, then create it
	err := removeOldHandler(router.name, router.apiClient)
	if err != nil {
		return err
	}

	externalHandler := &client.ExternalHandler{
		Name:           router.name,
		Uuid:           router.name,
		Priority:       int64(router.priority),
		ProcessConfigs: make([]interface{}, len(router.eventHandlers)),
	}

	idx := 0
	for event := range router.eventHandlers {
		externalHandler.ProcessConfigs[idx] = ProcessConfig{
			Name:    event,
			OnError: strings.ToLower(router.resourceName) + ".error",
		}
		idx++
	}
	err = createNewHandler(externalHandler, router.apiClient)
	if err != nil {
		return err
	}
	return nil
}

var createNewHandler = func(externalHandler *client.ExternalHandler, apiClient *client.RancherClient) error {
	handler, err := apiClient.ExternalHandler.Create(externalHandler)
	if err != nil {
		return err
	}
	return waitForTransition(func() (bool, error) {
		handler, err := apiClient.ExternalHandler.ById(handler.Id)
		if err != nil {
			return false, err
		}
		return handler.Transitioning != "yes", nil
	})
}

var removeOldHandler = func(name string, apiClient *client.RancherClient) error {
	listOpts := client.NewListOpts()
	listOpts.Filters["name"] = name
	listOpts.Filters["state"] = "active"
	handlers, err := apiClient.ExternalHandler.List(listOpts)
	if err != nil {
		return err
	}

	for _, handler := range handlers.Data {
		h := &handler
		log.WithFields(log.Fields{
			"handlerId": h.Id,
		}).Debug("Removing old handler")
		doneTransitioning := func() (bool, error) {
			handler, err := apiClient.ExternalHandler.ById(h.Id)
			if err != nil {
				return false, err
			}
			if handler == nil {
				return false, fmt.Errorf("Failed to lookup external handler %v.", handler.Id)
			}
			return handler.Transitioning != "yes", nil
		}

		if _, ok := h.Actions["deactivate"]; ok {
			h, err = apiClient.ExternalHandler.ActionDeactivate(h)
			if err != nil {
				return err
			}

			err = waitForTransition(doneTransitioning)
			if err != nil {
				return err
			}
		}

		h, err := apiClient.ExternalHandler.ById(h.Id)
		if err != nil {
			return err
		}
		if h != nil {
			if _, ok := h.Actions["remove"]; ok {
				h, err = apiClient.ExternalHandler.ActionRemove(h)
				if err != nil {
					return err
				}
				err = waitForTransition(doneTransitioning)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
