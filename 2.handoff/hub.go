package main

type Hub struct {
	// registry *Registry

	// incidentID string
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	done       chan byte
}

func NewHub() *Hub {
	return &Hub{
		// incidentID: incidentID,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client, 8),
		unregister: make(chan *Client, 8),
		broadcast:  make(chan []byte, 256),
		done:       make(chan byte),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			_, ok := h.clients[client]
			if ok == false {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case <-h.done:
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			close(h.register)
			close(h.unregister)
			close(h.broadcast)
			return
		}
	}
}
