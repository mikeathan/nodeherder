package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"node-herder/utils"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
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

	SaveHistoryConfig = "saveHistoryConfig"
	SaveDeviceConfig  = "saveDeviceConfig"
	LoadAppconfig     = "loadAppConfig"

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

	Metrics   = "metrics"
	AppConfig = "appConfig"

	// Maximum message size allowed from peer.
	maxMessageSize = 1 * 1024 * 1024

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var clientId atomic.Int64

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

type EventHub interface {
	Broadcast(eventName string, data interface{}) error
	Start()
	Close() error
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
	OnLoadAppConfig(action func() (interface{}, error))
	OnSaveDeviceConfig(func(payload interface{}) error)
	OnSaveHistoryConfig(func(payload interface{}) error)
	HandleRequest(w http.ResponseWriter, r *http.Request) error
}

type wsServer struct {
	clients                   map[int64]bool
	server                    *melody.Melody
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
	onLoadAppConfig           func() (interface{}, error)
	onSaveDeviceConfig        func(interface{}) error
	onSaveHistoryConfig       func(interface{}) error
}

func NewWsHub() EventHub {

	server := melody.New()
	server.Upgrader.ReadBufferSize = maxMessageSize
	server.Upgrader.WriteBufferSize = maxMessageSize
	server.Config.PingPeriod = pingPeriod
	server.Config.PongWait = pongWait
	server.Config.WriteWait = writeWait
	server.Config.MaxMessageSize = maxMessageSize
	server.Config.MessageBufferSize = maxMessageSize
	//		ConcurrentMessageHandling bool          // Handle messages from sessions concurrently.

	wsHub := &wsServer{
		server:                    server,
		clients:                   map[int64]bool{},
		onSaveAutomation:          func(payload interface{}) error { return nil },
		onLoadMetrics:             func(interface{}) (interface{}, error) { return nil, nil },
		onDeleteAutomation:        func(payload interface{}) (interface{}, error) { return nil, nil },
		onDeleteAutomationTrigger: func(payload interface{}) (interface{}, error) { return nil, nil },
		onDeviceSetValue:          func(payload interface{}) error { return nil },
		onDeviceRename:            func(payload interface{}) error { return nil },
		onLoadDevice:              func(id string) (interface{}, error) { return nil, nil },
		onLoadDeviceList:          func(ids []string) interface{} { return nil },
		onLoadDevices:             func() interface{} { return nil },
		onLoadAutomations:         func() interface{} { return nil },
		onLoadAppConfig:           func() (interface{}, error) { return nil, nil },
		onSaveDeviceConfig:        func(payload interface{}) error { return nil },
		onSaveHistoryConfig:       func(payload interface{}) error { return nil },
	}

	return wsHub
}

func (h *wsServer) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	return h.server.HandleRequest(w, r)
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

func (h *wsServer) OnLoadAppConfig(action func() (interface{}, error)) {
	h.onLoadAppConfig = action
}

func (h *wsServer) OnSaveDeviceConfig(action func(payload interface{}) error) {
	h.onSaveDeviceConfig = action
}

