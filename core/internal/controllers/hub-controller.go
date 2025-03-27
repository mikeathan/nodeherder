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
}

func RegisterHubController(eventHub ws.EventHub, store store.AppStore, mqtt mqtt.MqttClient, ctx context.Context) *HubController {
	h := &HubController{
		eventHub:                          eventHub,
		store:                             store,
		mqtt:                              mqtt,
		responseHandlers:                  map[string]handler{},
		DeviceAvailabilityTimeoutOverride: 3600,
		ctx:                               ctx,
	}

	h.registrar = services.NewHubRegisterService(store, eventHub, 3600)

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithContext(ctx),
		automations.WithAutomationsFuncs())

	h.automationEngine = automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, h.registrar, mqtt)
	h.wp = utils.NewWorkerPool(4, ctx)
	h.wp.Run()

	h.registerEventHubEvents()

	h.mqtt.OnMessageHandler(func(id string, payload []byte) {
		h.processMessage(id, payload, "mqtt")
	})

	// setup
	h.mqtt.Connect()
	h.mqtt.Publish("bridge/devices", nil) //zigbee2mqtt/ get devices for setup stuff
	return h
}

func (h *HubController) registerEventHubEvents() {
	appconfig := h.store.AppConfig()

	h.eventHub.OnLoadAppConfig(func() (interface{}, error) {
		return appconfig.LoadAppConfig()
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

	h.eventHub.OnSaveDeviceConfig(func(p interface{}) error {
		req := &settings.DeviceConfig{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)

		if err != nil {
			return fmt.Errorf("OnSaveDeviceConfig failed. Invalid payload type : %v ", err.Error())
		}
		return appconfig.SetDeviceConfig(req)
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
		automation := automations.NewDevice("")
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
func (h *HubController) WithAutomationStorage(storage storage.Storage[automations.Device]) {
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

			processor := m.createDeviceProcessor()
			var h = newDeviceHandler(processor)
			m.responseHandlers[id] = h
		}
	}

	var h handler = m.responseHandlers[id]
	return m.wp.AddTask(&mqttResponseTask{Id: id, Type: connType, Payload: payload, h: h})
}

// TODO: can be refactored to use a factory. for now we will keep it simple
func (d *HubController) createDeviceProcessor() *services.DeviceProcessor {
	events := devices.NewDeviceRequestEvents(d.DeviceAvailabilityTimeoutOverride)
	events.WithOnNewDevice(func(device *devices.Device, data map[string]interface{}) {
		d.handleDeviceAdded(device, data)
	})

	events.WithOnDeviceUpdated(func(device *devices.Device, p *devices.UpdatePackage) {
		d.handleDeviceUpdated(device, p)
	})
	events.WithOnDeviceAvailabilityChanged(func(p *devices.UpdatePackage) {
		d.handleDeviceAvailabilityChanged(p)
	})

	return services.NewDeviceProcessor(d.registrar, d.store, events)
}

func (d *HubController) handleDeviceAdded(device *devices.Device, data map[string]interface{}) error {
	// todo: execute in worker pool
	// 	action()
	// 	m.wp.AddTask(utils.NewWorkerTask(d.Id, action))

	d.eventHub.Broadcast(ws.DeviceAdded, device)

	if err := d.store.StoreDevice(device.FriendlyName, device); err != nil {
		return err
	}

	// TODO: refactor code is repeated
	// do we want to do that only if metrics are enabled ?
	// filter out any non measurement data for storing in metrics
	for k := range data {
		if e, ok := device.Exposes[k]; ok && e.Category != devices.MeasurementCategory {
			delete(data, k)
		}
	}

	return d.store.StoreMetrics(device.FriendlyName, data)
}

func (d *HubController) handleDeviceUpdated(device *devices.Device, p *devices.UpdatePackage) error {

	// todo: execute in worker pool
	// 	action()
	// 	m.wp.AddTask(utils.NewWorkerTask(d.Id, action))
	d.eventHub.Broadcast(ws.DeviceUpdated, p)

	d.automationEngine.HandleDevice(device)
	if err := d.store.StoreDevice(device.FriendlyName, device); err != nil {
		return err
	}

	// TODO: refactor code is repeated
	// do we want to do that only if metrics are enabled ?
	// filter out any non measurement data for storing in metrics
	for k := range p.Data {
		if e, ok := device.Exposes[k]; ok && e.Category != devices.MeasurementCategory {
			delete(p.Data, k)
		}
	}
	return d.store.StoreMetrics(device.FriendlyName, p.Data)
}

func (d *HubController) handleDeviceAvailabilityChanged(p *devices.UpdatePackage) {
	d.eventHub.Broadcast(ws.DeviceUpdated, p)
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
