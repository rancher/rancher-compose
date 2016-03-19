package events

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	tu "github.com/rancher/go-machine-service/test_utils"
	"github.com/rancher/go-rancher/client"
)

const eventServerPort string = "8005"
const baseUrl string = "http://localhost:" + eventServerPort
const pushUrl string = baseUrl + "/pushEvent"
const subscribeUrl string = baseUrl + "/subscribe"

func newRouter(eventHandlers map[string]EventHandler, workerCount int, t *testing.T) *EventRouter {
	// Mock out these functions
	createNewHandler = func(externalHandler *client.ExternalHandler, apiClient *client.RancherClient) error {
		return nil
	}
	removeOldHandler = func(name string, apiClient *client.RancherClient) error {
		return nil
	}
	fakeApiClient := &client.RancherClient{}
	router, err := NewEventRouter("testRouter", 2000, baseUrl, "accKey", "secret", fakeApiClient,
		eventHandlers, "physicalhost", workerCount)
	if err != nil {
		t.Fatal(err)
	}
	return router
}

// Tests the simplest case of successfully receiving, routing, and handling
// three events.
func TestSimpleRouting(t *testing.T) {
	eventsReceived := make(chan *Event)
	testHandler := func(event *Event, apiClient *client.RancherClient) error {
		eventsReceived <- event
		return nil
	}

	eventHandlers := map[string]EventHandler{"physicalhost.create": testHandler}
	router := newRouter(eventHandlers, 3, t)
	ready := make(chan bool, 1)
	go router.Start(ready)
	defer router.Stop()
	defer tu.ResetTestServer()
	// Wait for start to be ready
	<-ready

	preCount := 0
	pre := func(event *Event) {
		event.Id = strconv.Itoa(preCount)
		event.ResourceId = strconv.FormatInt(time.Now().UnixNano(), 10)
		preCount += 1
		event.Name = "physicalhost.create;handler=testRouter"
	}

	// Push 3 events
	for i := 0; i < 3; i++ {
		err := prepAndPostEvent("../test_utils/resources/create_virtualbox.json", pre)
		if err != nil {
			t.Fatal(err)
		}
	}
	receivedEvents := map[string]*Event{}
	for i := 0; i < 3; i++ {
		receivedEvent := awaitEvent(eventsReceived, 100, t)
		if receivedEvent != nil {
			receivedEvents[receivedEvent.Id] = receivedEvent
		}
	}

	for i := 0; i < 3; i++ {
		if _, ok := receivedEvents[strconv.Itoa(i)]; !ok {
			t.Errorf("Didn't get event %v", i)
		}
	}
}

// If no workers are available (because they're all busy), an event should simply be dropped.
// This tests that functionality
func TestEventDropping(t *testing.T) {
	eventsReceived := make(chan *Event)
	stopWaiting := make(chan bool)
	testHandler := func(event *Event, apiClient *client.RancherClient) error {
		eventsReceived <- event
		<-stopWaiting
		return nil
	}

	eventHandlers := map[string]EventHandler{"physicalhost.create": testHandler}

	// 2 workers, not 3, means the last event should be droppped
	router := newRouter(eventHandlers, 2, t)
	ready := make(chan bool, 1)
	go router.Start(ready)
	defer router.Stop()
	defer tu.ResetTestServer()
	// Wait for start to be ready
	<-ready

	preCount := 0
	pre := func(event *Event) {
		event.Id = strconv.Itoa(preCount)
		event.ResourceId = strconv.FormatInt(time.Now().UnixNano(), 10)
		preCount += 1
		event.Name = "physicalhost.create;handler=testRouter"
	}

	// Push 3 events
	for i := 0; i < 3; i++ {
		err := prepAndPostEvent("../test_utils/resources/create_virtualbox.json", pre)
		if err != nil {
			t.Fatal(err)
		}
	}
	receivedEvents := map[string]*Event{}
	for i := 0; i < 3; i++ {
		receivedEvent := awaitEvent(eventsReceived, 20, t)
		if receivedEvent != nil {
			receivedEvents[receivedEvent.Id] = receivedEvent
		}
	}

	if len(receivedEvents) != 2 {
		t.Errorf("Unexpected length %v", len(receivedEvents))
	}
}

// Tests that when we have more events than workers, workers are added back to the pool
// when they are done doing their work and capable of handling more work.
func TestWorkerReuse(t *testing.T) {
	eventsReceived := make(chan *Event)
	testHandler := func(event *Event, apiClient *client.RancherClient) error {
		time.Sleep(10 * time.Millisecond)
		eventsReceived <- event
		return nil
	}

	eventHandlers := map[string]EventHandler{"physicalhost.create": testHandler}

	router := newRouter(eventHandlers, 1, t)
	ready := make(chan bool, 1)
	go router.Start(ready)
	defer router.Stop()
	defer tu.ResetTestServer()
	// Wait for start to be ready
	<-ready
	preCount := 1
	pre := func(event *Event) {
		event.Id = strconv.Itoa(preCount)
		event.ResourceId = strconv.FormatInt(time.Now().UnixNano(), 10)
		preCount += 1
		event.Name = "physicalhost.create;handler=testRouter"
	}

	// Push 3 events
	receivedEvents := map[string]*Event{}
	for i := 0; i < 2; i++ {
		err := prepAndPostEvent("../test_utils/resources/create_virtualbox.json", pre)
		if err != nil {
			t.Fatal(err)
		}
		receivedEvent := awaitEvent(eventsReceived, 500, t)
		if receivedEvent != nil {
			receivedEvents[receivedEvent.Id] = receivedEvent
		}
	}

	if len(receivedEvents) != 2 {
		t.Errorf("Unexpected length %v", len(receivedEvents))
	}
}

func awaitEvent(eventsReceived chan *Event, millisToWait int, t *testing.T) *Event {
	timeout := make(chan bool, 1)
	timeoutFunc := func() {
		time.Sleep(time.Duration(millisToWait) * time.Millisecond)
		timeout <- true
	}
	go timeoutFunc()

	select {
	case e := <-eventsReceived:
		return e
	case <-timeout:
		return nil
	}
	return nil
}

type PreFunc func(*Event)

func prepAndPostEvent(eventFile string, preFunc PreFunc) (err error) {
	rawEvent, err := ioutil.ReadFile(eventFile)
	if err != nil {
		return err
	}

	event := &Event{}
	err = json.Unmarshal(rawEvent, &event)
	if err != nil {
		return err
	}
	preFunc(event)
	rawEvent, err = json.Marshal(event)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = json.Compact(buffer, rawEvent)
	if err != nil {
		return err
	}
	http.Post(pushUrl, "application/json", buffer)

	return nil
}

func TestMain(m *testing.M) {
	ready := make(chan string, 1)
	go tu.InitializeServer(eventServerPort, ready)
	<-ready
	result := m.Run()
	os.Exit(result)
}
