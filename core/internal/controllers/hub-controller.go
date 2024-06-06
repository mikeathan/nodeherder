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
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/store"
	"strconv"
	"time"

	"node-herder/utils"
	"strings"
)

type HubController struct {
	eventHub                          ws.EventHub
	mqtt                              mqtt.MqttClient
	store                             store.AppStore
	wp                                *utils.WorkerPool
	handlers                          map[string]handler
	DeviceAvailabilityTimeoutOverride int
	automationEngine                  automations.Engine
	registrar                         *services.HubRegisterService
}

func RegisterHubController(eventHub ws.EventHub, store store.AppStore, mqtt mqtt.MqttClient, ctx context.Context) *HubController {

	h := &HubController{
		eventHub:                          eventHub,
		store:                             store,
		mqtt:                              mqtt,
		handlers:                          map[string]handler{},
		DeviceAvailabilityTimeoutOverride: 3600,
	}

	h.registrar = services.NewHubRegisterService(store, eventHub, 3600)
	h.automationEngine = automations.NewEngine(h.registrar, mqtt)
	h.wp = utils.NewWorkerPool(4, ctx)
	h.wp.Run()

	h.eventHub.OnLoadAppConfig(func() (interface{}, error) {
		return h.store.LoadAppConfig()
	})

	h.eventHub.OnSaveDeviceConfig(func(p interface{}) error {
		req := &settings.DeviceConfig{}
		bytes, _ := json.Marshal(p)
		err := json.Unmarshal(bytes, &req)

		if err != nil {
			return fmt.Errorf("OnSaveDeviceConfig failed. Invalid payload type : %v ", err.Error())
		}
		return h.store.SaveDeviceConfig(req)
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

		from := time.Unix(req.From, 0)
		to := time.Unix(req.To, 0)
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
			return errors.New("device renamefailed. Invalid payload type")
		}

		json, _ := json.Marshal(payload)
		h.mqtt.Publish("bridge/request/device/rename", json)
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

	h.mqtt.OnMessageHandler(func(id string, payload []byte) {
		h.processMessage(id, payload, "mqtt")
	})

	// setup
	h.mqtt.Connect()
	h.mqtt.Publish("zigbee2mqtt/bridge/devices", nil) // get devices for setup stuff
	return h
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

func (m *HubController) deviceAdded(device *devices.Device) {
	m.registrar.Register(device.FriendlyName, device)
	// TESTING
	// d := device
	// action := func() error {
	// 	return m.registrar.Register(device.FriendlyName, device)
	// }
	// m.wp.AddTask(utils.NewWorkerTask(d.Id, action))
}

func (m *HubController) deviceUpdated(device *devices.Device) {

	m.automationEngine.HandleDevice(device)
	m.registrar.Register(device.FriendlyName, device)
	// // TESTING
	// d := device
	// action := func() error {
	// 	m.automationEngine.HandleDevice(device)
	// 	return m.registrar.Register(device.FriendlyName, device)
	// }
	// m.wp.AddTask(utils.NewWorkerTask(d.Id, action))
}

func (m *HubController) processMessage(id string, payload []byte, connType string) error {

	if _, ok := m.handlers[id]; !ok {
		if strings.HasPrefix(id, "bridge") {
			switch id {

			case "bridge/response/device/rename": // for now we support only rename
				var h = newbridgeDeviceResponseHandler(m.eventHub, m.mqtt)
				m.handlers[id] = h

			case "bridge/devices":
				var h = newBridgeConfigurationHandler(m.registrar, m.automationEngine, m.mqtt, m.eventHub, m.DeviceAvailabilityTimeoutOverride)
				m.handlers[id] = h
			case "bridge/logging":
				var h = newBridgeLoggingHandler(m.eventHub)
				m.handlers[id] = h
			}
		} else {
			var h = newDeviceHandler(m.registrar, m.eventHub, m)
			h.AvailabilityTimeoutInSeconds = m.DeviceAvailabilityTimeoutOverride
			m.handlers[id] = h
		}
	}

	var h handler = m.handlers[id]
	return m.wp.AddTask(&messageTask{Id: id, Type: connType, Payload: payload, h: h})
}

func convertToMap(payload []byte) (map[string]interface{}, error) {

	deviceMap := make(map[string]interface{})
	err := json.Unmarshal(payload, &deviceMap)
	if err != nil {
		return nil, errors.New("invalid device data")
	}
	return deviceMap, nil
}
