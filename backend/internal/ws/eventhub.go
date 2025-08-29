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
	LoadDevice      = "loadDevice"
	LoadDeviceList  = "loadDeviceList"

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
	OnRenameDashboardGroup(action func(payload interface{}) (interface{}, error))
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
	onRenameDashboardGroup       func(interface{}) (interface{}, error)
	onDeleteDashboardGroup       func(payload interface{}) error
	onImportDashboardGroups      func(payload interface{}) error
	onLoadDashboardGroups        func() (interface{}, error)
	requestContext               hub.Context
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
		onRenameDashboardGroup:       func(payload interface{}) (interface{}, error) { return nil, nil },
		onImportDashboardGroups:      func(payload interface{}) error { return nil },
		onLoadDashboardGroups:        func() (interface{}, error) { return nil, nil },
		requestContext:               NewRequestContext(),
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

func (h *eventHubImpl) OnRenameDashboardGroup(action func(payload interface{}) (interface{}, error)) {
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
		// TODO: dont report error here. needs refactoring
		c.execute(eventExecutorOptionsWithResult(eventMsg.Payload, wrapPayloadWithResult(c.onLoadMetrics), Metrics))

	case LoadAppconfig:
		c.execute(eventExecutorOptionsWithResult(nil, wrapNoPayload(c.onLoadAppConfig), AppConfig))

	case SaveDeviceConfigOverride:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onSaveDeviceConfigOverride)))

	case DeleteDeviceConfigOverride:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onDeleteDeviceConfigOverride)))

	case SaveDeviceConfigDefaults:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onSaveDeviceConfigDefaults)))

	case SaveHistoryConfig:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onSaveHistoryConfig)))

	case SaveAutomation:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onSaveAutomation)))

	case DeleteAutomation:
		c.execute(eventExecutorOptionsWithResult(eventMsg.Payload, wrapPayloadWithResult(c.onDeleteAutomation), Automations))

	case DeleteAutomationTrigger:
		c.execute(eventExecutorOptionsWithResult(eventMsg.Payload, wrapPayloadWithResult(c.onDeleteAutomationTrigger), AutomationUpdated))

	case DeviceSetValue:
		c.execute(eventExecutorOptionsWithoutSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onDeviceSetValue)))

	case DeviceRename:
		c.execute(eventExecutorOptionsWithoutSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onDeviceRename)))

	case DeviceRemove:
		c.execute(eventExecutorOptionsWithoutSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onDeviceRemove)))

	case DeviceInterview:
		c.execute(eventExecutorOptionsWithoutSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onDeviceInterview)))

	case BridgePermitJoin:
		c.execute(eventExecutorOptionsWithoutSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onBridgePermitJoin)))

	case SaveLoggerConfig:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onSaveLoggerConfig)))

	case SaveDashboardGroup:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onSaveDashboardGroup)))

	case RenameDashboardGroup:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadWithResult(c.onRenameDashboardGroup)))

	case DeleteDashboardGroup:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onDeleteDashboardGroup)))

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
		c.execute(eventExecutorOptionsWithResult(nil, wrapNoPayload(c.onLoadDashboardGroups), DashboardGroups))

	default:

		utils.LogWarnf("Unknown event type: %s", eventMsg.Type)
		return
	}
}

func (c *eventHubImpl) execute(opts *eventExecutorOptions) {

	if opts.Action == nil {
		return
	}

	//if we expect payload and is empty the
	// if payload == nil {
	// 	c.Broadcast(OperationFailed, "payload is empty")
	// 	return
	// }

	result, err := opts.Action(opts.Payload)
	if err != nil {
		if opts.ReportError {
			err = c.Broadcast(OperationFailed, err.Error())
			if err != nil {
				utils.LogErrorf("Failed to broadcast OperationFailed %s", err.Error())
			}
		}
		return
	}

	if !opts.ReportSuccess {
		return
	}

	var reportResult any = nil
	if opts.ReportResult {
		reportResult = result
	}

	var successEvent = opts.SuccessEvent
	if successEvent == "" {
		successEvent = OperationSuccess
	}
	err = c.Broadcast(successEvent, reportResult)
	if err != nil {
		utils.LogErrorf("Failed to broadcast successEvent %s. Error: %s", successEvent, err.Error())
	}
}

// Event Executors

type eventAction func(payload interface{}) (interface{}, error)
type eventExecutorOptions struct {
	Payload       interface{}
	Action        eventAction
	SuccessEvent  string
	ReportResult  bool
	ReportSuccess bool
	ReportError   bool
}

func eventExecutorOptionsWithSuccess(payload interface{}, event eventAction) *eventExecutorOptions {
	return &eventExecutorOptions{
		Action:        event,
		Payload:       payload,
		SuccessEvent:  OperationSuccess,
		ReportResult:  false,
		ReportSuccess: true,
		ReportError:   true,
	}
}

func eventExecutorOptionsWithResult(payload interface{}, event eventAction, successEvent string) *eventExecutorOptions {
	return &eventExecutorOptions{
		Action:        event,
		Payload:       payload,
		SuccessEvent:  successEvent,
		ReportResult:  true,
		ReportSuccess: true,
		ReportError:   true,
	}
}

func eventExecutorOptionsWithoutSuccess(payload interface{}, event eventAction) *eventExecutorOptions {
	return &eventExecutorOptions{
		Action:        event,
		Payload:       payload,
		ReportResult:  false,
		ReportSuccess: false,
		ReportError:   true,
	}
}

func wrapNoPayload(action func() (interface{}, error)) eventAction {
	return func(_ interface{}) (interface{}, error) {
		return action()
	}
}

func wrapPayloadWithResult(action func(interface{}) (interface{}, error)) eventAction {
	return func(payload interface{}) (interface{}, error) {
		return action(payload)
	}
}

func wrapPayloadNoResult(action func(interface{}) error) eventAction {
	return func(payload interface{}) (interface{}, error) {
		return nil, action(payload)
	}
}
