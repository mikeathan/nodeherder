package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/utils"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// requests

	DevicePropertiesUpdated = "devicePropertiesUpdated"
	LoadAutomations         = "loadAutomations"
	LoadDevices             = "loadDevices"
	LoadBridgeFeatures      = "loadBridgeFeatures"
	SaveAutomation          = "saveAutomation"
	DeleteAutomation        = "deleteAutomation"

	// response
	Automations      = "automations"
	Devices          = "devices"
	DeviceAdded      = "deviceAdded"
	DeviceUpdated    = "deviceUpdated"
	BridgeFeatures   = "bridgeFeatures"
	OperationFailed  = "operationFailed"
	OperationSuccess = "operationSuccess"
)

type EventMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (e EventMessage) MarshalJSON() ([]byte, error) {
	p := e.Payload
	if v, ok := e.Payload.([]byte); ok {
		p = string(v)
	}
	return json.Marshal(&struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}{
		Type:    e.Type,
		Payload: p,
	})
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

type WsClient struct {
	hub *wsServer

	conn *websocket.Conn
	send chan []byte
}

func newWsClient(hub *wsServer, conn *websocket.Conn) *WsClient {
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
				utils.LogErrorf("error: %v", err)
			}
			break
		}
		fmt.Println("[DEBUG] received ws message:", string(message))
		c.handleMessage(message)
	}
}

func (c *WsClient) handleMessage(message []byte) {
	var eventMsg = &EventMessage{}

	if err := json.Unmarshal(message, &eventMsg); err != nil {
		return
	}
	switch eventMsg.Type {
	case LoadAutomations:

		msg := c.hub.onLoadAutomations()
		c.Broadcast(Automations, msg)

	case LoadDevices:

		msg := c.hub.onLoadDevices()
		c.Broadcast(Devices, msg)

	case LoadBridgeFeatures:
		msg := c.hub.onLoadBridgeFeatures()
		c.Broadcast(BridgeFeatures, msg)

	case SaveAutomation:

		// todo: do some error checking
		err := c.hub.onSaveAutomation(eventMsg.Payload)
		if err != nil {
			c.Broadcast(OperationFailed, err.Error())
		} else {
			c.Broadcast(OperationSuccess, nil)
		}

	case DeleteAutomation:

		// todo: do some error checking
		c.hub.onDeleteAutomation(eventMsg.Payload)
		c.Broadcast(OperationSuccess, nil)

	default:
		utils.LogWarnf("Unknown event type: %s", eventMsg.Type)
		return
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
	var wsPayload = EventMessage{Type: eventName, Payload: data}
	bytes, err := wsPayload.MarshalJSON()
	if err != nil {
		return errors.New("failed to marshal client payload")
	}
	c.send <- bytes
	return nil
}

type EventHub interface {
	Broadcast(eventName string, data interface{}) error
	RegisterNewClient(conn *websocket.Conn)
	OnLoadAutomations(action func() interface{})
	OnLoadDevices(action func() interface{})
	OnLoadBridgeFeatures(action func() interface{})
	OnSaveAutomation(func(payload interface{}) error)
	OnDeleteAutomation(func(payload interface{}))
}

type wsServer struct {
	clients              map[*WsClient]bool
	broadcast            chan []byte
	register             chan *WsClient
	unregister           chan *WsClient
	onLoadAutomations    func() interface{}
	onLoadDevices        func() interface{}
	onLoadBridgeFeatures func() interface{}
	onSaveAutomation     func(interface{}) error
	onDeleteAutomation   func(interface{})
}

func NewWsHub() EventHub {
	wsHub := &wsServer{
		clients:    map[*WsClient]bool{},
		broadcast:  make(chan []byte),
		register:   make(chan *WsClient),
		unregister: make(chan *WsClient),

		onLoadBridgeFeatures: func() interface{} { return nil },
		onSaveAutomation:     func(i interface{}) error { return nil },
		onDeleteAutomation:   func(payload interface{}) {},
		onLoadDevices:        func() interface{} { return nil },
		onLoadAutomations:    func() interface{} { return nil }}

	go wsHub.run()
	return wsHub
}

func (h *wsServer) OnLoadAutomations(action func() interface{}) {
	h.onLoadAutomations = action
}

func (h *wsServer) OnLoadBridgeFeatures(action func() interface{}) {
	h.onLoadBridgeFeatures = action
}

func (h *wsServer) OnLoadDevices(action func() interface{}) {
	h.onLoadDevices = action
}

func (h *wsServer) OnSaveAutomation(action func(p interface{}) error) {
	h.onSaveAutomation = action
}

func (h *wsServer) OnDeleteAutomation(action func(p interface{})) {
	h.onDeleteAutomation = action
}

func (h *wsServer) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			utils.LogInfo("hub: client registered")
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				utils.LogInfo("hub: client unregistered")
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					utils.LogInfo("broadcast failed, client closed")
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *wsServer) RegisterNewClient(conn *websocket.Conn) {
	client := newWsClient(h, conn)
	client.hub.register <- client

	go client.readPump()
	go client.writePump()
}

func (h *wsServer) Broadcast(eventName string, data interface{}) error {
	var wsData = EventMessage{Type: eventName, Payload: data}
	bytes, err := json.Marshal(wsData)
	if err != nil {
		return errors.New("failed to marshal server payload")
	}
	h.broadcast <- bytes
	return nil
}
