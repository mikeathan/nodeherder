package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/lanes"
	mcpserver "node-herder/internal/mcp/server"
	metrics "node-herder/internal/metrics/domain"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/internal/ws"
	automationmodels "node-herder/models/automations"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/settings"
	"node-herder/store"
	"node-herder/utils/storage"
	"strconv"
	"strings"
	"sync"
	"time"

	"node-herder/utils"
)

var (
	ErrorEmptyPayload = fmt.Errorf("empty payload")
)

const (
	// ingressKey is the single lane that receives every message in arrival order.
	ingressKey = "ingress"
	// bridgeKey is the lane shared by all bridge/* topics, preserving their order.
	bridgeKey = "bridge"
	// ingressCapacity bounds messages awaiting routing; routing is a map lookup and
	// a non-blocking submit, so this lane drains far faster than device lanes.
	ingressCapacity = 10 * lanes.DefaultCapacity
	// laneShutdownTimeout bounds draining the lanes when the hub context ends.
	laneShutdownTimeout = 5 * time.Second
)

// laneRunner runs work in per-key FIFO order (implemented by *lanes.Executor).
type laneRunner interface {
	Submit(key string, task func()) error
	Do(ctx context.Context, key string, fn func() error) error
	Shutdown(ctx context.Context) error
}

type HubController struct {
	eventHub                                 ws.EventHub
	mqtt                                     mqtt.MqttClient
	store                                    store.AppStore
	ingress                                  laneRunner
	deviceLanes                              laneRunner
	responseHandlers                         map[string]handler // owned by the ingress lane
	DeviceAvailabilityTimeoutOverrideInHours int
	automationEngine                         automations.Engine
	registrar                                *services.HubRegisterService
	ctx                                      context.Context
	getDeviceProcessor                       func() *services.DeviceProcessor
	automationHandlers                       []automations.AutomationHandler
	mcpServer                                mcpserver.MCPStatusProvider
}
type HubControllerOption func(*HubController)

func WithContext(ctx context.Context) HubControllerOption {
	return func(h *HubController) {
		h.ctx = ctx
	}
}
func WithAutomationHandlers(handlers []automations.AutomationHandler) HubControllerOption {
	return func(h *HubController) {
		h.automationHandlers = handlers
	}
}

func WithMCPServer(mcp mcpserver.MCPStatusProvider) HubControllerOption {
	return func(h *HubController) {
		h.mcpServer = mcp
	}
}

func RegisterHubController(eventHub ws.EventHub, store store.AppStore, mqtt mqtt.MqttClient, options ...HubControllerOption) *HubController {
	h := &HubController{
		eventHub:                                 eventHub,
		store:                                    store,
		mqtt:                                     mqtt,
		responseHandlers:                         map[string]handler{},
		DeviceAvailabilityTimeoutOverrideInHours: 24,
		ctx:                                      context.Background(),
		automationHandlers:                       []automations.AutomationHandler{},
	}

	h.automationHandlers = []automations.AutomationHandler{
		automations.NewAutomationScheduler(
			automations.WithContext(h.ctx),
			automations.WithAutomationsFuncs(),
		),
	}

	for _, option := range options {
		option(h)
	}

	h.registrar = services.NewHubRegisterService(store, eventHub, 3600) // 3600 - is not used!!!!!!!!!!!!!!!!
	h.automationEngine = automations.NewEngine(h.automationHandlers, h.registrar, mqtt)

	// Messages are routed on one ordered ingress lane, then handled on a lane per
	// device topic (all bridge topics share one), so each device's messages are
	// processed one at a time in arrival order while devices run concurrently.
	h.ingress = lanes.New(lanes.WithName("ingress"), lanes.WithCapacity(ingressCapacity))
	h.deviceLanes = lanes.New(lanes.WithName("device lanes"))
	context.AfterFunc(h.ctx, h.shutdownLanes)

	h.registerEventHubEvents()

	h.mqtt.OnMessageHandler(func(id string, payload []byte) {
		if err := h.Ingest(id, payload, "mqtt"); err != nil {
			utils.LogErrorf("mqtt message %s not accepted: %v", id, err)
		}
	})

	//
	var once sync.Once
	var deviceProcessor *services.DeviceProcessor
	h.getDeviceProcessor = func() *services.DeviceProcessor {
		once.Do(func() {
			deviceProcessor = h.createDeviceProcessor()
		})
		return deviceProcessor
	}

	// setup
	err := h.mqtt.Connect()
	if err != nil {
		utils.LogErrorf("Failed to connect to MQTT broker: %v", err)
		return nil
	}
	h.mqtt.Publish("bridge/devices", nil) //zigbee2mqtt/ get devices for setup stuff
	return h
}

