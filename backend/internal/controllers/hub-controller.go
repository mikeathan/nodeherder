package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"node-herder/internal/automations"
	"node-herder/internal/lanes"
	mcpserver "node-herder/internal/mcp/server"
	metrics "node-herder/internal/metrics/domain"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/internal/spatial"
	"node-herder/internal/spatialhistory"
	"node-herder/internal/ws"
	automationmodels "node-herder/models/automations"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/settings"
	"node-herder/store"
	"node-herder/utils/storage"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	presenceEngine                           atomic.Pointer[spatial.Engine]
	spatialRecorder                          atomic.Pointer[spatialhistory.Recorder]
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

// WithSpatialRecorder injects the historical recorder owned by the hub composition root.
func WithSpatialRecorder(recorder *spatialhistory.Recorder) HubControllerOption {
	return func(h *HubController) {
		if recorder != nil {
			h.spatialRecorder.Store(recorder)
		}
	}
}

func WithMCPServer(mcp mcpserver.MCPStatusProvider) HubControllerOption {
	return func(h *HubController) {
		h.mcpServer = mcp
	}
}

// SetProviderCommands connects UI and automation commands to the active provider manager.
func (h *HubController) SetProviderCommands(commands services.ProviderCommands) {
	h.registrar.SetProviderCommands(commands)
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

		id, ok := payload["id"].(string)
		if !ok || id == "" {
			return errors.New("device id is required")
		}

		device, err := h.registrar.LookupById(id)
		if err != nil {
			return fmt.Errorf("device %s not found", id)
		}

		name, ok := payload["name"].(string)
		if !ok || name == "" {
			return errors.New("device field is required")
		}
		expose, exists := device.GetExpose(name)
		if !exists || (expose.AccessMode != bridge.WriteAccessMode && expose.AccessMode != bridge.ReadWriteAccessMode) {
			return fmt.Errorf("device field %q is not writable", name)
		}
		value, exists := payload["value"]
		if !exists || (value == nil && device.Provider == "") {
			return errors.New("device value is required")
		}
		// Use the hub policy shared with the browser. A report remains evidence;
		// routine cached configuration echoes no longer force observation broadcasts.
		expose.Data.ExpectReportUntil(time.Now().Add(appconfig.GetDeviceConfigCache().GetCommandConfirmationTimeout(id)))
		if device.Provider != "" {
			ctx, cancel := context.WithTimeout(h.ctx, 10*time.Second)
			defer cancel()
			if err := h.registrar.SendProviderCommand(ctx, id, name, value); err != nil {
				expose.Data.ExpectReportUntil(time.Time{})
				return err
			}
			return nil
		}
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

// RegisterProviderDevice installs a discovered device on the same ordered lane
// used by its automations and later reports. The caller owns the device until
// this call and must not mutate its inventory after registration.
func (h *HubController) RegisterProviderDevice(ctx context.Context, device *devices.Device) error {
	if ctx == nil || device == nil || device.Provider == "" || device.FriendlyName == "" {
		return errors.New("invalid provider device")
	}
	return h.deviceLanes.Do(ctx, device.FriendlyName, func() error { return h.getDeviceProcessor().RegisterProviderDevice(device) })
}

// UpdateProviderDevice applies a confirmed report on its device lane. The map
// is copied before enqueueing; its values must be immutable scalar observations.
func (h *HubController) UpdateProviderDevice(ctx context.Context, id string, payload map[string]interface{}) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	name, ok := h.getDeviceProcessor().DeviceName(id)
	if !ok {
		return errors.New("provider device not registered")
	}
	copy := maps.Clone(payload)
	return h.deviceLanes.Do(ctx, name, func() error { return h.getDeviceProcessor().UpdateProviderDevice(id, copy) })
}

// ProviderDeviceOffline serializes a provider disconnect with reports and marks
// the stored generation unavailable without discarding its last known values.
func (h *HubController) ProviderDeviceOffline(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	name, ok := h.getDeviceProcessor().DeviceName(id)
	if !ok {
		found, err := h.store.FindDeviceByIds([]string{id})
		if err != nil {
			return err
		}
		if len(found) == 0 {
			return nil
		}
		device := found[0]
		if device == nil || device.Provider == "" {
			return errors.New("provider device not registered")
		}
		return h.deviceLanes.Do(ctx, device.FriendlyName, func() error {
			device.SetAvailability(devices.OfflineAvailability)
			if err := h.store.StoreDevice(device.FriendlyName, device); err != nil {
				return err
			}
			update := devices.NewUpdatePackage(id)
			update.Availability = devices.OfflineAvailability
			return h.eventHub.Broadcast(ws.DeviceUpdated, update)
		})
	}
	return h.deviceLanes.Do(ctx, name, func() error { return h.getDeviceProcessor().ProviderOffline(id) })
}

