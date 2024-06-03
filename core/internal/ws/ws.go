package ws

import (
	"encoding/json"
	"errors"
	"node-herder/utils"
	"time"

	"github.com/gorilla/websocket"
)

const (

	// requests
	LoadAutomations = "loadAutomations"
	LoadDevices     = "loadDevices"
	LoadDevice      = "loadDevice"
	LoadDeviceList  = "loadDeviceList"

	SaveAutomation          = "saveAutomation"
	DeleteAutomation        = "deleteAutomation"
	DeleteAutomationTrigger = "deleteAutomationTrigger"
	DeviceSetValue          = "deviceSetValue"
	DeviceRename            = "deviceRename"

	LoadMetrics = "loadMetrics"

	// response
	Automations       = "automations"
	Devices           = "devices"
	DeviceList        = "deviceList"
	Device            = "device"
	DeviceAdded       = "deviceAdded"
	DeviceUpdated     = "deviceUpdated" // returns back updated properties of type DeviceUpdated
	OperationFailed   = "operationFailed"
	OperationSuccess  = "operationSuccess"
	AutomationUpdated = "automationUpdated" // returns back upated automation

	Metrics        = "metrics"
	MetricsEnabled = "metricsEnabled"
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
	maxMessageSize = 5 * 1048
)

type WsClient struct {
	hub *wsServer

	conn *websocket.Conn
	send chan []byte
}

// NOTE:
// websocket closes - then buffered channel didnt work
// read for fix: https://stackoverflow.com/questions/66104210/gorilla-websocket-example-hangs-when-trying-to-send-data-to-a-channel-whilst-han
func newWsClient(hub *wsServer, conn *websocket.Conn) *WsClient {
	return &WsClient{hub: hub, conn: conn, send: make(chan []byte, 1024)}
}

func (c *WsClient) readPump() {
	defer func() {
		utils.LogError("readPump closed")
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		utils.LogDebugf("ws SetPongHandler")
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			utils.LogErrorf("ws ReadMessage error: %s", err.Error())
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.LogErrorf("ws IsUnexpectedCloseError : %v", err)
			}
			break
		}
		utils.LogDebugf("ws received: %s", string(message))
		c.handleMessage(message)
	}
}

func (c *WsClient) handleMessage(message []byte) {
	var eventMsg = &EventMessage{}

	if err := json.Unmarshal(message, &eventMsg); err != nil {
		utils.LogWarnf("handleMessage unmarshal error: %s", err.Error())
		return
	}

	switch eventMsg.Type {

	case LoadAutomations:
		msg := c.hub.onLoadAutomations()
		c.Broadcast(Automations, msg)

	case LoadDevices:
		msg := c.hub.onLoadDevices()
		c.Broadcast(Devices, msg) 

	case LoadMetrics:
		c.executeActionWithEvent(eventMsg.Payload, c.hub.onLoadMetrics, Metrics)

	case SaveAutomation:
		c.executeAction(eventMsg.Payload, c.hub.onSaveAutomation, true)

	case DeleteAutomation:
		c.executeActionWithEvent(eventMsg.Payload, c.hub.onDeleteAutomation, Automations)

	case DeleteAutomationTrigger:
		c.executeActionWithEvent(eventMsg.Payload, c.hub.onDeleteAutomationTrigger, AutomationUpdated)

	case DeviceSetValue:
		c.executeAction(eventMsg.Payload, c.hub.onDeviceSetValue, false)

	case DeviceRename:
		c.executeAction(eventMsg.Payload, c.hub.onDeviceRename, false)

	default:

		utils.LogWarnf("Unknown event type: %s", eventMsg.Type)
		return
	}
}

func (c *WsClient) executeActionWithEvent(payload interface{}, action func(interface{}) (interface{}, error), successEvent string) {

	if payload == nil {
		c.Broadcast(OperationFailed, "payload is empty")
		return
	}

	result, err := action(payload)
	if err != nil {
		c.Broadcast(OperationFailed, err.Error())
	} else {
		c.Broadcast(successEvent, result)
	}
}

func (c *WsClient) executeAction(payload interface{}, action func(interface{}) error, reportSuccess bool) {
	if payload == nil {
		c.Broadcast(OperationFailed, "payload is empty")
		return
	}

	err := action(payload)
	if err != nil {
		c.Broadcast(OperationFailed, err.Error())
	} else if reportSuccess {
		c.Broadcast(OperationSuccess, nil)
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
				utils.LogErrorf("ws writePump message %s failed", string(message))
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				utils.LogErrorf("ws NextWriter error %s ", err.Error())
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				utils.LogErrorf("ws close() error %s ", err.Error())
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				utils.LogErrorf("Ping error %s", err.Error())
				return
			}
		}
	}
}