func (h *HubController) registerEventHubEvents() {
	appconfig := h.store.AppConfig()

	h.eventHub.OnLoadAppConfig(func() (interface{}, error) {

		cfg, err := appconfig.LoadAppConfig()
		if err != nil {
			return nil, err
		}

		return cfg, nil
	})

	h.eventHub.OnLoadBridgeConfig(func() (interface{}, error) {
		return appconfig.LoadBridgeConfig()
	})

	h.eventHub.OnSaveHistoryConfig(func(p interface{}) error {
		req := &settings.HistoryConfig{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)

		if err != nil {
			return fmt.Errorf("OnSaveAppOnSaveHistoryConfigConfig failed. Invalid payload type : %v ", err.Error())
		}

		_, err = appconfig.SaveHistoryConfig(req)
		return err
	})

	h.eventHub.OnSaveDeviceConfigOverride(func(p interface{}) error {
		req := &settings.DeviceConfig{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)

		if err != nil {
			return fmt.Errorf("OnSaveDeviceConfigOverride failed. Invalid payload type : %v ", err.Error())
		}

		return appconfig.SetDeviceConfigOverrides(req)
	})

	h.eventHub.OnDeleteDeviceConfigOverride((func(p interface{}) error {
		bytes, _ := json.Marshal(p)
		payload := make(map[string]interface{})
		err := json.Unmarshal(bytes, &payload)
		if err != nil {
			return fmt.Errorf("OnDeleteDeviceConfigOverrides failed. Invalid payload type : %v ", err.Error())
		}

		id, ok := payload["id"].(string)
		if !ok {
			return fmt.Errorf("OnDeleteDeviceConfigOverrides failed. Invalid payload type missing group id")
		}

		return appconfig.DeleteDeviceConfigOverrides(id)
	}))

	h.eventHub.OnSaveDeviceConfigDefaults(func(p interface{}) error {
		req := &settings.DeviceConfig{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return fmt.Errorf("OnSaveDeviceConfigDefaults failed. Invalid payload type : %v ", err.Error())
		}

		return appconfig.SetDeviceConfigDefaults(req)
	})

	h.eventHub.OnSaveDashboardGroup(func(payload interface{}) error {
		req := &settings.DashboardGroup{}
		bytes, _ := json.Marshal(payload)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return fmt.Errorf("OnSaveDashboardGroup failed. Invalid payload type : %v ", err.Error())
		}

		// validate request
		for id := range req.DeviceGroup {
			_, err := h.registrar.LookupById(id)
			if err != nil {
				utils.LogErrorf("OnSaveDashboardGroup. Deleting invalid expose id : %v ", err.Error())
				delete(req.DeviceGroup, id)
			}
		}

		return appconfig.SaveDashboardGroup(req)
	})

	h.eventHub.OnRenameDashboardGroup(func(p interface{}) (interface{}, error) {

		req := devices.DashboardGroupRenameRequest{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return nil, fmt.Errorf("OnRenameDashboardGroup failed. Invalid payload type : %v ", err.Error())
		}

		return appconfig.RenameDashboardGroup(req.OldName, req.NewName)
	})

	h.eventHub.OnDeleteDashboardGroup(func(p interface{}) error {
		bytes, _ := json.Marshal(p)
		payload := make(map[string]interface{})
		err := json.Unmarshal(bytes, &payload)
		if err != nil {
			return fmt.Errorf("OnDeleteDashboardGroup failed. Invalid payload type : %v ", err.Error())
		}

		id, ok := payload["groupName"].(string)
		if !ok {
			return fmt.Errorf("OnDeleteDashboardGroup failed. Invalid payload type missing group id")
		}
		return appconfig.DeleteDashboardGroup(id)
	})

	h.eventHub.OnImportDashboardGroups(func(payload interface{}) error {
		req := make(map[string]*settings.DashboardGroup)
		bytes, _ := json.Marshal(payload)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return fmt.Errorf("OnImportDashboardGroups failed. Invalid payload type : %v ", err.Error())
		}

		return appconfig.ImportDashboardGroups(req)
	})

	h.eventHub.OnLoadDashboardGroups(func() (interface{}, error) {
		appConfig, err := appconfig.LoadAppConfig()
		if err != nil {
			return nil, err
		}

		return appConfig.Hub.DashboardGroups, nil
	})

	h.eventHub.OnLoadAutomations(func() interface{} {
		return h.automationEngine.GetAllTriggers()
	})

	h.eventHub.OnLoadDevices(func() interface{} {
		devs, _ := h.store.AllDevices()
		return devs
	})

	h.eventHub.OnLoadMetrics(func(p interface{}) (interface{}, error) {

		req := metrics.LoadDeviceMetricsRequest{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)

		if err != nil {
			utils.LogErrorf("LoadDeviceMetricsRequest failed. Invalid payload type : %v ", err.Error())
			return nil, fmt.Errorf("LoadDeviceMetricsRequest failed. Invalid payload type : %v ", err.Error())
		}

		utils.LogDebugf("OnLoadMetrics: request id=%s from=%d to=%d", req.Id, req.From, req.To)

		device, err := h.registrar.LookupById(req.Id)
		if err != nil {
			utils.LogErrorf("LoadDeviceMetricsRequest failed. Device %s not found : %v", req.Id, err.Error())
			return nil, fmt.Errorf("LoadDeviceMetricsRequest failed. Device %s not found", req.Id)
		}

		utils.LogDebugf("OnLoadMetrics: device found id=%s friendlyName=%s exposeCount=%d", device.Id, device.FriendlyName, len(device.Exposes))

		from := time.Unix(req.From, 0).UTC()
		to := time.Unix(req.To, 0).UTC()
		return h.store.ViewMetrics(device, from, to)
	})

	h.eventHub.OnLoadDeviceList(func(ids []string) interface{} {
		devs, _ := h.store.FindDeviceByIds(ids)
		return devs
	})

	h.eventHub.OnLoadDevice(func(name string) (interface{}, error) {
		return h.registrar.LookupByName(name)
	})

	h.eventHub.OnDeviceSetValue(func(p interface{}) error {
		bytes, _ := json.Marshal(p)
		payload := make(map[string]interface{})
		err := json.Unmarshal(bytes, &payload)

		if err != nil {
			return errors.New("set device value failed. Invalid payload type")
		}

		id := payload["id"].(string)

		device, err := h.registrar.LookupById(id)
		if err != nil {
			return fmt.Errorf("device %s not found", id)
		}

		name := payload["name"].(string)
		value := payload["value"]
		msg := map[string]any{
			name: value,
		}

		json, _ := json.Marshal(msg)
		action := fmt.Sprintf("%s/set", device.FriendlyName)
		h.mqtt.Publish(action, json)

		return nil
	})

	h.eventHub.OnDeviceRename(func(p interface{}) error {
		bytes, _ := json.Marshal(p)
		payload := make(map[string]interface{})
		err := json.Unmarshal(bytes, &payload)

		if err != nil {
			return errors.New("device rename failed. Invalid payload type")
		}

		json, _ := json.Marshal(payload)
		h.mqtt.Publish("bridge/request/device/rename", json)
		return nil
	})

	h.eventHub.OnDeviceRemove(func(p interface{}) error {

		req := devices.DeviceRemoveRequest{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return errors.New("device remove failed. Invalid payload type")
		}

		json, _ := json.Marshal(p)
		h.mqtt.Publish("bridge/request/device/remove", json)
		return nil
	})

	h.eventHub.OnDeviceInterview(func(p interface{}) error {
		bytes, _ := json.Marshal(p)
		payload := make(map[string]interface{})
		err := json.Unmarshal(bytes, &payload)

		if err != nil {
			return errors.New("device interview failed. Invalid payload type")
		}

		json, _ := json.Marshal(payload)
		h.mqtt.Publish("bridge/request/device/interview", json)
		return nil
	})

	h.eventHub.OnBridgePermitJoin(func(p interface{}) error {
		req := settings.BridgeConfig{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)

		if err != nil {
			return errors.New("permit join failed. Invalid payload type")
		}
		if req.TimeExpireAt.Value > 254 {
			return errors.New("permit join failed. Invalid timeout. (Max 254 seconds)")
		}

		f := func(value bool) error {

			err := appconfig.SaveBridgePermitJoin(value)
			if err != nil {
				return err
			}

			// emit bridgeConfig back to clients
			h.eventHub.EmitBridgeConfig()
			return nil
		}

		mqttReq := hub.NewBridgePermitJoinRequest(&req, f)
		h.eventHub.Context().Enqueue(mqttReq)

		json, _ := json.Marshal(mqttReq)
		h.mqtt.Publish("bridge/request/permit_join", json)

		return nil
	})

	h.eventHub.OnDeleteAutomationTrigger(func(p interface{}) (interface{}, error) {

		// todo:see if we can cast p to string and then to bytes
		bytes, _ := json.Marshal(p)
		payload := make(map[string]interface{})
		err := json.Unmarshal(bytes, &payload)

		if err != nil {
			return nil, errors.New("delete automation trigger failed. Invalid payload type")
		}

		automationId := payload["automationId"].(string)
		triggerId, err := strconv.Atoi(fmt.Sprint(payload["triggerId"]))
		if err != nil {
			return nil, fmt.Errorf("delete automation trigger failed. Invalid triggerId type  %s", err.Error())
		}
		err = h.automationEngine.DeleteTrigger(automationId, triggerId)
		if err != nil {
			return nil, err
		}

		return h.automationEngine.Load(automationId)
	})

	h.eventHub.OnSaveAutomation(func(p interface{}) error {

		bytes, _ := json.Marshal(p)
		serializer := automations.NewAutomationSerialiser()
		automation, err := serializer.Unmarshal(bytes)
		if err != nil {
			utils.LogErrorf("Save automation failed. Invalid payload type: %s", err.Error())
			return errors.New("save automation failed. Invalid payload type")
		}

		err = h.automationEngine.Add(automation)
		if err != nil {
			return fmt.Errorf("save automation failed. %s", err.Error())
		}

		// NOTE: Do NOT trigger automation on save. The automation will
		// naturally run on the next genuine device state change from MQTT.
		// Triggering here causes a feedback loop: save → action → MQTT publish
		// → bounce-back → device update → re-trigger → action → …
		return nil
	})

	h.eventHub.OnDeleteAutomation(func(p interface{}) (interface{}, error) {

		// todo:see if we can cast p to string and then to bytes

		bytes, _ := json.Marshal(p)
		payload := make(map[string]interface{})
		err := json.Unmarshal(bytes, &payload)

		if err != nil {
			utils.LogErrorf("Delete automation failed. Invalid payload type")
			return nil, errors.New("delete automation failed. Invalid payload payload type")
		}
		id, ok := payload["id"].(string)
		if !ok {
			utils.LogErrorf("Delete automation failed. Invalid payload type")
			return nil, errors.New("delete automation failed. Invalid payload payload type")
		}

		err = h.automationEngine.Delete(id)
		if err != nil {
			return nil, fmt.Errorf("delete automation failed. %s", err.Error())
		}

		return h.automationEngine.GetAllTriggers(), nil
	})

	h.eventHub.OnSaveLoggerConfig(func(p interface{}) error {
		req := &settings.LoggerConfig{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return errors.New("enable remote logger failed. Invalid payload type")
		}

		_, err = appconfig.SaveLoggerConfig(req)
		if err != nil {
			return err
		}

		utils.SetLogLevel(req.Level)
		utils.LogInfof("Remote logger enabled: %v, level: %s", req.EnableRemoteLogger, req.Level)
		return nil
	})

	h.eventHub.OnSaveAssistantConfig(func(p interface{}) error {
		req := &settings.AssistantConfig{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return errors.New("save assistant config failed. Invalid payload type")
		}

		_, err = appconfig.SaveAssistantConfig(req)
		if err != nil {
			return err
		}

		utils.LogInfof("Assistant config updated: Url=%s", req.Url)
		return nil
	})

	if h.mcpServer == nil {
		return
	}

	// MCP status events
	h.eventHub.OnLoadMCPStatus(func() (interface{}, error) {
		return h.mcpServer.Status(), nil
	})

	h.eventHub.OnRestartMCP(func() (interface{}, error) {
		utils.LogInfo("MCP server restart requested via settings UI")
		if err := h.mcpServer.Restart(); err != nil {
			return nil, err
		}
		return h.mcpServer.Status(), nil
	})

	h.eventHub.OnStopMCP(func() (interface{}, error) {
		utils.LogInfo("MCP server stop requested via settings UI")
		if err := h.mcpServer.Stop(); err != nil {
			return nil, err
		}
		_, err := appconfig.SaveMCPConfig(&settings.MCPConfig{Enabled: false})
		if err != nil {
			return nil, err
		}
		return h.mcpServer.Status(), nil
	})

	h.eventHub.OnStartMCP(func() (interface{}, error) {
		utils.LogInfo("MCP server start requested via settings UI")
		if err := h.mcpServer.Start(); err != nil {
			return nil, err
		}
		_, err := appconfig.SaveMCPConfig(&settings.MCPConfig{Enabled: true})
		if err != nil {
			return nil, err
		}
		return h.mcpServer.Status(), nil
	})

	h.mcpServer.SetOnStatusChange(func() {
		h.eventHub.Broadcast(ws.MCPStatus, h.mcpServer.Status())
	})
}

