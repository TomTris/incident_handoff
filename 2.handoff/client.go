package main

import (
	"bytes"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 1024
)

type Client struct {
	incidentID string
	registry   *Registry //for establishing and removing connection
	hub        *Hub      //for handling realtime messages - no needed for handoff
	conn       *websocket.Conn
	send       chan []byte
}

func newClient(incidentID string, conn *websocket.Conn) *Client {
	return &Client{
		incidentID: incidentID,
		conn:       conn,
		send:       make(chan []byte, 256),
	}
}

func (c *Client) joinRegistry(registry *Registry) {
	c.registry = registry
	c.registry.register <- c
}

func (c *Client) readPump() {
	defer func() {
		c.registry.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { return c.conn.SetReadDeadline(time.Now().Add(pongWait)) })
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("unexpected close", "err", err)
			}
			break
		}
		// these 2 lines are not important for Handoff yet.
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		// c.hub.broadcast <- msg
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if ok == false {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				slog.Error("WriteMessage broken", "err", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				return
			}
		}
	}
}
