package ws

import (
	"encoding/json"
	"net/http"
	"node-herder/models/hub"
	"node-herder/models/settings"
	"node-herder/utils"
)

const (

	// requests
	LoadAutomations = "loadAutomations"
	LoadDevice     = "loadDevice"
	LoadDeviceList = "loadDeviceList"

	SaveAutomation          = "saveAutomation"
	DeleteAutomation        = "deleteAutomation"
	DeleteAutomationTrigger = "deleteAutomationTrigger"
	DeviceSetValue          = "deviceSetValue"
	DeviceRename            = "deviceRename"
	DeviceRemove            = "deviceRemove"
	DeviceInterview         = "deviceInterview"

	BridgePermitJoin           = "bridgePermitJoin"
	SaveLoggerConfig           = "saveLoggerConfig"
	SaveHistoryConfig          = "saveHistoryConfig"
	SaveDeviceConfigOverride   = "saveDeviceConfigOverride"
	DeleteDeviceConfigOverride = "deleteDeviceConfigOverride"
	SaveDeviceConfigDefaults   = "saveDeviceConfigDefaults"
	SaveDashboardGroup         = "saveDashboardGroup"
	RenameDashboardGroup       = "renameDashboardGroup"
	DeleteDashboardGroup       = "deleteDashboardGroup"
	ImportDashboardGroups      = "importDashboardGroups"
	LoadDashboardGroups        = "loadDashboardGroups"
	LoadAppconfig              = "loadAppConfig"

	LoadMetrics = "loadMetrics"

	// response
	Automations       = "automations"
	DeviceList        = "deviceList"
	Device            = "device"
	DeviceAdded       = "deviceAdded"
	DeviceUpdated     = "deviceUpdated" // returns back updated properties of type DeviceUpdated
	OperationFailed   = "operationFailed"
	OperationSuccess  = "operationSuccess"
	AutomationUpdated = "automationUpdated" // returns back upated automation

	Metrics         = "metrics"
	AppConfig       = "appConfig"
	BridgeConfig    = "bridgeConfig"
	DashboardGroups = "dashboardGroups"
)

type EventHub interface {
	Broadcast(eventName string, data interface{}) error
	Start()
	Close() error
	EmitBridgeConfig()
	EmitDeviceList(names []string)
	EmitDevice(name string) error
	OnLoadAutomations(action func() interface{})
	OnLoadDevices(action func() interface{})
	OnLoadDevice(action func(id string) (interface{}, error))
	OnLoadDeviceList(action func(ids []string) interface{})
	OnDeviceSetValue(func(payload interface{}) error)
	OnDeviceRename(func(payload interface{}) error)
	OnDeviceRemove(func(payload interface{}) error)
	OnDeviceInterview(func(payload interface{}) error)
	OnBridgePermitJoin(func(payload interface{}) error)
	OnSaveAutomation(func(payload interface{}) error)
	OnDeleteAutomation(func(payload interface{}) (interface{}, error))
	OnDeleteAutomationTrigger(func(payload interface{}) (interface{}, error))
	OnLoadMetrics(action func(interface{}) (interface{}, error))
	OnLoadAppConfig(action func() (interface{}, error))
	OnLoadBridgeConfig(action func() (interface{}, error))
	OnSaveDeviceConfigOverride(func(payload interface{}) error)
	OnDeleteDeviceConfigOverride(func(payload interface{}) error)
	OnSaveDeviceConfigDefaults(func(payload interface{}) error)
	OnSaveHistoryConfig(func(payload interface{}) error)
	OnSaveLoggerConfig(func(payload interface{}) error)
	OnSaveDashboardGroup(action func(payload interface{}) error)
	OnRenameDashboardGroup(action func(payload interface{}) error)
	OnImportDashboardGroups(action func(payload interface{}) error)
	OnLoadDashboardGroups(action func() (interface{}, error))
	OnDeleteDashboardGroup(action func(payload interface{}) error)
	HandleRequest(w http.ResponseWriter, r *http.Request) error
	Context() hub.Context
}

