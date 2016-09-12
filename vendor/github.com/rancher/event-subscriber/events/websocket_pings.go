package events

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

type PingConfig struct {
	SendPingInterval  int
	CheckPongInterval int
	MaxPongWait       int
}

var DefaultPingConfig = PingConfig{
	SendPingInterval:  5000,
	CheckPongInterval: 5000,
	MaxPongWait:       10000,
}

func (router *EventRouter) sendWebsocketPings() {
	log.Infof("Starting websocket pings")
	ticker := time.NewTicker(time.Millisecond * time.Duration(router.pingConfig.SendPingInterval))
	defer ticker.Stop()
	for range ticker.C {
		if err := router.eventStream.WriteControl(websocket.PingMessage, []byte(""), time.Now().Add(time.Second)); err != nil {
			// websocket closed, return
			log.Warnf("websocket closed: %s", err)
			return
		}
	}
}

func newPongHandler(r *EventRouter) *pongHandler {
	ph := &pongHandler{
		r:        r,
		mu:       &sync.Mutex{},
		lastPing: time.Now(),
		done:     make(chan bool),
	}

	go ph.startTimer(r.pingConfig.CheckPongInterval, r.pingConfig.MaxPongWait)

	return ph
}

type pongHandler struct {
	r        *EventRouter
	mu       *sync.Mutex
	lastPing time.Time
	done     chan bool
}

func (h *pongHandler) startTimer(checkInterval, maxWait int) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(checkInterval))
	defer ticker.Stop()
	for {
		select {
		case <-h.done:
			return
		case <-ticker.C:
			h.mu.Lock()
			t := h.lastPing
			h.mu.Unlock()
			timeoutAt := t.Add(time.Millisecond * time.Duration(maxWait))
			if time.Now().After(timeoutAt) {
				// bad!
				log.Infof("Hit websocket pong timeout. Last websocket ping received at %v. Closing connection.", t)
				h.r.Stop()
			}
		}
	}
}

func (h *pongHandler) handle(appData string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastPing = time.Now()
	return nil
}

func (h *pongHandler) stop() {
	h.done <- true
}
