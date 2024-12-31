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

type bridgeHash struct {
	root      string
	deviceMap map[string]string
}

func newBridgeHash() *bridgeHash {
	return &bridgeHash{root: "", deviceMap: map[string]string{}}
}

type bridgeConfigurationHandler struct {
	ws                        ws.EventHub
	mqtt                      mqtt.MqttClient
	registrar                 *services.HubRegisterService
	automationEngine          automations.Engine
	bridgeHash                *bridgeHash
	deviceAvailabilityTimeout int
}

func newBridgeConfigurationHandler(registrar *services.HubRegisterService, engine automations.Engine, mqtt mqtt.MqttClient, ws ws.EventHub, deviceAvailabilityTimeout int) *bridgeConfigurationHandler {
	return &bridgeConfigurationHandler{
		registrar:                 registrar,
		automationEngine:          engine,
		ws:                        ws,
		mqtt:                      mqtt,
		deviceAvailabilityTimeout: deviceAvailabilityTimeout,
		bridgeHash:                newBridgeHash()}
}

func (b *bridgeConfigurationHandler) ProcessPayload(id string, connType string, payload []byte) error {
	if len(payload) == 0 {
		return nil
	}

	if id != "bridge/devices" {
		return fmt.Errorf("invalid hub configuration topic %s", id)
	}

	h := utils.HashData(payload)
	if b.bridgeHash.root == h {
		utils.LogDebugf("bridge/devices event. Skipping payload not changed")
		return nil
	}

	bridgeInfoList, err := devices.LoadBridgeDevices(payload)
	if err != nil {
		return err
	}

	var updatedDeviceMap map[string]string = make(map[string]string)

	b.registrar.RegisterBridge(bridgeInfoList, b.deviceAvailabilityTimeout)
	b.automationEngine.Initialize()

	for _, device := range bridgeInfoList {
		if !device.IsActive() {
			utils.LogInfof("Bridge registration: skipping  %s", device.FriendlyName)

			continue
		}

		// device hashing
		bytes, _ := json.Marshal(device)
		dh := utils.HashData(bytes)
		if dh != b.bridgeHash.deviceMap[device.IeeeAddress] {
			updatedDeviceMap[device.IeeeAddress] = dh
		}

		// TODO:
		// do we need to unsubsribe from removed/renamed topic
		err := b.mqtt.AddTopic(device.FriendlyName)
		if err != nil {
			utils.LogErrorf("error %s conffgure topic %s", device.FriendlyName, err.Error())
		}
	}

	//
	if b.bridgeHash.root != "" && len(updatedDeviceMap) > 0 {
		ids := make([]string, 0, len(updatedDeviceMap))
		for k := range updatedDeviceMap {
			ids = append(ids, k)
		}
		b.ws.EmitDeviceList(ids)
	}

	// update bridge hash
	for id, hash := range updatedDeviceMap {
		b.bridgeHash.deviceMap[id] = hash
	}
	b.bridgeHash.root = h

	return nil
}

type bridgeDeviceRemoveRequestHandler struct {
	topic     string
	ws        ws.EventHub
	mqtt      mqtt.MqttClient
	registrar *services.HubRegisterService
}

func newBridgeDeviceRemoveResponseHandler(registrar *services.HubRegisterService, ws ws.EventHub, mqtt mqtt.MqttClient) *bridgeDeviceRemoveRequestHandler {
	return &bridgeDeviceRemoveRequestHandler{topic: "bridge/request/device/remove", registrar: registrar, ws: ws, mqtt: mqtt}
}

func (b *bridgeDeviceRemoveRequestHandler) ProcessPayload(id string, connType string, payload []byte) error {

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
		if deviceId, ok := resp.Data["id"].(string); ok {
			err = b.registrar.RemoveDevice(deviceId)
			if err != nil {
				utils.LogErrorf("error removing device %s from store = %s", deviceId, err.Error())

				b.ws.Broadcast(ws.OperationFailed, fmt.Sprintf("Device %s removed from bridge but not from store", deviceId))
				return nil
			}
			b.ws.Broadcast(ws.OperationSuccess, fmt.Sprintf("Device %s removed", deviceId))
		}
	} else {
		b.ws.Broadcast(ws.OperationFailed, resp.Error)
	}

	return nil
}

type bridgeDeviceInterviewRequestHandler struct {
	topic string
	ws    ws.EventHub
	mqtt  mqtt.MqttClient
}

func newBridgeDeviceInterviewRequestHandler(ws ws.EventHub, mqtt mqtt.MqttClient) *bridgeDeviceInterviewRequestHandler {
	return &bridgeDeviceInterviewRequestHandler{topic: "bridge/request/device/interview", ws: ws, mqtt: mqtt}
}

func (b *bridgeDeviceInterviewRequestHandler) ProcessPayload(id string, connType string, payload []byte) error {
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
		if deviceId, ok := resp.Data["id"].(string); ok {
			b.ws.Broadcast(ws.OperationSuccess, fmt.Sprintf("Device %s interview successful", deviceId))
		}
	} else {
		b.ws.Broadcast(ws.OperationFailed, resp.Status)
	}

	return nil
}

type bridgeDeviceRenameResponseHandler struct {
	topic string
	ws    ws.EventHub
	mqtt  mqtt.MqttClient
}

func newBridgeDeviceRenameResponseHandler(ws ws.EventHub, mqtt mqtt.MqttClient) *bridgeDeviceRenameResponseHandler {
	return &bridgeDeviceRenameResponseHandler{topic: "bridge/response/device/rename", ws: ws, mqtt: mqtt}
}

type bridgeResponse struct {
	Data   map[string]interface{} `json:"data"`
	Status string                 `json:"status"`
	Error  string                 `json:"error"`
}

func (b *bridgeDeviceRenameResponseHandler) ProcessPayload(id string, connType string, payload []byte) error {

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
		if oldName, ok := resp.Data["from"].(string); ok {
			err = b.mqtt.RemoveTopic(oldName)
			if err != nil {
				return err
			}
			b.ws.Broadcast(ws.OperationSuccess, fmt.Sprintf("Device %s renamed to %s", oldName, resp.Data["to"].(string)))
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
		//utils.LogDebugf(string(payload))
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
		c.hub.deviceAdded(device, dataMap)
	} else {

		updatedData := device.Update(dataMap)
		if !updatedData.HasData() {
			return nil
		}

		// check to see if we have an automation for current device
		c.eventHub.Broadcast(ws.DeviceUpdated, updatedData)
		c.hub.deviceUpdated(device, updatedData.Data)
	}

	return nil
}