type eventHubImpl struct {
	server                       WebSocket
	onLoadAutomations            func() interface{}
	onLoadDevices                func() interface{}
	onLoadDeviceList             (func([]string) interface{})
	onLoadDevice                 func(id string) (interface{}, error)
	onLoadMetrics                func(interface{}) (interface{}, error)
	onSaveAutomation             func(interface{}) error
	onDeviceSetValue             func(interface{}) error
	onDeviceRename               func(interface{}) error
	onDeviceRemove               func(interface{}) error
	onDeviceInterview            func(interface{}) error
	onBridgePermitJoin           func(interface{}) error
	onDeleteAutomation           func(interface{}) (interface{}, error)
	onDeleteAutomationTrigger    func(interface{}) (interface{}, error)
	onLoadAppConfig              func() (interface{}, error)
	onLoadBridgeConfig           func() (interface{}, error)
	onSaveDeviceConfigOverride   func(interface{}) error
	onDeleteDeviceConfigOverride func(interface{}) error
	onSaveDeviceConfigDefaults   func(interface{}) error
	onSaveHistoryConfig          func(interface{}) error
	onSaveLoggerConfig           func(interface{}) error
	onSaveDashboardGroup         func(interface{}) error
	onRenameDashboardGroup       func(interface{}) error
	onDeleteDashboardGroup       func(payload interface{}) error
	onImportDashboardGroups      func(payload interface{}) error
	onLoadDashboardGroups        func() (interface{}, error)
	requestContext hub.Context
}

func NewWsHub() EventHub {
	return &eventHubImpl{
		server:                       NewWebSocket(),
		onSaveAutomation:             func(payload interface{}) error { return nil },
		onLoadMetrics:                func(interface{}) (interface{}, error) { return nil, nil },
		onDeleteAutomation:           func(payload interface{}) (interface{}, error) { return nil, nil },
		onDeleteAutomationTrigger:    func(payload interface{}) (interface{}, error) { return nil, nil },
		onDeviceSetValue:             func(payload interface{}) error { return nil },
		onDeviceRename:               func(payload interface{}) error { return nil },
		onDeviceRemove:               func(payload interface{}) error { return nil },
		onDeviceInterview:            func(payload interface{}) error { return nil },
		onBridgePermitJoin:           func(payload interface{}) error { return nil },
		onLoadDevice:                 func(id string) (interface{}, error) { return nil, nil },
		onLoadDeviceList:             func(ids []string) interface{} { return nil },
		onLoadDevices:                func() interface{} { return nil },
		onLoadAutomations:            func() interface{} { return nil },
		onLoadAppConfig:              func() (interface{}, error) { return nil, nil },
		onLoadBridgeConfig:           func() (interface{}, error) { return nil, nil },
		onSaveDeviceConfigOverride:   func(payload interface{}) error { return nil },
		onDeleteDeviceConfigOverride: func(payload interface{}) error { return nil },
		onSaveDeviceConfigDefaults:   func(payload interface{}) error { return nil },
		onSaveHistoryConfig:          func(payload interface{}) error { return nil },
		onSaveLoggerConfig:           func(payload interface{}) error { return nil },
		onSaveDashboardGroup:         func(payload interface{}) error { return nil },
		onDeleteDashboardGroup:       func(payload interface{}) error { return nil },
		onRenameDashboardGroup:       func(payload interface{}) error { return nil },
		onImportDashboardGroups:      func(payload interface{}) error { return nil },
		onLoadDashboardGroups:        func() (interface{}, error) { return nil, nil },
		requestContext: NewRequestContext(),
	}
}

func (h *eventHubImpl) Context() hub.Context {
	return h.requestContext
}

func (h *eventHubImpl) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	return h.server.HandleRequest(w, r)
}

func (h *eventHubImpl) OnDeviceSetValue(action func(p interface{}) error) {
	h.onDeviceSetValue = action
}

func (h *eventHubImpl) OnDeviceRename(action func(p interface{}) error) {
	h.onDeviceRename = action
}

func (h *eventHubImpl) OnDeviceRemove(action func(p interface{}) error) {
	h.onDeviceRemove = action
}

func (h *eventHubImpl) OnDeviceInterview(action func(p interface{}) error) {
	h.onDeviceInterview = action
}

func (h *eventHubImpl) OnBridgePermitJoin(action func(p interface{}) error) {
	h.onBridgePermitJoin = action
}

func (h *eventHubImpl) OnLoadAutomations(action func() interface{}) {
	h.onLoadAutomations = action
}

func (h *eventHubImpl) OnLoadDevice(action func(id string) (interface{}, error)) {
	h.onLoadDevice = action
}

func (h *eventHubImpl) OnLoadAppConfig(action func() (interface{}, error)) {
	h.onLoadAppConfig = action
}

func (h *eventHubImpl) OnLoadBridgeConfig(action func() (interface{}, error)) {
	h.onLoadBridgeConfig = action
}

func (h *eventHubImpl) OnSaveDeviceConfigOverride(action func(payload interface{}) error) {
	h.onSaveDeviceConfigOverride = action
}

