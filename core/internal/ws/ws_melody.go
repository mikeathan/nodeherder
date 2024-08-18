package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"node-herder/utils"
	"sync/atomic"

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

	SaveDeviceConfig = "saveDeviceConfig"
	LoadAppconfig    = "loadAppConfig"

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

type EventHubMelogy interface {
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
	HandleRequest(w http.ResponseWriter, r *http.Request) error
}

type wsMelodyServer struct {
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
}

func NewWsHubMelody() EventHubMelogy {

	wsHub := &wsMelodyServer{
		server:                    melody.New(),
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
	}

	return wsHub
}

func (h *wsMelodyServer) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	return h.server.HandleRequest(w, r)
}

func (h *wsMelodyServer) OnDeviceSetValue(action func(p interface{}) error) {
	h.onDeviceSetValue = action
}

func (h *wsMelodyServer) OnDeviceRename(action func(p interface{}) error) {
	h.onDeviceRename = action
}

func (h *wsMelodyServer) OnLoadAutomations(action func() interface{}) {
	h.onLoadAutomations = action
}

func (h *wsMelodyServer) OnLoadDevice(action func(id string) (interface{}, error)) {
	h.onLoadDevice = action
}

func (h *wsMelodyServer) OnLoadAppConfig(action func() (interface{}, error)) {
	h.onLoadAppConfig = action
}

func (h *wsMelodyServer) OnSaveDeviceConfig(action func(payload interface{}) error) {
	h.onSaveDeviceConfig = action
}

func (h *wsMelodyServer) OnLoadMetrics(action func(p interface{}) (interface{}, error)) {
	h.onLoadMetrics = action
}

func (h *wsMelodyServer) OnLoadDeviceList(action func(ids []string) interface{}) {
	h.onLoadDeviceList = action
}

func (h *wsMelodyServer) OnLoadDevices(action func() interface{}) {
	h.onLoadDevices = action
}

func (h *wsMelodyServer) OnSaveAutomation(action func(p interface{}) error) {
	h.onSaveAutomation = action
}

func (h *wsMelodyServer) OnDeleteAutomation(action func(p interface{}) (interface{}, error)) {
	h.onDeleteAutomation = action
}

func (h *wsMelodyServer) OnDeleteAutomationTrigger(action func(p interface{}) (interface{}, error)) {
	h.onDeleteAutomationTrigger = action
}

func (h *wsMelodyServer) EmitDevice(name string) error {
	msg, err := h.onLoadDevice(name)
	if err != nil {
		return err
	}
	h.Broadcast(Device, msg)
	return nil
}

func (h *wsMelodyServer) EmitDeviceList(ids []string) {
	msg := h.onLoadDeviceList(ids)

	h.Broadcast(DeviceList, msg)
}

func (h *wsMelodyServer) EmitDevices() {
	msg := h.onLoadDevices()
	h.Broadcast(Devices, msg)
}

func (h *wsMelodyServer) RegisterNewClient(conn *websocket.Conn) {

}

func (h *wsMelodyServer) Broadcast(eventName string, data interface{}) error {
	var wsData = EventMessage{Type: eventName, Payload: data}
	bytes, err := json.Marshal(wsData)
	if err != nil {
		utils.LogErrorf("wsMelodyServer.Broadcast failed to marshal server payload %s", err.Error())

		return errors.New("failed to marshal server payload")
	}

	return h.server.Broadcast(bytes)
}

func (h *wsMelodyServer) Close() error {
	utils.LogDebugf("wsMelodyServer.Close")
	return h.server.Close()
}

func (h *wsMelodyServer) Start() {

	h.server.HandleConnect(func(s *melody.Session) {
		utils.LogDebugf("wsMelodyServer: New client connected")
		id := clientId.Add(1)

		h.clients[id] = true
		s.Set("id", id)

		s.Write([]byte(fmt.Sprintf("client id %d connected", id)))
	})

	h.server.HandleDisconnect(func(s *melody.Session) {
		if id, ok := s.Get("id"); ok {
			s.Write([]byte(fmt.Sprintf("client id %d disconnected", id)))

			h.clients[id.(int64)] = false
			//h.server.BroadcastOthers([]byte(fmt.Sprintf("dis %d", id)), s)
		}
	})

	h.server.HandleError(func(s *melody.Session, err error) {
		if id, ok := s.Get("id"); ok {
			fmt.Printf("client id %d Session error: %s\n", id, err.Error())
		} // Handle the error
	})

	h.server.HandleClose(func(s *melody.Session, code int, reason string) error {
		if id, ok := s.Get("id"); ok {
			fmt.Printf("client id %d Session closed: %d, %s\n", id, code, reason)
		}
		// do cleanup
		return nil
	})

	h.server.HandleMessage(func(s *melody.Session, msg []byte) {
		h.handleHubEvents(msg)
	})
}

func (c *wsMelodyServer) handleHubEvents(message []byte) {
	var eventMsg = &EventMessage{}

	if err := json.Unmarshal(message, &eventMsg); err != nil {
		utils.LogWarnf("handleMessage unmarshal error: %s", err.Error())
		return
	}

	switch eventMsg.Type {

	case LoadAutomations:
		msg := c.onLoadAutomations()
		c.Broadcast(Automations, msg)

	case LoadDevices:
		msg := c.onLoadDevices()
		c.Broadcast(Devices, msg)

	// case LoadMetrics:
	// 	c.executePayloadActionWithEvent(eventMsg.Payload, c.hub.onLoadMetrics, Metrics)

	// case LoadAppconfig:
	// 	c.executeActionWithEvent(c.hub.onLoadAppConfig, AppConfig)

	// case SaveDeviceConfig:
	// 	c.executeAction(eventMsg.Payload, c.hub.onSaveDeviceConfig, true)

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

func (c *wsMelodyServer) executeActionWithEvent(action func() (interface{}, error), successEvent string) {

	result, err := action()
	if err != nil {
		c.Broadcast(OperationFailed, err.Error())
	} else {
		c.Broadcast(successEvent, result)
	}
}

func (c *wsMelodyServer) executePayloadActionWithEvent(payload interface{}, action func(interface{}) (interface{}, error), successEvent string) {

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

func (c *wsMelodyServer) executeAction(payload interface{}, action func(interface{}) error, reportSuccess bool) {
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