// we only use that to override the default automation storage, lame but we cant easily refactor as weget alot of cyclic dependencies
func (h *HubController) WithAutomationStorage(storage storage.Storage[automations.Automation]) {
	h.automationEngine.WithStorage(storage)
}

func (c *HubController) Enqueue(id string, payload map[string]interface{}, connType string) error {

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.Ingest(id, bytes, connType)
}

// Ingest is the single entry for MQTT and HTTP messages. It only queues the
// message on the ordered ingress lane, so it never blocks the caller (paho's
// router goroutine for MQTT).
func (h *HubController) Ingest(id string, payload []byte, connType string) error {
	return h.ingress.Submit(ingressKey, func() { h.route(id, payload, connType) })
}

func (m *HubController) TriggerAutomation(device *devices.Device, payload map[string]interface{}) {
	m.automationEngine.HandleDevice(device, payload)
}

// TriggerManual runs a manual trigger on the automation's device lane, ordered
// with that device's messages. Engine errors are returned unchanged; if the
// lane is full, shut down, or does not start the trigger before ctx ends, the
// error wraps ErrLaneBusy and the trigger does not run.
func (m *HubController) TriggerManual(ctx context.Context, automationId string, triggerName string) error {
	key := automationId
	if device, err := m.registrar.LookupById(automationId); err == nil {
		key = device.FriendlyName
	}

	err := m.deviceLanes.Do(ctx, key, func() error {
		return m.automationEngine.HandleManual(automationId, triggerName)
	})
	if errors.Is(err, lanes.ErrLaneFull) || errors.Is(err, lanes.ErrClosed) || errors.Is(err, lanes.ErrNotStarted) {
		return fmt.Errorf("%w: %w", automationmodels.ErrLaneBusy, err)
	}
	return err
}

