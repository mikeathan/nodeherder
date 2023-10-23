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
	"strconv"

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
	registrar                         *services.DeviceRegistrar
}

func RegisterHubController(eventHub ws.EventHub, mqtt mqtt.MqttClient, repo devices.Repository, ctx context.Context) *HubController {

	h := &HubController{
		eventHub:                          eventHub,
		mqtt:                              mqtt,
		repo:                              repo,
		handlers:                          map[string]handler{},
		DeviceAvailabilityTimeoutOverride: 3600,
	}

	h.registrar = services.NewHubRegisterService(repo, eventHub, 3600)
	h.automationEngine = automations.NewEngine(h.registrar, mqtt)
	h.wp = utils.NewWorkerPool(1, ctx)
	h.wp.Run()

	h.eventHub.OnLoadAutomations(func() interface{} {
		return h.automationEngine.GetAllTriggers()
	})

	h.eventHub.OnLoadDevices(func() interface{} {
		return h.repo.ListAllDevicesV2()
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
			fmt.Println(err.Error())
			return nil, errors.New("delete automation trigger failed. Invalid triggerId type")
		}
		err = h.automationEngine.DeleteTrigger(automationId, triggerId)
		if err != nil {
			return nil, err
		}

		return h.automationEngine.Load(automationId)
	})

	h.eventHub.OnSaveAutomation(func(p interface{}) error {

		automation := &automations.Device{}
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

func (m *HubController) TriggerAutomationV2(device *devices.DeviceV2) {
	m.automationEngine.HandleDeviceV2(device)
}

func (m *HubController) ProcessMessage(id string, payload []byte, connType string) error {

	if _, ok := m.handlers[id]; !ok {

		if strings.HasPrefix(id, "bridge") {
			switch id {
			case "bridge/devices":
				var h = newBridgeConfigurationHandler(m.registrar, m.automationEngine, m.mqtt, m.DeviceAvailabilityTimeoutOverride)
				m.handlers[id] = h
			case "bridge/logging":
				var h = newBridgeLoggingHandler(m.eventHub)
				m.handlers[id] = h
			}
		} else {
			var h = newDeviceV2Handler(m.registrar, m.eventHub, m)
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
