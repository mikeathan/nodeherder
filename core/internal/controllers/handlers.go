package controllers

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils"
	"strings"
)

type messageTask struct {
	Id      string
	Payload []byte
	Type    string
	h       handler
}

func (m *messageTask) OnFailure(err error) {
	// TODO: maybe do somethng wit the error
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
	registrar                 *services.HubRegisterService
	automationEngine          automations.Engine
	bridgeHash                string
	deviceAvailabilityTimeout int
}

func newBridgeConfigurationHandler(registrar *services.HubRegisterService, engine automations.Engine, mqtt mqtt.MqttClient, deviceAvailabilityTimeout int) *bridgeConfigurationHandler {
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

type bridgeDeviceResponseHandler struct {
	topic string "bridge/response/device/rename" // for now we support only rename
	ws    ws.EventHub
	mqtt  mqtt.MqttClient
}

func newbridgeDeviceResponseHandler(ws ws.EventHub, mqtt mqtt.MqttClient) *bridgeDeviceResponseHandler {
	return &bridgeDeviceResponseHandler{ws: ws, mqtt: mqtt}
}

type bridgeResponse struct {
	Data   map[string]interface{} `json:"data"`
	Status string                 `json:"status"`
}

func (b *bridgeDeviceResponseHandler) ProcessPayload(id string, connType string, payload []byte) error {

	if !strings.HasPrefix(id, b.topic) {
		return nil
	}

	resp := new(bridgeResponse)
	resp.Data = map[string]interface{}{}
	err := json.Unmarshal(payload, &resp)
	if err != nil {
		return err
	}

	if resp.Status == "ok" {

		newName := resp.Data["to"].(string)
		oldName := resp.Data["from"].(string)

		err = b.ws.EmitDevice(oldName)
		if err != nil {
			utils.LogErrorf("Rename success, failed to find device %s: %s", newName, err.Error())
			return err
		}

		err = b.mqtt.RemoveTopic(oldName)
		if err != nil {
			return err
		}

	} else {
		b.ws.Broadcast(ws.OperationFailed, resp.Status)
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
		utils.LogDebugf(string(payload))
	}

	return nil
}

type deviceHandler struct {
	AvailabilityTimeoutInSeconds int
	registrar                    *services.HubRegisterService
	eventHub                     ws.EventHub
	hub                          *HubController
}

func newDeviceHandler(registrar *services.HubRegisterService, eventHub ws.EventHub, hub *HubController) *deviceHandler {
	return &deviceHandler{
		registrar:                    registrar,
		eventHub:                     eventHub,
		hub:                          hub,
		AvailabilityTimeoutInSeconds: 3600, // 1 Hour
	}
}

func (c *deviceHandler) ProcessPayload(friendlyName string, connType string, payload []byte) error {

	if len(payload) == 0 {
		return nil
	}
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