func (h *eventHubImpl) OnDeleteDeviceConfigOverride(action func(payload interface{}) error) {
	h.onDeleteDeviceConfigOverride = action
}

func (h *eventHubImpl) OnSaveDeviceConfigDefaults(action func(payload interface{}) error) {
	h.onSaveDeviceConfigDefaults = action
}

func (h *eventHubImpl) OnSaveHistoryConfig(action func(payload interface{}) error) {
	h.onSaveHistoryConfig = action
}

func (h *eventHubImpl) OnLoadMetrics(action func(p interface{}) (interface{}, error)) {
	h.onLoadMetrics = action
}

func (h *eventHubImpl) OnLoadDeviceList(action func(ids []string) interface{}) {
	h.onLoadDeviceList = action
}

func (h *eventHubImpl) OnLoadDevices(action func() interface{}) {
	h.onLoadDevices = action
}

func (h *eventHubImpl) OnSaveAutomation(action func(p interface{}) error) {
	h.onSaveAutomation = action
}

func (h *eventHubImpl) OnDeleteAutomation(action func(p interface{}) (interface{}, error)) {
	h.onDeleteAutomation = action
}

func (h *eventHubImpl) OnDeleteAutomationTrigger(action func(p interface{}) (interface{}, error)) {
	h.onDeleteAutomationTrigger = action
}

func (h *eventHubImpl) OnSaveLoggerConfig(action func(payload interface{}) error) {
	h.onSaveLoggerConfig = action
}

func (h *eventHubImpl) OnSaveDashboardGroup(action func(payload interface{}) error) {
	h.onSaveDashboardGroup = action
}

func (h *eventHubImpl) OnRenameDashboardGroup(action func(payload interface{}) error) {
	h.onRenameDashboardGroup = action
}

func (h *eventHubImpl) OnDeleteDashboardGroup(action func(payload interface{}) error) {
	h.onDeleteDashboardGroup = action
}

func (h *eventHubImpl) OnImportDashboardGroups(action func(payload interface{}) error) {
	h.onImportDashboardGroups = action
}

func (h *eventHubImpl) OnLoadDashboardGroups(action func() (interface{}, error)) {
	h.onLoadDashboardGroups = action
}

func (h *eventHubImpl) EmitDevice(name string) error {
	msg, err := h.onLoadDevice(name)
	if err != nil {
		return err
	}
	h.Broadcast(Device, msg)
	return nil
}

func (h *eventHubImpl) EmitBridgeConfig() {
	cfg, err := h.onLoadBridgeConfig()
	if err != nil {
		utils.LogErrorf("Failed to load bridge config: %v", err)
		return
	}

	h.Broadcast(BridgeConfig, cfg)
}

func (h *eventHubImpl) EmitDeviceList(ids []string) {
	msg := h.onLoadDeviceList(ids)

	h.Broadcast(DeviceList, msg)
}

func (h *eventHubImpl) Broadcast(eventName string, data interface{}) error {
	return h.server.Broadcast(eventName, data)
}

func (h *eventHubImpl) Close() error {
	return h.server.Close()
}

func (h *eventHubImpl) Start() {
	h.server.Start(func(message []byte) { h.handleHubEvents(message) })
}

