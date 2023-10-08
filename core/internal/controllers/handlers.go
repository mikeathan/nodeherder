package controllers

import (
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils"
)

type messageTask struct {
	Id      string
	Payload []byte
	Type    string
	h       handler
}

func (m *messageTask) OnFailure(err error) {
	utils.LogErrorf("Job: %s Error: %s", m.Id, err.Error())
}

func (m *messageTask) Process() error {
	return m.h.ProcessPayload(m.Id, m.Type, m.Payload)
}

type handler interface {
	ProcessPayload(id string, connType string, payload []byte) error
}

type bridgeConfigurationHandler struct {
	ws         ws.EventHub
	mqtt       mqtt.MqttClient
	hub        *HubController
	configured bool
}

func newBridgeConfigurationHandler(ws ws.EventHub, mqtt mqtt.MqttClient, hub *HubController) *bridgeConfigurationHandler {
	return &bridgeConfigurationHandler{ws: ws, mqtt: mqtt, hub: hub}
}

func (b *bridgeConfigurationHandler) ProcessPayload(id string, connType string, payload []byte) error {

	if b.configured {
		//todo:
		// we need to reconfigure mqtt topics if friendly name changed
		// automations if names changed - maybe use id instead of name ?
		return fmt.Errorf("hub is already configured ")
	}

	if id != "bridge/devices" {
		return fmt.Errorf("invalid hub configuration topic %s", id)
	}

	devices, err := devices.LoadBridgeDevices(payload)
	if err != nil {
		return err
	}

	b.hub.configureBridge(devices)
	if err != nil {
		utils.LogErrorf("Configure bridge error %s", err.Error())
	}

	for _, device := range devices {
		if device.Disabled || device.Type == "Coordinator" || !device.InterviewCompleted {
			continue
		}

		err := b.mqtt.AddTopic(device.FriendlyName)
		if err != nil {
			utils.LogErrorf("error %s conffgure topic %s", device.FriendlyName, err.Error())
		}
	}
	b.configured = true

	return nil
}

type bridgeLoggingHandler struct {
	ws ws.EventHub
}

func newBridgeLoggingHandler(ws ws.EventHub) *bridgeLoggingHandler {
	return &bridgeLoggingHandler{ws: ws}
}

func (b *bridgeLoggingHandler) ProcessPayload(id string, connType string, payload []byte) error {

	if id == "bridge/logging" {
		// todo: handle
		fmt.Println(string(payload))
	}
	return nil
}

type deviceHandler struct {
	AvailabilityTimeoutInSeconds int
	repo                         devices.Repository
	eventHub                     ws.EventHub
	hub                          *HubController
}

type deviceV2Handler struct {
	AvailabilityTimeoutInSeconds int
	repo                         devices.Repository
	eventHub                     ws.EventHub
	hub                          *HubController
}

func newDeviceHV2andler(repo devices.Repository, eventHub ws.EventHub, hub *HubController) *deviceV2Handler {
	return &deviceV2Handler{
		repo:                         repo,
		eventHub:                     eventHub,
		hub:                          hub,
		AvailabilityTimeoutInSeconds: 3600, // 1 Hour
	}
}

func newDeviceHandler(repo devices.Repository, eventHub ws.EventHub, hub *HubController) *deviceHandler {
	return &deviceHandler{
		repo:                         repo,
		eventHub:                     eventHub,
		hub:                          hub,
		AvailabilityTimeoutInSeconds: 3600, // 1 Hour
	}
}

func (c *deviceHandler) ProcessPayload(friendlyName string, connType string, payload []byte) error {

	dataMap, err := convertToMap(payload)
	if err != nil {
		return err
	}

	device, _ := c.repo.FindDevice(friendlyName)
	if device == nil {
		device, err = devices.CreateNewDevice(friendlyName, connType, dataMap)
		if err != nil {
			return err
		}

		device.StartAvailabilityTimer(c.AvailabilityTimeoutInSeconds, func() {
			c.eventHub.Broadcast(ws.DeviceUpdated, device)
		})
	} else {
		updated := device.TryUpdateDevice(dataMap)
		if !updated {
			return nil
		}
		// check to see if we have an automation for current device
		c.hub.TriggerAutomation(device.Id, device.Sensors)
	}

	c.repo.Store(friendlyName, device)
	c.eventHub.Broadcast(ws.DeviceUpdated, device)

	return nil
}

func newDeviceV2Handler(repo devices.Repository, eventHub ws.EventHub, hub *HubController) *deviceV2Handler {
	return &deviceV2Handler{
		repo:                         repo,
		eventHub:                     eventHub,
		hub:                          hub,
		AvailabilityTimeoutInSeconds: 3600, // 1 Hour
	}
}

func (c *deviceV2Handler) ProcessPayload(friendlyName string, connType string, payload []byte) error {

	dataMap, err := convertToMap(payload)
	if err != nil {
		return err
	}

	device, _ := c.repo.FindDeviceV2(friendlyName)
	if device == nil {

		device, err = devices.CreateNewDeviceV2(c.repo, friendlyName, connType, dataMap)
		if err != nil {
			return err
		}

		device.Monitor(c.AvailabilityTimeoutInSeconds, func(payload map[string]string) {

			c.eventHub.Broadcast(ws.DeviceUpdated, payload)
		})

		c.eventHub.Broadcast(ws.DeviceAdded, device)

	} else {
		// ##################
		if friendlyName != "Human presence" {
			fmt.Println("attempt update", friendlyName)
		}
		// ##################

		updatedData := device.TryUpdate(dataMap)
		if len(updatedData) == 0 {
			return nil
		}

		// check to see if we have an automation for current device
		c.hub.TriggerAutomationV2(device)

		c.eventHub.Broadcast(ws.DeviceUpdated, updatedData)
	}

	c.repo.StoreV2(friendlyName, device)

	return nil
}