func (c *WsClient) Broadcast(eventName string, data interface{}) error {
	var wsPayload = EventMessage{Type: eventName, Payload: data}
	bytes, err := wsPayload.MarshalJSON()
	if err != nil {
		utils.LogErrorf("WsClient.Broadcast failed to marshal server payload %s", err.Error())

		return errors.New("failed to marshal client payload")
	}
	//fmt.Print(string(bytes))
	c.send <- bytes
	return nil
}

type EventHub interface {
	Broadcast(eventName string, data interface{}) error
	RegisterNewClient(conn *websocket.Conn)
	EmitDevices()
	EmitDeviceList(names []string)
	EmitDevice(name string) error
	OnLoadAutomations(action func() interface{})
	OnLoadDevices(action func() interface{})
	OnLoadDevice(action func(id string) (interface{}, error))
	OnLoadDeviceList(action func(ids []string) interface{})
	OnDeviceSetValue(func(payload interface{}) error)
	OnDeviceRename(func(payload interface{}) error)
	OnSaveAutomation(func(payload interface{}) error)
	OnDeleteAutomation(func(payload interface{}) (interface{}, error))
	OnDeleteAutomationTrigger(func(payload interface{}) (interface{}, error))
	OnLoadMetrics(action func(interface{}) (interface{}, error))
}

type wsServer struct {
	clients                   map[*WsClient]bool
	broadcast                 chan []byte
	register                  chan *WsClient
	unregister                chan *WsClient
	onLoadAutomations         func() interface{}
	onLoadDevices             func() interface{}
	onLoadDeviceList          (func(ids []string) interface{})
	onLoadDevice              func(id string) (interface{}, error)
	onLoadMetrics             func(interface{}) (interface{}, error)
	onSaveAutomation          func(interface{}) error
	onDeviceSetValue          func(interface{}) error
	onDeviceRename            func(interface{}) error
	onDeleteAutomation        func(interface{}) (interface{}, error)
	onDeleteAutomationTrigger func(interface{}) (interface{}, error)
}

func NewWsHub() EventHub {
	wsHub := &wsServer{
		clients:    map[*WsClient]bool{},
		broadcast:  make(chan []byte),
		register:   make(chan *WsClient),
		unregister: make(chan *WsClient),

		onSaveAutomation: func(payload interface{}) error { return nil },
		onLoadMetrics:    func(interface{}) (interface{}, error) { return nil, nil },

		onDeleteAutomation:        func(payload interface{}) (interface{}, error) { return nil, nil },
		onDeleteAutomationTrigger: func(payload interface{}) (interface{}, error) { return nil, nil },
		onDeviceSetValue:          func(payload interface{}) error { return nil },
		onDeviceRename:            func(payload interface{}) error { return nil },
		onLoadDevice:              func(id string) (interface{}, error) { return nil, nil },
		onLoadDeviceList:          func(ids []string) interface{} { return nil },
		onLoadDevices:             func() interface{} { return nil },
		onLoadAutomations:         func() interface{} { return nil }}

	go wsHub.run()
	return wsHub
}

func (h *wsServer) EmitDevice(name string) error {
	msg, err := h.onLoadDevice(name)
	if err != nil {
		return err
	}
	h.Broadcast(Device, msg)
	return nil
}

func (h *wsServer) EmitDeviceList(ids []string) {
	msg := h.onLoadDeviceList(ids)

	h.Broadcast(DeviceList, msg)
}

func (h *wsServer) EmitDevices() {
	msg := h.onLoadDevices()
	h.Broadcast(Devices, msg)
}

func (h *wsServer) OnDeviceSetValue(action func(p interface{}) error) {
	h.onDeviceSetValue = action
}

func (h *wsServer) OnDeviceRename(action func(p interface{}) error) {
	h.onDeviceRename = action
}

func (h *wsServer) OnLoadAutomations(action func() interface{}) {
	h.onLoadAutomations = action
}

func (h *wsServer) OnLoadDevice(action func(id string) (interface{}, error)) {
	h.onLoadDevice = action
}

func (h *wsServer) OnLoadMetrics(action func(p interface{}) (interface{}, error)) {
	h.onLoadMetrics = action
}

func (h *wsServer) OnLoadDeviceList(action func(ids []string) interface{}) {
	h.onLoadDeviceList = action
}

func (h *wsServer) OnLoadDevices(action func() interface{}) {
	h.onLoadDevices = action
}

func (h *wsServer) OnSaveAutomation(action func(p interface{}) error) {
	h.onSaveAutomation = action
}

func (h *wsServer) OnDeleteAutomation(action func(p interface{}) (interface{}, error)) {
	h.onDeleteAutomation = action
}

func (h *wsServer) OnDeleteAutomationTrigger(action func(p interface{}) (interface{}, error)) {
	h.onDeleteAutomationTrigger = action
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
					utils.LogErrorf("broadcast failed, client closed. last message %s", string(message))
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
		utils.LogErrorf("wsServer.Broadcast failed to marshal server payload %s", err.Error())

		return errors.New("failed to marshal server payload")
	}
	h.broadcast <- bytes
	return nil
}
