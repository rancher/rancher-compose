package events

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/rancher/event-subscriber/locks"
	"github.com/rancher/go-rancher/v2"
)

type EventLocker func(event *Event) locks.Locker

func nopLocker(_ *Event) locks.Locker { return locks.NopLocker() }

func resourceIDLocker(event *Event) locks.Locker {
	if event.ResourceID == "" {
		return locks.NopLocker()
	}
	key := fmt.Sprintf("%s:%s", event.ResourceType, event.ResourceID)
	return locks.KeyLocker(key)
}

type WorkerPool interface {
	HandleWork(event *Event, eventHandlers map[string]EventHandler, apiClient *client.RancherClient)
}

type skippingWorkerPool struct {
	workers     chan int
	eventLocker EventLocker
}

func SkippingWorkerPool(size int, eventLocker EventLocker) WorkerPool {
	if eventLocker == nil {
		eventLocker = nopLocker
	}
	wp := &skippingWorkerPool{workers: make(chan int, size), eventLocker: eventLocker}
	for i := 0; i < size; i++ {
		wp.workers <- i
	}
	return wp
}

func (wp *skippingWorkerPool) HandleWork(event *Event, eventHandlers map[string]EventHandler, apiClient *client.RancherClient) {
	select {
	case w := <-wp.workers:
		go func() {
			defer func() { wp.workers <- w }()
			doWork(event, eventHandlers, apiClient, wp.eventLocker(event))
		}()
	default:
		log.Warnf("No workers available, dropping event. workerCount: %v, event: %v", cap(wp.workers), *event)
	}
}

type nonSkippingWorkerPool struct {
	workers chan int
}

func NonSkippingWorkerPool(size int) WorkerPool {
	wp := &nonSkippingWorkerPool{workers: make(chan int, size)}
	go func() {
		for i := 0; i < size; i++ {
			wp.workers <- i
		}
	}()
	return wp
}

func (wp *nonSkippingWorkerPool) HandleWork(event *Event, eventHandlers map[string]EventHandler, apiClient *client.RancherClient) {
	w := <-wp.workers
	go func() {
		defer func() { wp.workers <- w }()
		doWork(event, eventHandlers, apiClient, nopLocker(event))
	}()
}

func doWork(event *Event, eventHandlers map[string]EventHandler, apiClient *client.RancherClient, locker locks.Locker) {

	if event.Name != "ping" {
		log.WithFields(log.Fields{
			"event": *event,
		}).Debug("Processing event.")
	}

	unlocker := locker.Lock()
	if unlocker == nil {
		log.WithFields(log.Fields{
			"resourceId": event.ResourceID,
		}).Debug("Resource locked. Dropping event")
		return
	}
	defer unlocker.Unlock()

	if fn, ok := eventHandlers[event.Name]; ok {
		if err := fn(event, apiClient); err != nil {
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