// route runs on the ingress lane, which owns responseHandlers, and hands the
// message to its device (or bridge) lane.
func (m *HubController) route(id string, payload []byte, connType string) {
	h := m.responseHandler(id)
	if h == nil {
		utils.LogErrorf("no handler for topic %s, message ignored", id)
		return
	}

	key := id
	if strings.HasPrefix(id, "bridge") {
		key = bridgeKey
	}
	err := m.deviceLanes.Submit(key, func() {
		if err := h.ProcessPayload(id, connType, payload); err != nil {
			utils.LogErrorf("Job: %s Error: %s", id, err.Error())
		}
	})
	if err != nil {
		utils.LogErrorf("message %s not processed: %v", id, err)
	}
}

func (m *HubController) shutdownLanes() {
	ctx, cancel := context.WithTimeout(context.Background(), laneShutdownTimeout)
	defer cancel()
	// Ingress first: draining it may still hand messages to device lanes.
	if err := m.ingress.Shutdown(ctx); err != nil {
		utils.LogErrorf("ingress lane shutdown: %v", err)
	}
	if err := m.deviceLanes.Shutdown(ctx); err != nil {
		utils.LogErrorf("device lanes shutdown: %v", err)
	}
}

func (m *HubController) responseHandler(id string) handler {

	if _, ok := m.responseHandlers[id]; !ok {
		if strings.HasPrefix(id, "bridge") {
			switch id {

			case "bridge/response/device/rename":
				var h = newBridgeDeviceRenameResponseHandler(m.eventHub, m.mqtt)
				m.responseHandlers[id] = h
			case "bridge/response/device/remove":
				var h = newBridgeDeviceRemoveResponseHandler(m.registrar, m.eventHub, m.mqtt)
				m.responseHandlers[id] = h
			case "bridge/response/device/interview":
				var h = newBridgeDeviceInterviewResponseHandler(m.eventHub, m.mqtt)
				m.responseHandlers[id] = h
			case "bridge/response/permit_join":
				var h = newBridgePermitJoinResponseHandler(m.eventHub, m.mqtt)
				m.responseHandlers[id] = h
			case "bridge/devices":
				var h = newBridgeConfigurationHandler(m.registrar, m.automationEngine, m.mqtt, m.eventHub)
				m.responseHandlers[id] = h
			case "bridge/logging":
				var h = newBridgeLoggingHandler(m.eventHub)
				m.responseHandlers[id] = h
			}
		} else {

			processor := m.getDeviceProcessor()
			var h = newDeviceHandler(processor)
			m.responseHandlers[id] = h
		}
	}

	return m.responseHandlers[id]
}

