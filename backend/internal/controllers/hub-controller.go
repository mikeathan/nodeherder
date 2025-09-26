package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/store"
	"node-herder/utils/storage"
	"strconv"
	"sync"
	"time"

	"node-herder/utils"
	"strings"
)

var (
	ErrorEmptyPayload = fmt.Errorf("empty payload")
)

type HubController struct {
	eventHub                          ws.EventHub
	mqtt                              mqtt.MqttClient
	store                             store.AppStore
	wp                                *utils.WorkerPool
	responseHandlers                  map[string]handler
	DeviceAvailabilityTimeoutOverride int
	automationEngine                  automations.Engine
	registrar                         *services.HubRegisterService
	ctx                               context.Context
	getDeviceProcessor                func() *services.DeviceProcessor
	automationHandlers                []automations.AutomationHandler
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

func RegisterHubController(eventHub ws.EventHub, store store.AppStore, mqtt mqtt.MqttClient, options ...HubControllerOption) *HubController {
	h := &HubController{
		eventHub:                          eventHub,
		store:                             store,
		mqtt:                              mqtt,
		responseHandlers:                  map[string]handler{},
		DeviceAvailabilityTimeoutOverride: 3600,
		ctx:                               context.Background(),
		automationHandlers:                []automations.AutomationHandler{},
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
	h.wp = utils.NewWorkerPool(4, h.ctx)
	h.wp.Run()

	h.registerEventHubEvents()

	h.mqtt.OnMessageHandler(func(id string, payload []byte) {
		h.processMessage(id, payload, "mqtt")
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
	h.mqtt.Connect()
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
			return nil, fmt.Errorf("LoadDeviceMetricsRequest failed. Invalid payload type : %v ", err.Error())
		}

		device, err := h.registrar.LookupById(req.Id)
		if err != nil {
			return nil, fmt.Errorf("LoadDeviceMetricsRequest failed. Device %s not found", req.Id)
		}

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

		// TODO: move that in automations package
		// pass payload and return model
		automation := automations.NewBaseAutomation()
		bytes, _ := json.Marshal(p)

		err := json.Unmarshal(bytes, &automation)
		if err != nil {
			utils.LogErrorf("Save automation failed. Invalid payload type")
			return errors.New("save automation failed. Invalid payload type")
		}

		err = h.automationEngine.Add(automation)
		if err != nil {
			return fmt.Errorf("save automation failed. %s", err.Error())
		}

		// trigger automation for changes to apply
		if automation.Enabled {
			device, err := h.registrar.LookupById(automation.Id)
			if err == nil {
				utils.LogInfof("Trigger automation %s[%s] after update", automation.FriendlyName, automation.Id)
				h.TriggerAutomation(device)
			}
		}

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

		utils.LogInfof("Remote logger enabled: %v", req.EnableRemoteLogger)
		return nil
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

	return c.processMessage(id, bytes, connType)
}

func (m *HubController) TriggerAutomation(device *devices.Device) {
	m.automationEngine.HandleDevice(device)
}

func (m *HubController) TriggerManual(automationId string, triggerName string) error {
	return m.automationEngine.HandleManual(automationId, triggerName)
}

func (m *HubController) processMessage(id string, payload []byte, connType string) error {

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

	var h handler = m.responseHandlers[id]
	return m.wp.AddTask(&mqttResponseTask{Id: id, Type: connType, Payload: payload, h: h})
}

// TODO: can be refactored to use a factory. for now we will keep it simple
// will have to create some shared context for hub controller so i can add that thre as well with the others
func (d *HubController) createDeviceProcessor() *services.DeviceProcessor {

	events := devices.NewDeviceRequestEvents(d.DeviceAvailabilityTimeoutOverride)
	events.WithOnNewDevice(func(device *devices.Device) {
		d.handleDeviceAdded(device)
	})

	events.WithOnDeviceUpdated(func(device *devices.Device, p *devices.UpdatePackage) {
		d.handleDeviceUpdated(device, p)
	})
	events.WithOnDeviceAvailabilityChanged(func(p *devices.UpdatePackage) {
		d.handleDeviceAvailabilityChanged(p)
	})

	if any change happesn we endup here and we then attempt to trigger automation
	need to think. 

	events.WithOnDeviceMeasurementsUpdated(func(device *devices.Device, p map[string]interface{}) {
		d.handleDeviceMeasurementsUpdated(device, p)
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
	// todo: execute in worker pool
	// 	action()
	// 	m.wp.AddTask(utils.NewWorkerTask(d.Id, action))

	d.eventHub.Broadcast(ws.DeviceAdded, device)

	return d.store.StoreDevice(device.FriendlyName, device)
}

// /
func (d *HubController) handleDeviceUpdated(device *devices.Device, payload *devices.UpdatePackage) error {

	// todo: execute in worker pool
	// 	action()
	// 	m.wp.AddTask(utils.NewWorkerTask(d.Id, action))

	d.eventHub.Broadcast(ws.DeviceUpdated, payload)

	return d.store.StoreDevice(device.FriendlyName, device)
}

func (d *HubController) handleDeviceAvailabilityChanged(p *devices.UpdatePackage) {
	d.eventHub.Broadcast(ws.DeviceUpdated, p)
}

func (d *HubController) handleDeviceMeasurementsUpdated(device *devices.Device, payload map[string]interface{}) error {

	d.automationEngine.HandleDevice(device)

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
