package hub

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"node-herder/models"
	"time"

	"github.com/gorilla/websocket"
)

const (
	DeviceUpdated   = "deviceUpdated"
	ClientConnected = "connected"
)

type HubMessage struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

func NewHubMessage(name string, payload []byte) HubMessage {
	return HubMessage{Name: name, Payload: json.RawMessage(payload)}
}

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

type WsClient struct {
	hub *WsHub

	conn *websocket.Conn
	send chan []byte
}

func newWsClient(hub *WsHub, conn *websocket.Conn) *WsClient {
	return &WsClient{hub: hub, conn: conn, send: make(chan []byte)}
}
func (c *WsClient) readPump() {
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

func (c *WsClient) writePump() {
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

func (c *WsClient) Broadcast(eventName string, data interface{}) error {
	var wsData = payload{Name: eventName, Data: data}
	bytes, err := json.Marshal(wsData)
	if err != nil {
		return errors.New("failed to marshal client payload")
	}

	c.send <- bytes
	return nil
}

type WsServer interface {
	Broadcast(eventName string, data interface{}) error
	RegisterNewClient(conn *websocket.Conn) models.EventEmitter
}

type WsConfig struct {
	OnConnected func() []byte
}

type WsHub struct {
	clients           map[*WsClient]bool
	broadcast         chan []byte
	register          chan *WsClient
	unregister        chan *WsClient
	OnClientConnected func() []byte
}

func NewWsHub(config *WsConfig) WsServer {
	wsHub := &WsHub{
		clients:           map[*WsClient]bool{},
		broadcast:         make(chan []byte),
		register:          make(chan *WsClient),
		unregister:        make(chan *WsClient),
		OnClientConnected: config.OnConnected,
	}

	go wsHub.run()
	return wsHub
}

func (h *WsHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			fmt.Println("hub: client registered")
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				fmt.Println("hub: client unregistered")
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					fmt.Println("broadcast failed, client closed")
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *WsHub) RegisterNewClient(conn *websocket.Conn) *WsClient {
	client := newWsClient(h, conn)
	client.hub.register <- client

	go client.readPump()
	go client.writePump()

	payload := h.OnClientConnected()
	client.Broadcast(ClientConnected, payload)
	return client
}

func (h *WsHub) Broadcast(eventName string, data interface{}) error {
	var wsData = payload{Name: eventName, Data: data}
	bytes, err := json.Marshal(wsData)
	if err != nil {
		return errors.New("failed to marshal server payload")
	}

	h.broadcast <- bytes
	return nil
}
