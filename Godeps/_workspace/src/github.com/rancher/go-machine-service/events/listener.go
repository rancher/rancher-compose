package events

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"

	"github.com/rancher/go-machine-service/locks"
	"github.com/rancher/go-rancher/client"
)

const MaxWait = time.Duration(time.Second * 10)

// Defines the function "interface" that handlers must conform to.
type EventHandler func(*Event, *client.RancherClient) error

type EventRouter struct {
	name          string
	priority      int
	apiUrl        string
	accessKey     string
	secretKey     string
	apiClient     *client.RancherClient
	subscribeUrl  string
	eventHandlers map[string]EventHandler
	workerCount   int
	eventStream   *websocket.Conn
	mu            *sync.Mutex
	resourceName  string
}

type ProcessConfig struct {
	Name    string `json:"name"`
	OnError string `json:"onError"`
}

func (router *EventRouter) Start(ready chan<- bool) (err error) {
	workers := make(chan *Worker, router.workerCount)
	for i := 0; i < router.workerCount; i++ {
		w := newWorker()
		workers <- w
	}

	log.WithFields(log.Fields{
		"workerCount": router.workerCount,
	}).Info("Initializing event router")

	// If it exists, delete it, then create it
	err = removeOldHandler(router.name, router.apiClient)
	if err != nil {
		return err
	}

	externalHandler := &client.ExternalHandler{
		Name:           router.name,
		Uuid:           router.name,
		Priority:       int64(router.priority),
		ProcessConfigs: make([]interface{}, len(router.eventHandlers)),
	}

	handlers := map[string]EventHandler{}

	if pingHandler, ok := router.eventHandlers["ping"]; ok {
		// Ping doesnt need registered in the POST and ping events don't have the handler suffix.
		//If we start handling other non-suffix events, we might consider improving this.
		handlers["ping"] = pingHandler
	}

	idx := 0
	subscribeParams := url.Values{}
	eventHandlerSuffix := ";handler=" + router.name
	for event, handler := range router.eventHandlers {
		processConfig := ProcessConfig{
			Name:    event,
			OnError: router.resourceName + ".error",
		}
		externalHandler.ProcessConfigs[idx] = processConfig
		fullEventKey := event + eventHandlerSuffix
		subscribeParams.Add("eventNames", fullEventKey)
		handlers[fullEventKey] = handler
		idx++
	}
	err = createNewHandler(externalHandler, router.apiClient)
	if err != nil {
		return err
	}

	eventStream, err := subscribeToEvents(router.subscribeUrl, router.accessKey, router.secretKey, subscribeParams)
	if err != nil {
		return err
	}
	log.Info("Connection established")
	router.eventStream = eventStream
	defer router.Stop()

	if ready != nil {
		ready <- true
	}

	for {
		_, message, err := eventStream.ReadMessage()
		if err != nil {
			// Error here means the connection is closed. It's normal, so just return.
			return nil
		}

		message = bytes.TrimSpace(message)
		if len(message) == 0 {
			continue
		}

		select {
		case worker := <-workers:
			go worker.DoWork(message, handlers, router.apiClient, workers)
		default:
			log.WithFields(log.Fields{
				"workerCount": router.workerCount,
			}).Info("No workers available dropping event.")
		}
	}

	return nil
}

func (router *EventRouter) Stop() (err error) {
	if router.eventStream != nil {
		router.mu.Lock()
		defer router.mu.Unlock()
		if router.eventStream != nil {
			router.eventStream.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
			router.eventStream = nil
		}
	}
	return nil
}

// TODO Privatize worker
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

	unlocker := locks.Lock(event.ResourceId)
	if unlocker == nil {
		log.WithFields(log.Fields{
			"resourceId": event.ResourceId,
		}).Debug("Resource locked. Dropping event")
		return
	}
	defer unlocker.Unlock()

	if fn, ok := eventHandlers[event.Name]; ok {
		err = fn(event, apiClient)
		if err != nil {
			log.WithFields(log.Fields{
				"eventName":  event.Name,
				"eventId":    event.Id,
				"resourceId": event.ResourceId,
				"err":        err,
			}).Error("Error processing event")

			reply := &client.Publish{
				Name:                 event.ReplyTo,
				PreviousIds:          []string{event.Id},
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

func NewEventRouter(name string, priority int, apiUrl string, accessKey string, secretKey string,
	apiClient *client.RancherClient, eventHandlers map[string]EventHandler, resourceName string, workerCount int) (*EventRouter, error) {

	if apiClient == nil {
		var err error
		apiClient, err = client.NewRancherClient(&client.ClientOpts{

			Url:       apiUrl,
			AccessKey: accessKey,
			SecretKey: secretKey,
		})
		if err != nil {
			return nil, err
		}
	}

	// TODO Get subscribe collection URL from API instead of hard coding
	subscribeUrl := strings.Replace(apiUrl+"/subscribe", "http", "ws", -1)

	return &EventRouter{
		name:          name,
		priority:      priority,
		apiUrl:        apiUrl,
		accessKey:     accessKey,
		secretKey:     secretKey,
		apiClient:     apiClient,
		subscribeUrl:  subscribeUrl,
		eventHandlers: eventHandlers,
		workerCount:   workerCount,
		mu:            &sync.Mutex{},
		resourceName:  resourceName,
	}, nil
}

func newWorker() *Worker {
	return &Worker{}
}

func subscribeToEvents(subscribeUrl string, accessKey string, secretKey string, data url.Values) (*websocket.Conn, error) {
	dialer := &websocket.Dialer{}
	headers := http.Header{}
	headers.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(accessKey+":"+secretKey)))
	subscribeUrl = subscribeUrl + "?" + data.Encode()
	ws, _, err := dialer.Dial(subscribeUrl, headers)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"subscribeUrl": subscribeUrl,
		}).Error("Failed to subscribe to events.")
		return nil, err
	}

	return ws, nil
}

var createNewHandler = func(externalHandler *client.ExternalHandler, apiClient *client.RancherClient) error {
	_, err := apiClient.ExternalHandler.Create(externalHandler)
	return err
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
			h, err := apiClient.ExternalHandler.ById(h.Id)
			if err != nil {
				return false, err
			}
			return h.Transitioning != "yes", nil
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
	return nil
}

type doneTranitioningFunc func() (bool, error)

func waitForTransition(waitFunc doneTranitioningFunc) error {
	timeoutAt := time.Now().Add(MaxWait)
	ticker := time.NewTicker(time.Millisecond * 250)
	defer ticker.Stop()
	for tick := range ticker.C {
		done, err := waitFunc()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if tick.After(timeoutAt) {
			return fmt.Errorf("Timed out waiting for transtion.")
		}
	}
	return fmt.Errorf("Timed out waiting for transtion.")
}
