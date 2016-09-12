package handlers

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/event-subscriber/events"
	"github.com/rancher/event-subscriber/locks"
	"github.com/rancher/go-rancher/v2"
)

var (
	ErrTimeout = errors.New("Timeout waiting service")
)

func emptyReply(event *events.Event, apiClient *client.RancherClient) error {
	reply := newReply(event)
	return publishReply(reply, apiClient)
}

func publishReply(reply *client.Publish, apiClient *client.RancherClient) error {
	_, err := apiClient.Publish.Create(reply)
	return err
}

func publishTransitioningReply(msg string, event *events.Event, apiClient *client.RancherClient) {
	// Since this is only updating the msg for the state transition, we will ignore errors here
	replyT := newReply(event)
	replyT.Transitioning = "yes"
	replyT.TransitioningMessage = msg
	publishReply(replyT, apiClient)
}

func newReply(event *events.Event) *client.Publish {
	return &client.Publish{
		Name:        event.ReplyTo,
		PreviousIds: []string{event.ID},
	}
}

func reply(event *events.Event, apiClient *client.RancherClient, data map[string]interface{}) error {
	reply := newReply(event)
	reply.Data = data
	return publishReply(reply, apiClient)
}

func WithLock(f func(event *events.Event, apiClient *client.RancherClient) error) func(event *events.Event, apiClient *client.RancherClient) error {
	return func(event *events.Event, apiClient *client.RancherClient) error {
		lockKey := fmt.Sprintf("%s:%s", event.ResourceType, event.ResourceID)
		lock := locks.Lock(lockKey)
		if lock == nil {
			logrus.Infof("Busying processing %s", lockKey)
			return nil
		}
		defer lock.Unlock()

		err := f(event, apiClient)
		if err == ErrTimeout {
			logrus.Infof("Timeout processing %s", lockKey)
			return nil
		}
		return nil
	}
}
