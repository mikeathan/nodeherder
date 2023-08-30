package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		DeviceAvailabilityTimeoutOverride: 3600, // 1 Hour
		automationEngine:                  automations.NewEngine(mqtt),
	}

	h.wp = utils.NewWorkerPool(1, ctx)
	h.wp.Run()

	h.eventHub.OnConnected(func() interface{} {
		return h.repo.ListAllDevices()
	})

	h.mqtt.OnMessageHandler(func(id string, payload []byte) {

		// TODO: we cant do that here as we are blocking mqtt
		// create new hubconfiguration handler
		// which needs to return results

		if !h.configured {
			err := h.configureHub(id, payload)
			if err != nil {
				utils.LogErrorf("configuring hub %s", err.Error())
			}
			return
		}

		h.ProcessMessage(id, payload, "mqtt")
	})

	// setup
	h.mqtt.Connect()
	h.mqtt.Publish("zigbee2mqtt/bridge/devices", nil) // get devices for setup stuff
	return h
}

func (h *HubController) configureHub(id string, payload []byte) error {

	if id != "bridge/devices" {
		return fmt.Errorf("invalid hub configuration topic %s", id)
	}

	devices, err := devices.LoadBridgeDevices(payload)
	if err != nil {
		return err
	}

	h.automationEngine.Load(devices)
	if err != nil {
		return err
	}

	for _, device := range devices {
		if device.Disabled || device.Type == "Coordinator" || !device.InterviewCompleted {
			continue
		}

		err := h.mqtt.AddTopic(device.FriendlyName)
		if err != nil {
			utils.LogErrorf("error %s conffgure topic %s", device.FriendlyName, err.Error())
		}
	}
	h.configured = true

	return nil
}

func (c *HubController) Enqueue(id string, payload map[string]interface{}, connType string) error {

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.ProcessMessage(id, bytes, connType)
}

func (m *HubController) ProcessMessage(id string, payload []byte, connType string) error {

	if _, ok := m.handlers[id]; !ok {

		// TODO add configuration handler
		if strings.HasSuffix(id, "logging") {
			if id == "bridge/logging" {
				var h = newBridgeLoggingHandler(m.eventHub)
				m.handlers[id] = h
			}

		} else {

			var h = newDeviceHandler(m.repo, m.eventHub, m.automationEngine)
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