// TODO: can be refactored to use a factory. for now we will keep it simple
// will have to create some shared context for hub controller so i can add that thre as well with the others
func (d *HubController) createDeviceProcessor() *services.DeviceProcessor {

	events := devices.NewDeviceRequestEvents()
	events.WithAvailabilityTimeout(time.Duration(d.DeviceAvailabilityTimeoutOverrideInHours) * time.Hour)
	events.WithOnNewDevice(func(device *devices.Device) {
		d.handleDeviceAdded(device)
	})

	events.WithOnDeviceUpdated(func(device *devices.Device, p *devices.UpdatePackage) {
		d.handleDeviceUpdated(device, p)
	})
	events.WithOnDeviceAvailabilityChanged(func(p *devices.UpdatePackage) {
		d.handleDeviceAvailabilityChanged(p)
	})

	events.WithOnDeviceMeasurementsUpdated(func(device *devices.Device, p map[string]interface{}) {
		d.handleDeviceMeasurementsUpdated(device, p)
	})

	events.WithOnDeviceAutomationTriggered(func(device *devices.Device, payload map[string]interface{}) {
		d.automationEngine.HandleDevice(device, payload)
	})

	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(d.registrar).
		WithStore(d.store).
		WithEvents(events).
		WithAutomationQuerier(d.automationEngine).
		Build()

	appconfig := d.store.AppConfig()
	appconfig.RegisterDeviceConfigUpdateListener(func(cfg *settings.DeviceConfig) {
		processor.OnDeviceConfigUpdated(cfg)
	})

	return processor
}