// TODO: abstract this so we can mock it
// https://gemini.google.com/app/b3118f0d9cdcdba6
func (c *eventHubImpl) handleHubEvents(message []byte) {
	var eventMsg = &EventMessage{}

	if err := json.Unmarshal(message, &eventMsg); err != nil {
		utils.LogWarnf("wsServer handleHubEvents:  unmarshal error: %s", err.Error())
		return
	}

	switch eventMsg.Type {

	case LoadAutomations:
		msg := c.onLoadAutomations()
		err := c.Broadcast(Automations, msg)
		if err != nil {
			utils.LogErrorf("Failed to broadcast onLoadAutomations %s", err.Error())
		}

	case LoadMetrics:
		c.executePayloadActionWithSuccessfullyEvent(eventMsg.Payload, c.onLoadMetrics, Metrics)

	case LoadAppconfig:
		c.executeActionWithEvent(c.onLoadAppConfig, AppConfig)

	case SaveDeviceConfigOverride:
		c.executeAction(eventMsg.Payload, c.onSaveDeviceConfigOverride, true)

	case DeleteDeviceConfigOverride:
		c.executeAction(eventMsg.Payload, c.onDeleteDeviceConfigOverride, true)

	case SaveDeviceConfigDefaults:
		c.executeAction(eventMsg.Payload, c.onSaveDeviceConfigDefaults, true)

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

	case DeviceRemove:
		c.executeAction(eventMsg.Payload, c.onDeviceRemove, false)

	case DeviceInterview:
		c.executeAction(eventMsg.Payload, c.onDeviceInterview, false)

	case BridgePermitJoin:
		c.executeAction(eventMsg.Payload, c.onBridgePermitJoin, false)

	case SaveLoggerConfig:
		c.executeAction(eventMsg.Payload, c.onSaveLoggerConfig, true)

	case SaveDashboardGroup:
		c.executeAction(eventMsg.Payload, c.onSaveDashboardGroup, true)
	case RenameDashboardGroup:
		c.executeAction(eventMsg.Payload, c.onRenameDashboardGroup, true)

	case DeleteDashboardGroup:
		c.executeAction(eventMsg.Payload, c.onDeleteDashboardGroup, true)

	case ImportDashboardGroups:
		// will need refactoring
		err := c.onImportDashboardGroups(eventMsg.Payload)
		if err != nil {
			utils.LogErrorf("Failed to import dashboard groups: %v", err)
			c.Broadcast(OperationFailed, err.Error())
			return
		}
		res, err := c.onLoadDashboardGroups()
		if err != nil {
			utils.LogErrorf("Failed to load dashboard groups: %v", err)
			c.Broadcast(OperationFailed, err.Error())

			return
		}
		ds, ok := res.(map[string]*settings.DashboardGroup)
		if !ok {
			utils.LogErrorf("onLoadDashboardGroups: Failed to cast to map[string]*settings.DashboardGroup")
			c.Broadcast(OperationFailed, "Failed to cast to map[string]*settings.DashboardGroup")
			return
		}

		err = c.Broadcast(DashboardGroups, ds)
		if err != nil {
			utils.LogErrorf("Failed to broadcast onLoadDashboardGroups %s", err.Error())
			c.Broadcast(OperationFailed, err.Error())
		}

	case LoadDashboardGroups:
		c.executeActionWithEvent(c.onLoadDashboardGroups, DashboardGroups)

	default:

		utils.LogWarnf("Unknown event type: %s", eventMsg.Type)
		return
	}
}

func (c *eventHubImpl) executeActionWithEvent(action func() (interface{}, error), successEvent string) {

	result, err := action()
	if err != nil {
		err = c.Broadcast(OperationFailed, err.Error())
		if err != nil {
			utils.LogErrorf("Failed to broadcast OperationFailed %s", err.Error())
		}
	} else {
		err = c.Broadcast(successEvent, result)
		if err != nil {
			utils.LogErrorf("Failed to broadcast successEvent %s", err.Error())
		}
	}
}

func (c *eventHubImpl) executePayloadActionWithSuccessfullyEvent(payload interface{}, action func(interface{}) (interface{}, error), successEvent string) {

	if payload == nil {
		c.Broadcast(OperationFailed, "payload is empty")
		return
	}

	result, err := action(payload)
	if err == nil {
		err = c.Broadcast(successEvent, result)
		if err != nil {
			utils.LogErrorf("Failed to broadcast successEvent %s", err.Error())
		}
	}
}

func (c *eventHubImpl) executePayloadActionWithEvent(payload interface{}, action func(interface{}) (interface{}, error), successEvent string) {

	if payload == nil {
		c.Broadcast(OperationFailed, "payload is empty")
		return
	}

	result, err := action(payload)
	if err != nil {
		err = c.Broadcast(OperationFailed, err.Error())
		if err != nil {
			utils.LogErrorf("Failed to broadcast OperationFailed %s", err.Error())
		}
	} else {
		err = c.Broadcast(successEvent, result)
		if err != nil {
			utils.LogErrorf("Failed to broadcast OperationSuccess %s", err.Error())
		}
	}
}

func (c *eventHubImpl) executeAction(payload interface{}, action func(interface{}) error, reportSuccess bool) {
	if payload == nil {
		c.Broadcast(OperationFailed, "payload is empty")
		return
	}

	err := action(payload)
	if err != nil {
		err = c.Broadcast(OperationFailed, err.Error())
		if err != nil {
			utils.LogErrorf("Failed to broadcast OperationFailed %s", err.Error())
		}
	} else if reportSuccess {
		err = c.Broadcast(OperationSuccess, nil)
		if err != nil {
			utils.LogErrorf("Failed to broadcast OperationSuccess %s", err.Error())
		}
	}
}
