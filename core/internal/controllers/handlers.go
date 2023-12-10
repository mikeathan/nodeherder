package controllers

import (
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
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
	mqtt                      mqtt.MqttClient
	registrar                 *services.DeviceRegistrar
	automationEngine          automations.Engine
	bridgeHash                string
	deviceAvailabilityTimeout int
}

func newBridgeConfigurationHandler(registrar *services.DeviceRegistrar, engine automations.Engine, mqtt mqtt.MqttClient, deviceAvailabilityTimeout int) *bridgeConfigurationHandler {
	return &bridgeConfigurationHandler{registrar: registrar, automationEngine: engine, mqtt: mqtt, deviceAvailabilityTimeout: deviceAvailabilityTimeout}
}

func (b *bridgeConfigurationHandler) ProcessPayload(id string, connType string, payload []byte) error {
	if id != "bridge/devices" {
		return fmt.Errorf("invalid hub configuration topic %s", id)
	}

	h := utils.HashData(payload)
	if b.bridgeHash == h {
		utils.LogDebugf("bridge/devices event. Skipping payload not changed")
		return nil
	}

	b.bridgeHash = h
	bridgeInfoList, err := devices.LoadBridgeDevices(payload)
	if err != nil {
		return err
	}

	b.registrar.RegisterBridge(bridgeInfoList, b.deviceAvailabilityTimeout)
	b.automationEngine.Initialize()

	for _, device := range bridgeInfoList {
		if !device.IsActive() {
			utils.LogInfof("Bridge registration: skipping  %s", device.FriendlyName)

			continue
		}

		// TODO:
		// do we need to unsubsribe from removed/renamed topic
		err := b.mqtt.AddTopic(device.FriendlyName)
		if err != nil {
			utils.LogErrorf("error %s conffgure topic %s", device.FriendlyName, err.Error())
		}
	}
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
		//utils.LogDebugf(string(payload))
	}

	return nil
}

type deviceHandler struct {
	AvailabilityTimeoutInSeconds int
	registrar                    *services.DeviceRegistrar
	eventHub                     ws.EventHub
	hub                          *HubController
}

func newDeviceHandler(registrar *services.DeviceRegistrar, eventHub ws.EventHub, hub *HubController) *deviceHandler {
	return &deviceHandler{
		registrar:                    registrar,
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

	device, _ := c.registrar.LookupByName(friendlyName)
	if device == nil {

		device, err = c.registrar.CreateNewDevice(friendlyName, connType, dataMap)
		if err != nil {
			return err
		}

		c.eventHub.Broadcast(ws.DeviceAdded, device)
	} else {

		updatedData := device.Update(dataMap)
		if !updatedData.HasData() {
			return nil
		}

		// check to see if we have an automation for current device

		c.eventHub.Broadcast(ws.DeviceUpdated, updatedData)
		c.hub.TriggerAutomation(device)
	}
	c.registrar.Register(friendlyName, device)

	return nil
}