func (d *HubController) handleDeviceAdded(device *devices.Device) error {
	d.eventHub.Broadcast(ws.DeviceAdded, device)

	return d.store.StoreDevice(device.FriendlyName, device)
}

// /
func (d *HubController) handleDeviceUpdated(device *devices.Device, payload *devices.UpdatePackage) error {
	d.eventHub.Broadcast(ws.DeviceUpdated, payload)

	return d.store.StoreDevice(device.FriendlyName, device)
}

func (d *HubController) handleDeviceAvailabilityChanged(p *devices.UpdatePackage) {
	d.eventHub.Broadcast(ws.DeviceUpdated, p)
}

func (d *HubController) handleDeviceMeasurementsUpdated(device *devices.Device, payload map[string]interface{}) error {

	utils.LogDebugf("handleDeviceMeasurementsUpdated: deviceId=%s friendlyName=%s keys=%d", device.Id, device.FriendlyName, len(payload))

	return d.store.StoreMetrics(device.FriendlyName, payload)
}

func convertToMap(payload []byte) (map[string]interface{}, error) {

	if len(payload) == 0 {
		return nil, ErrorEmptyPayload
	}

	deviceMap := make(map[string]interface{})
	err := json.Unmarshal(payload, &deviceMap)
	if err != nil {
		return nil, err
	}

	return deviceMap, nil
}