func (h *wsServer) OnSaveHistoryConfig(action func(payload interface{}) error) {
	h.onSaveHistoryConfig = action
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

func (h *wsServer) RegisterNewClient(conn *websocket.Conn) {

}

func (h *wsServer) Broadcast(eventName string, data interface{}) error {
	var wsData = EventMessage{Type: eventName, Payload: data}
	bytes, err := json.Marshal(wsData)
	if err != nil {
		utils.LogErrorf("wsMelodyServer.Broadcast failed to marshal server payload %s", err.Error())

		return errors.New("failed to marshal server payload")
	}

	return h.server.Broadcast(bytes)
}

func (h *wsServer) Close() error {
	utils.LogDebugf("wsMelodyServer.Close")
	return h.server.Close()
}

func (h *wsServer) Start() {

	h.server.HandleConnect(func(s *melody.Session) {
		utils.LogDebugf("wsMelodyServer: New client connected")
		id := clientId.Add(1)

		h.clients[id] = true
		s.Set("id", id)

		//s.Write([]byte(fmt.Sprintf("client id %d connected", id)))
	})

	h.server.HandleDisconnect(func(s *melody.Session) {
		if id, ok := s.Get("id"); ok {
			s.Write([]byte(fmt.Sprintf("client id %d disconnected", id)))

			h.clients[id.(int64)] = false
			h.server.BroadcastOthers([]byte(fmt.Sprintf("dis %d", id)), s)
		} else {
			fmt.Println("client diconnected")
		}
	})

	h.server.HandleError(func(s *melody.Session, err error) {
		if id, ok := s.Get("id"); ok {
			fmt.Printf("client id %d Session error: %s\n", id, err.Error())
		} else {
			fmt.Printf("client Session error: %s\n", err.Error())
		}
	})

	h.server.HandleClose(func(s *melody.Session, code int, reason string) error {
		if id, ok := s.Get("id"); ok {
			fmt.Printf("client id %d Session closed: %d, %s\n", id, code, reason)
		} else {
			fmt.Println("client session closed")
		}

		return nil
	})

	h.server.HandleMessage(func(s *melody.Session, msg []byte) {
		// Check message size here:
		if len(msg) > maxMessageSize {
			// Handle message too large error
			fmt.Println("message too large")
			s.CloseWithMsg([]byte(fmt.Sprintf("%d message too large", websocket.CloseMessageTooBig)))
			return
		}
		h.handleHubEvents(msg)
	})
}

func (c *wsServer) handleHubEvents(message []byte) {
	var eventMsg = &EventMessage{}

	if err := json.Unmarshal(message, &eventMsg); err != nil {
		utils.LogWarnf("wsServer handleHubEvents:  unmarshal error: %s", err.Error())
		return
	}

	switch eventMsg.Type {

	case LoadAutomations:
		msg := c.onLoadAutomations()
		c.Broadcast(Automations, msg)

	case LoadDevices:
		msg := c.onLoadDevices()
		c.Broadcast(Devices, msg)

	case LoadMetrics:
		c.executePayloadActionWithEvent(eventMsg.Payload, c.onLoadMetrics, Metrics)

	case LoadAppconfig:
		c.executeActionWithEvent(c.onLoadAppConfig, AppConfig)

	case SaveDeviceConfig:
		c.executeAction(eventMsg.Payload, c.onSaveDeviceConfig, true)

	case SaveHistoryConfig:
		c.executeAction(eventMsg.Payload, c.onSaveHistoryConfig, true)

	case SaveAutomation:
		c.executeAction(eventMsg.Payload, c.onSaveAutomation, true)

	case DeleteAutomation:
		c.executePayloadActionWithEvent(eventMsg.Payload, c.onDeleteAutomation, Automations)

	case DeleteAutomationTrigger:
		c.executePayloadActionWithEvent(eventMsg.Payload, c.onDeleteAutomationTrigger, AutomationUpdated)

	case DeviceSetValue:
		c.executeAction(eventMsg.Payload, c.onDeviceSetValue, false)

	case DeviceRename:
		c.executeAction(eventMsg.Payload, c.onDeviceRename, false)

	default:

		utils.LogWarnf("Unknown event type: %s", eventMsg.Type)
		return
	}
}

func (c *wsServer) executeActionWithEvent(action func() (interface{}, error), successEvent string) {

	result, err := action()
	if err != nil {
		c.Broadcast(OperationFailed, err.Error())
	} else {
		c.Broadcast(successEvent, result)
	}
}

func (c *wsServer) executePayloadActionWithEvent(payload interface{}, action func(interface{}) (interface{}, error), successEvent string) {

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

func (c *wsServer) executeAction(payload interface{}, action func(interface{}) error, reportSuccess bool) {
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