// RemoveProviderDevice retires one provider generation on its device lane and
// broadcasts a fresh list. Missing devices are safe during pre-discovery removal.
func (h *HubController) RemoveProviderDevice(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	if spatial.IsDerivedDeviceID(id) {
		return errors.New("derived presence devices follow their source mapping and cannot be removed directly")
	}
	found, err := h.store.FindDeviceByIds([]string{id})
	if err != nil {
		return err
	}
	if len(found) == 0 {
		return nil
	}
	device := found[0]
	if device == nil || device.Provider == "" {
		return errors.New("provider device not found")
	}
	if err := h.deviceLanes.Do(ctx, device.FriendlyName, func() error { return h.getDeviceProcessor().RemoveDevice(id) }); err != nil {
		return err
	}
	if engine := h.presenceEngine.Load(); engine != nil {
		engine.Remove(id)
	}
	all, err := h.store.AllDevices()
	if err != nil {
		return err
	}
	return h.eventHub.Broadcast(ws.DeviceList, all)
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

// removeDevice is called from the bridge lane. Join the device's own lane before
// retiring it so in-flight state/automation callbacks finish ahead of deletion.
func (m *HubController) removeDevice(id string) error {
	if spatial.IsDerivedDeviceID(id) {
		return errors.New("derived presence devices follow their source mapping and cannot be removed directly")
	}
	processor := m.getDeviceProcessor()
	key := id
	if name, active := processor.DeviceName(id); active {
		key = name
	} else if device, err := m.registrar.LookupById(id); err == nil {
		key = device.GetFriendlyName()
	}
	remove := func() error {
		err := processor.RemoveDevice(id)
		if err == nil {
			if engine := m.presenceEngine.Load(); engine != nil {
				engine.Remove(id)
			}
		}
		return err
	}
	if key == bridgeKey {
		return remove()
	} // Already on this lane; Do would wait on itself.
	ctx, cancel := context.WithTimeout(m.ctx, laneShutdownTimeout)
	defer cancel()
	return m.deviceLanes.Do(ctx, key, remove)
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

// SpatialHistoryService exposes the read side of the recorder; it lazily creates the device
// processor so a fresh boot can serve history once the recorder exists.
func (m *HubController) SpatialHistoryService() *spatialhistory.Service {
	_ = m.getDeviceProcessor()
	if recorder := m.spatialRecorder.Load(); recorder != nil {
		return recorder.Service()
	}
	return nil
}

func (m *HubController) shutdownLanes() {
	if recorder := m.spatialRecorder.Load(); recorder != nil {
		recorder.Close()
	}
	if engine := m.presenceEngine.Load(); engine != nil {
		engine.Close()
	}
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
				var h = newBridgeDeviceRemoveResponseHandler(m.removeDevice, m.eventHub)
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

	var engine *spatial.Engine
	events := devices.NewDeviceRequestEvents()
	events.WithAvailabilityTimeout(time.Duration(d.DeviceAvailabilityTimeoutOverrideInHours) * time.Hour)
	events.WithOnNewDevice(func(device *devices.Device) {
		reportDeviceOutputError(device.Id, d.handleDeviceAdded(device))
		if engine != nil {
			// Register or refresh the derived presence device as soon as its source is known.
			engine.Sync(device.Id)
		}
	})

	events.WithOnDeviceUpdated(func(device *devices.Device, p *devices.UpdatePackage) {
		reportDeviceOutputError(device.Id, d.handleDeviceUpdated(device, p))
	})
	events.OnSpatialTargets = func(frame *devices.SpatialFrame) {
		reportDeviceOutputError(frame.Id, d.eventHub.Broadcast(ws.SpatialTargets, frame))
		if engine != nil {
			engine.Observe(frame)
		}
		if recorder := d.spatialRecorder.Load(); recorder != nil {
			recorder.Observe(frame)
		}
	}
	events.WithOnDeviceAvailabilityChanged(func(p *devices.UpdatePackage) {
		d.handleDeviceAvailabilityChanged(p)
		if engine != nil {
			engine.Sync(p.Id)
		}
	})

	events.WithOnDeviceMeasurementsUpdated(func(device *devices.Device, p map[string]interface{}) {
		reportDeviceOutputError(device.Id, d.handleDeviceMeasurementsUpdated(device, p))
	})

	events.WithOnDeviceAutomationTriggered(func(device *devices.Device, payload map[string]interface{}) {
		d.automationEngine.HandleDevice(device, payload)
	})

	processor := services.NewDeviceProcessorBuilder().
		WithContext(d.ctx).
		WithRegistrar(d.registrar).
		WithStore(d.store).
		WithEvents(events).
		WithAutomationQuerier(d.automationEngine).
		Build()

	appconfig := d.store.AppConfig()
	cache := appconfig.GetDeviceConfigCache()
	engine = spatial.NewEngine(spatial.Deps{
		Context: d.ctx,
		Config:  presenceConfig{cache: cache},
		Source: func(id string) (*devices.Device, bool) {
			device, err := d.registrar.LookupById(id)
			if err != nil || device == nil {
				return nil, false
			}
			return device, true
		},
		Sink:  presenceSink{processor: processor, store: d.store, eventHub: d.eventHub},
		Clock: utils.NewRealClock(),
	})
	engine.Start()
	d.presenceEngine.Store(engine)
	// Reconcile derived devices for already stored devices after a restart, so an offline or
	// quiet source does not keep pre-restart values.
	if stored, err := d.store.AllDevices(); err == nil {
		for _, device := range stored {
			engine.Sync(device.Id)
		}
	}

	appconfig.RegisterDeviceConfigUpdateListener(func(cfg *settings.DeviceConfig) {
		processor.OnDeviceConfigUpdated(cfg)
		if cfg.Id != "" {
			// A saved mapping or zone change must reconcile its derived device immediately.
			engine.Sync(cfg.Id)
		}
	})

	return processor
}

// presenceSink routes derived presence reports through the normal device processor, so they
// reach automations, history and WebSocket clients exactly like provider reports.
type presenceSink struct {
	processor *services.DeviceProcessor
	store     store.AppStore
	eventHub  ws.EventHub
}

func (s presenceSink) Register(device *devices.Device) error {
	return s.processor.RegisterProviderDevice(device)
}

func (s presenceSink) Update(id string, payload map[string]interface{}) error {
	return s.processor.UpdateProviderDevice(id, payload)
}

func (s presenceSink) Offline(id string) error { return s.processor.ProviderOffline(id) }

// Remove retires the derived device and broadcasts the new device list, so an open browser
// drops it instead of keeping stale presence rows after a mapping is cleared.
func (s presenceSink) Remove(id string) error {
	if err := s.processor.RemoveDevice(id); err != nil {
		return err
	}
	all, err := s.store.AllDevices()
	if err != nil {
		return err
	}
	return s.eventHub.Broadcast(ws.DeviceList, all)
}

// presenceConfig adapts saved device settings for the presence engine. A disabled device is
// treated as unmapped, which retires its derived device instead of freezing stale values.
type presenceConfig struct{ cache *settings.DeviceConfigCache }

func (c presenceConfig) Mapped(id string) bool {
	if c.cache.IsDeviceDisabled(id) {
		return false
	}
	return services.HasCompleteSpatialMapping(c.cache.SpatialMappings(id))
}

func (c presenceConfig) Zones(id string) []settings.SpatialZone { return c.cache.SpatialZones(id) }

func (d *HubController) handleDeviceAdded(device *devices.Device) error {
	return d.broadcastAndStoreDevice(ws.DeviceAdded, device, device)
}

func (d *HubController) handleDeviceUpdated(device *devices.Device, payload *devices.UpdatePackage) error {
	return d.broadcastAndStoreDevice(ws.DeviceUpdated, payload, device)
}

// Keep both independent effects in their existing order, even if either fails.
func (d *HubController) broadcastAndStoreDevice(event string, payload any, device *devices.Device) error {
	broadcastErr := d.eventHub.Broadcast(event, payload)
	if broadcastErr != nil {
		broadcastErr = fmt.Errorf("broadcast %s: %w", event, broadcastErr)
	}
	storeErr := d.store.StoreDevice(device.FriendlyName, device)
	if storeErr != nil {
		storeErr = fmt.Errorf("store device: %w", storeErr)
	}
	return errors.Join(broadcastErr, storeErr)
}

func reportDeviceOutputError(id string, err error) {
	if err != nil {
		utils.LogErrorf("device %s output failed: %v", id, err)
	}
}

func (d *HubController) handleDeviceAvailabilityChanged(p *devices.UpdatePackage) {
	if err := d.eventHub.Broadcast(ws.DeviceUpdated, p); err != nil {
		reportDeviceOutputError(p.Id, fmt.Errorf("broadcast availability: %w", err))
	}
}

func (d *HubController) handleDeviceMeasurementsUpdated(device *devices.Device, payload map[string]interface{}) error {

	utils.LogTracef("handleDeviceMeasurementsUpdated: deviceId=%s friendlyName=%s keys=%d", device.Id, device.FriendlyName, len(payload))

	if err := d.store.StoreMetrics(device.FriendlyName, payload); err != nil {
		return fmt.Errorf("store metrics: %w", err)
	}
	return nil
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
	if deviceMap == nil {
		return nil, fmt.Errorf("expected device JSON object")
	}

	return deviceMap, nil
}
