package handlers

import (
	"github.com/rancher/go-machine-service/events"
	"github.com/rancher/go-rancher/client"
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
		PreviousIds: []string{event.Id},
	}
}

func reply(event *events.Event, apiClient *client.RancherClient, data map[string]interface{}) error {
	reply := newReply(event)
	reply.Data = data
	return publishReply(reply, apiClient)
}
