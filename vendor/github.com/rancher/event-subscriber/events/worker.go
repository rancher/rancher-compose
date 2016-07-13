package events

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/event-subscriber/locks"
	"github.com/rancher/go-rancher/client"
)

func newWorker() *Worker {
	return &Worker{}
}

type Worker struct {
}

func (w *Worker) DoWork(rawEvent []byte, eventHandlers map[string]EventHandler, apiClient *client.RancherClient,
	workers chan *Worker) {
	defer func() { workers <- w }()

	event := &Event{}
	err := json.Unmarshal(rawEvent, &event)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error unmarshalling event")
		return
	}

	if event.Name != "ping" {
		log.WithFields(log.Fields{
			"event": string(rawEvent[:]),
		}).Debug("Processing event.")
	}

	unlocker := locks.Lock(event.ResourceID)
	if unlocker == nil {
		log.WithFields(log.Fields{
			"resourceId": event.ResourceID,
		}).Debug("Resource locked. Dropping event")
		return
	}
	defer unlocker.Unlock()

	if fn, ok := eventHandlers[event.Name]; ok {
		err = fn(event, apiClient)
		if err != nil {
			log.WithFields(log.Fields{
				"eventName":  event.Name,
				"eventId":    event.ID,
				"resourceId": event.ResourceID,
				"err":        err,
			}).Error("Error processing event")

			reply := &client.Publish{
				Name:                 event.ReplyTo,
				PreviousIds:          []string{event.ID},
				Transitioning:        "error",
				TransitioningMessage: err.Error(),
			}
			_, err := apiClient.Publish.Create(reply)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("Error sending error-reply")
			}
		}
	} else {
		log.WithFields(log.Fields{
			"eventName": event.Name,
		}).Warn("No event handler registered for event")
	}
}
