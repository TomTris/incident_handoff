package main

import (
	"encoding/json"
)

type Registry struct {
	hubs          map[string]*Hub
	clientCounter map[string]int
	register      chan *Client
	unregister    chan *Client
	broadcast     chan BroadcastMessage
	done          chan struct{}
	metrics       *RegistryMetric
}

type BroadcastMessage struct {
	msg        json.RawMessage
	incidentID string
}

func NewRegistry(metrics *RegistryMetric) Registry {
	return Registry{
		hubs:          make(map[string]*Hub),
		clientCounter: make(map[string]int),
		register:      make(chan *Client), // no buffered on purpose
		unregister:    make(chan *Client), // no buffered on purpose
		broadcast:     make(chan BroadcastMessage),
		done:          make(chan struct{}),
		metrics:       metrics,
	}
}

func (r *Registry) run() {
	for {
		select {
		case client := <-r.register:
			r.joinRegistry(client)
		case client := <-r.unregister:
			r.leaveRegister(client)
		case broadcast := <-r.broadcast:
			r.broadcastMessage(broadcast)
		// For testing / graceful shutdown
		case <-r.done:
			for _, hub := range r.hubs {
				close(hub.done)
			}
			return
		}
	}
}

func (r *Registry) joinRegistry(client *Client) {
	r.metrics.wsConnections.Inc()

	incID := client.incidentID
	r.clientCounter[incID]++

	// for the first Client
	if r.clientCounter[incID] == 1 {
		r.hubs[incID] = NewHub()
		go r.hubs[incID].run()
	}
	client.hub = r.hubs[incID]
	client.hub.register <- client
}

func (r *Registry) leaveRegister(client *Client) {
	r.metrics.wsConnections.Dec()

	incID := client.incidentID
	r.clientCounter[incID]--

	hub, _ := r.hubs[incID]
	if r.clientCounter[incID] == 0 {
		close(hub.done)
		delete(r.hubs, incID)
		return
	}
	hub.unregister <- client
}

func (r *Registry) broadcastMessage(b BroadcastMessage) {
	hub, ok := r.hubs[b.incidentID]
	if ok {
		hub.broadcast <- b.msg
	}
}
