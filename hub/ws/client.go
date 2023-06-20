package ws

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	DeviceUpdated   = "deviceUpdated"
	ClientConnected = "connected"
)

type payload struct {
	Name string
	Data interface{}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	hub *Server

	conn *websocket.Conn
	send chan []byte
}

func newClient(hub *Server, conn *websocket.Conn) *Client {
	return &Client{hub: hub, conn: conn, send: make(chan []byte)}
}

func RegisterConnection(hub *Server, conn *websocket.Conn) *Client {
	client := newClient(hub, conn)
	client.hub.register <- client

	go client.readPump()
	go client.writePump()
	return client
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.hub.broadcast <- message
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
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Broadcast(eventName string, data interface{}) error {
	var wsData = payload{Name: eventName, Data: data}
	bytes, err := json.Marshal(wsData)
	if err != nil {
		return errors.New("failed to marshal client payload")
	}

	c.send <- bytes
	return nil
}
