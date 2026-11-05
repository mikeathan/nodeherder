package ws

// This file registers the hub WebSocket protocol, dispatches browser actions and
// broadcasts authoritative updates. Correlated settings saves are acknowledged
// separately so a browser can distinguish submission from persistence.

import (
	"encoding/json"
	"net/http"
	"node-herder/models/hub"
	"node-herder/models/settings"
	"node-herder/utils"
	"strings"

	"github.com/google/uuid"
)

const (
	settingsRequestIDKey = "requestId"
	settingsDeviceIDKey  = "id"
	maxSettingsDeviceID  = 256

	settingsSaveFailure     = "Device settings could not be saved. Previous settings remain active."
	invalidSettingsRequest  = "Invalid settings save request identifier."
	invalidSettingsDeviceID = "Invalid settings device identifier."
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
	SaveAssistantConfig        = "saveAssistantConfig"
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

	LoadMCPStatus = "loadMCPStatus"
	RestartMCP    = "restartMCP"
	StopMCP       = "stopMCP"
	StartMCP      = "startMCP"

	// response
	Automations   = "automations"
	DeviceList    = "deviceList"
	Device        = "device"
	DeviceAdded   = "deviceAdded"
	DeviceUpdated = "deviceUpdated" // returns back updated properties of type DeviceUpdated
	// SpatialTargets carries a complete, synchronized snapshot of a device's mapped target
	// slots (devices.SpatialFrame). It is additive: consumers fall back to DeviceUpdated.
	SpatialTargets   = "spatialTargets"
	OperationFailed  = "operationFailed"
	OperationSuccess = "operationSuccess"
	// DeviceConfigSaveResult carries correlated persistence results for opt-in settings saves.
	DeviceConfigSaveResult = "deviceConfigSaveResult"
	AutomationUpdated      = "automationUpdated" // returns back upated automation

	Metrics         = "metrics"
	AppConfig       = "appConfig"
	BridgeConfig    = "bridgeConfig"
	DashboardGroups = "dashboardGroups"
	MCPStatus       = "mcpStatus"
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
	OnSaveAssistantConfig(func(payload interface{}) error)
	OnSaveDashboardGroup(action func(payload interface{}) error)
	OnRenameDashboardGroup(action func(payload interface{}) (interface{}, error))
	OnImportDashboardGroups(action func(payload interface{}) error)
	OnLoadDashboardGroups(action func() (interface{}, error))
	OnDeleteDashboardGroup(action func(payload interface{}) error)
	OnLoadMCPStatus(action func() (interface{}, error))
	OnRestartMCP(action func() (interface{}, error))
	OnStopMCP(action func() (interface{}, error))
	OnStartMCP(action func() (interface{}, error))
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
	onSaveAssistantConfig        func(interface{}) error
	onSaveDashboardGroup         func(interface{}) error
	onRenameDashboardGroup       func(interface{}) (interface{}, error)
	onDeleteDashboardGroup       func(payload interface{}) error
	onImportDashboardGroups      func(payload interface{}) error
	onLoadDashboardGroups        func() (interface{}, error)
	onLoadMCPStatus              func() (interface{}, error)
	onRestartMCP                 func() (interface{}, error)
	onStopMCP                    func() (interface{}, error)
	onStartMCP                   func() (interface{}, error)
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
		onSaveAssistantConfig:        func(payload interface{}) error { return nil },
		onSaveDashboardGroup:         func(payload interface{}) error { return nil },
		onDeleteDashboardGroup:       func(payload interface{}) error { return nil },
		onRenameDashboardGroup:       func(payload interface{}) (interface{}, error) { return nil, nil },
		onImportDashboardGroups:      func(payload interface{}) error { return nil },
		onLoadDashboardGroups:        func() (interface{}, error) { return nil, nil },
		onLoadMCPStatus:              func() (interface{}, error) { return nil, nil },
		onRestartMCP:                 func() (interface{}, error) { return nil, nil },
		onStopMCP:                    func() (interface{}, error) { return nil, nil },
		onStartMCP:                   func() (interface{}, error) { return nil, nil },
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

func (h *eventHubImpl) OnSaveAssistantConfig(action func(payload interface{}) error) {
	h.onSaveAssistantConfig = action
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

func (h *eventHubImpl) OnLoadMCPStatus(action func() (interface{}, error)) {
	h.onLoadMCPStatus = action
}

func (h *eventHubImpl) OnRestartMCP(action func() (interface{}, error)) {
	h.onRestartMCP = action
}

func (h *eventHubImpl) OnStopMCP(action func() (interface{}, error)) {
	h.onStopMCP = action
}

func (h *eventHubImpl) OnStartMCP(action func() (interface{}, error)) {
	h.onStartMCP = action
}

func (h *eventHubImpl) EmitDevice(name string) error {
	msg, err := h.onLoadDevice(name)
	if err != nil {
		return err
	}
	return h.Broadcast(Device, msg)
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
		c.execute(eventExecutorOptionsWithResultNoError(eventMsg.Payload, wrapPayloadWithResult(c.onLoadMetrics), Metrics))

	case LoadAppconfig:
		c.execute(eventExecutorOptionsWithResult(nil, wrapNoPayload(c.onLoadAppConfig), AppConfig))

	case SaveDeviceConfigOverride:
		c.saveDeviceConfigOverride(eventMsg.Payload)

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

	case SaveAssistantConfig:
		c.execute(eventExecutorOptionsWithSuccess(eventMsg.Payload, wrapPayloadNoResult(c.onSaveAssistantConfig)))

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

	case LoadMCPStatus:
		c.execute(eventExecutorOptionsWithResult(nil, wrapNoPayload(c.onLoadMCPStatus), MCPStatus))

	case RestartMCP:
		c.execute(eventExecutorOptionsWithResult(nil, wrapNoPayload(c.onRestartMCP), MCPStatus))

	case StopMCP:
		c.execute(eventExecutorOptionsWithResult(nil, wrapNoPayload(c.onStopMCP), MCPStatus))

	case StartMCP:
		c.execute(eventExecutorOptionsWithResult(nil, wrapNoPayload(c.onStartMCP), MCPStatus))

	default:

		utils.LogWarnf("Unknown event type: %s", eventMsg.Type)
		return
	}
}

func (c *eventHubImpl) execute(opts *eventExecutorOptions) {

	if opts.Action == nil {
		return
	}

	// TODO:
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

// validSettingsRequestID accepts a canonical UUID correlation token. Tokens
// identify one pending browser request; they never authorize the underlying save.
func validSettingsRequestID(id string) bool {
	parsed, err := uuid.Parse(id)
	return err == nil && parsed.String() == id
}

type deviceConfigSaveResult struct {
	RequestID string `json:"requestId"`
	DeviceID  string `json:"deviceId"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

func (c *eventHubImpl) saveDeviceConfigOverride(payload interface{}) {
	fields, _ := payload.(map[string]interface{})
	token, correlated := fields[settingsRequestIDKey]
	if !correlated {
		c.execute(eventExecutorOptionsWithSuccess(payload, wrapPayloadNoResult(c.onSaveDeviceConfigOverride)))
		return
	}
	requestID, ok := token.(string)
	if !ok || !validSettingsRequestID(requestID) {
		if err := c.Broadcast(OperationFailed, invalidSettingsRequest); err != nil {
			utils.LogErrorf("Failed to report invalid settings request: %v", err)
		}
		return
	}
	deviceID, _ := fields[settingsDeviceIDKey].(string)
	if len(deviceID) > maxSettingsDeviceID || strings.TrimSpace(deviceID) == "" || strings.ContainsRune(deviceID, '\x00') {
		if err := c.Broadcast(OperationFailed, invalidSettingsDeviceID); err != nil {
			utils.LogErrorf("Failed to report invalid settings device: %v", err)
		}
		return
	}
	result := deviceConfigSaveResult{RequestID: requestID, DeviceID: deviceID}
	// Transport correlation metadata must not become part of the configuration callback.
	config := make(map[string]interface{}, len(fields)-1)
	for key, value := range fields {
		if key != settingsRequestIDKey {
			config[key] = value
		}
	}
	if c.onSaveDeviceConfigOverride == nil {
		result.Error = settingsSaveFailure
	} else if err := c.onSaveDeviceConfigOverride(config); err != nil {
		// Return bounded recovery guidance, not repository paths or storage internals.
		result.Error = settingsSaveFailure
	} else {
		result.Success = true
	}
	if err := c.Broadcast(DeviceConfigSaveResult, result); err != nil {
		utils.LogErrorf("Failed to report settings save result: %v", err)
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
func eventExecutorOptionsWithResultNoError(payload interface{}, event eventAction, successEvent string) *eventExecutorOptions {
	return &eventExecutorOptions{
		Action:        event,
		Payload:       payload,
		SuccessEvent:  successEvent,
		ReportResult:  true,
		ReportSuccess: true,
		ReportError:   false,
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
