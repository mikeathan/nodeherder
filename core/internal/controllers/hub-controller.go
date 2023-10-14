package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils"
	"strings"
)

type HubController struct {
	eventHub                          ws.EventHub
	mqtt                              mqtt.MqttClient
	repo                              devices.Repository
	wp                                *utils.WorkerPool
	handlers                          map[string]handler
	DeviceAvailabilityTimeoutOverride int
	automationEngine                  automations.Engine
	configured                        bool
}

func RegisterHubController(ws ws.EventHub, mqtt mqtt.MqttClient, repo devices.Repository, ctx context.Context) *HubController {
	h := &HubController{
		eventHub:                          ws,
		mqtt:                              mqtt,
		repo:                              repo,
		handlers:                          map[string]handler{},
		DeviceAvailabilityTimeoutOverride: 3600,
		automationEngine:                  automations.NewEngine(mqtt, repo),
		configured:                        false,
	}

	h.wp = utils.NewWorkerPool(1, ctx)
	h.wp.Run()

	h.eventHub.OnLoadAutomations(func() []byte {

		return h.automationEngine.GetAllTriggers()
	})

	h.eventHub.OnLoadDevices(func() []byte {

		devices := h.repo.ListAllDevicesV2()

		// todo: move it to function instead of here
		bytes, err := json.Marshal(devices)
		if err != nil {
			utils.LogErrorf("marshal devices failed. error %s", err.Error())
		}
		return bytes
	})

	h.eventHub.OnLoadBridgeFeatures(func() []byte {
		// todo: move it to function instead of here
		features := h.repo.GetBridgeFeatures()
		bytes, err := json.Marshal(features)
		if err != nil {
			utils.LogErrorf("marshal bridge features failed. error %s", err.Error())
		}
		return bytes
	})

	h.eventHub.OnSaveAutomation(func(p interface{}) {
		automation, ok := p.(*automations.Device)
		if !ok {
			utils.LogErrorf("save automation failed. error invalid type")
			// todo: error handling. message back error message
			//h.eventHub.Broadcast(error message)
			return
		}
		h.automationEngine.Add(automation)
	})

	h.eventHub.OnDeleteAutomation(func(p interface{}) {
		id, ok := p.(string)
		if !ok {
			utils.LogErrorf("delete automation failed. error invalid type")
			// todo: error handling. message back error message
			//h.eventHub.Broadcast(error message)
			return
		}
		h.automationEngine.Delete(id)
	})

	h.eventHub.OnConnected(func() interface{} {
		return h.repo.ListAllDevicesV2()
	})

	h.mqtt.OnMessageHandler(func(id string, payload []byte) {
		h.ProcessMessage(id, payload, "mqtt")
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

	return c.ProcessMessage(id, bytes, connType)
}

func (m *HubController) configureBridge(bridgeInfoList []*devices.BridgeInfo) {

	m.repo.RegisterBridge(bridgeInfoList)
	m.automationEngine.Initialize()
}

func (m *HubController) TriggerAutomation(id string, data map[string]any) {
	m.automationEngine.HandleDevice(id, data)
}

func (m *HubController) TriggerAutomationV2(device *devices.DeviceV2) {
	m.automationEngine.HandleDeviceV2(device)
}

func (m *HubController) ProcessMessage(id string, payload []byte, connType string) error {

	if _, ok := m.handlers[id]; !ok {

		if strings.HasPrefix(id, "bridge") {
			switch id {
			case "bridge/devices":
				var h = newBridgeConfigurationHandler(m.eventHub, m.mqtt, m)
				m.handlers[id] = h
			case "bridge/logging":
				var h = newBridgeLoggingHandler(m.eventHub)
				m.handlers[id] = h
			}
		} else {
			var h = newDeviceV2Handler(m.repo, m.eventHub, m)
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
