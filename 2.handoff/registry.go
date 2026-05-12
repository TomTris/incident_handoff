package main

import "encoding/json"

type Registry struct {
	hubs          map[string]*Hub
	clientCounter map[string]int
	register      chan *Client
	unregister    chan *Client
	broadcast     chan BroadcastMessage
}

type BroadcastMessage struct {
	msg        json.RawMessage
	incidentID string
}

func NewRegistry() Registry {
	return Registry{
		hubs:          make(map[string]*Hub),
		clientCounter: make(map[string]int),
		register:      make(chan *Client), // no buffered on purpose
		unregister:    make(chan *Client), // no buffered on purpose
		broadcast:     make(chan BroadcastMessage),
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
		}
	}
}

func (r *Registry) joinRegistry(client *Client) {
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
	incID := client.incidentID
	r.clientCounter[incID]--

	hub, _ := r.hubs[incID]
	if r.clientCounter[incID] == 0 {
		close(hub.done)
		r.hubs[incID] = nil
		return
	}
	hub.unregister <- client
}

func (r *Registry) broadcastMessage(b BroadcastMessage) {
	r.hubs[b.incidentID].broadcast <- b.msg
}
