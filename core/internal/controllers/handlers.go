package controllers

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/models/logging"
	"node-herder/utils"
	"strings"
)

type mqttResponseTask struct {
	Id      string
	Payload []byte
	Type    string
	h       handler
}

func (m *mqttResponseTask) OnFailure(err error) {
	// TODO: maybe do somethng wit the error
	utils.LogErrorf("Job: %s Error: %s", m.Id, err.Error())
}

func (m *mqttResponseTask) Process() error {
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
	ws               ws.EventHub
	mqtt             mqtt.MqttClient
	registrar        *services.HubRegisterService
	automationEngine automations.Engine
	bridgeHash       *bridgeHash
}

func newBridgeConfigurationHandler(registrar *services.HubRegisterService, engine automations.Engine, mqtt mqtt.MqttClient, ws ws.EventHub) *bridgeConfigurationHandler {
	return &bridgeConfigurationHandler{
		registrar:        registrar,
		automationEngine: engine,
		ws:               ws,
		mqtt:             mqtt,
		bridgeHash:       newBridgeHash()}
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

	b.registrar.RegisterBridge(bridgeInfoList)
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

type bridgeDeviceRemoveResponseHandler struct {
	topic     string
	ws        ws.EventHub
	mqtt      mqtt.MqttClient
	registrar *services.HubRegisterService
}

func newBridgeDeviceRemoveResponseHandler(registrar *services.HubRegisterService, ws ws.EventHub, mqtt mqtt.MqttClient) *bridgeDeviceRemoveResponseHandler {
	return &bridgeDeviceRemoveResponseHandler{topic: "bridge/response/device/remove", registrar: registrar, ws: ws, mqtt: mqtt}
}

func (b *bridgeDeviceRemoveResponseHandler) ProcessPayload(id string, connType string, payload []byte) error {

	utils.LogDebugf("bridge/response/device/remove %s", string(payload))
	if !strings.HasPrefix(id, b.topic) {
		return nil
	}

	resp := new(BridgeResponse)
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
				b.ws.Broadcast(ws.OperationFailed, fmt.Sprintf("Device %s failed to remove from store", deviceId))
				return nil
			}
			b.ws.Broadcast(ws.OperationSuccess, fmt.Sprintf("Device %s removed", deviceId))
		}
	} else {
		b.ws.Broadcast(ws.OperationFailed, resp.Error)
	}

	return nil
}

type bridgeDeviceInterviewResponseHandler struct {
	topic string
	ws    ws.EventHub
	mqtt  mqtt.MqttClient
}

func newBridgeDeviceInterviewResponseHandler(ws ws.EventHub, mqtt mqtt.MqttClient) *bridgeDeviceInterviewResponseHandler {
	return &bridgeDeviceInterviewResponseHandler{topic: "bridge/response/device/interview", ws: ws, mqtt: mqtt}
}

func (b *bridgeDeviceInterviewResponseHandler) ProcessPayload(id string, connType string, payload []byte) error {

	utils.LogDebugf("bridge/response/device/interview: %s", string(payload))
	if !strings.HasPrefix(id, b.topic) {
		return nil
	}

	resp := new(BridgeResponse)
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
		b.ws.Broadcast(ws.OperationFailed, resp.Error)
	}

	return nil
}

type bridgePermitJoinResponseHandler struct {
	topic string
	ws    ws.EventHub
	mqtt  mqtt.MqttClient
}

func newBridgePermitJoinResponseHandler(ws ws.EventHub, mqtt mqtt.MqttClient) *bridgePermitJoinResponseHandler {

	return &bridgePermitJoinResponseHandler{topic: "bridge/response/permit_join", ws: ws, mqtt: mqtt}
}

func (b *bridgePermitJoinResponseHandler) ProcessPayload(id string, connType string, payload []byte) error {
	utils.LogDebugf("bridge/response/permit_join: %s", string(payload))
	if !strings.HasPrefix(id, b.topic) {
		return nil
	}
	resp := new(BridgeResponse)
	resp.Data = map[string]interface{}{}
	err := json.Unmarshal(payload, &resp)

	if err != nil {
		return err
	}

	if resp.Status == "ok" {
		utils.LogInfof("Bridge Permit join set to %v ", resp.Data["time"])
		if resp.Transaction != "" {
			// TODO: if we dont have a request item eg the response came from zigbee2mqtt form their ui
			// then currently we cant update the status. maybe create new request object with state using the resp.Data["time"]

			err := b.ws.Context().Process(resp.Transaction)
			if err != nil {
				utils.LogErrorf("error starting permit join %s", err.Error())
				b.ws.Broadcast(ws.OperationFailed, fmt.Sprintf("error starting permit join %s", err.Error()))
				return err
			}
		}
	} else {
		b.ws.Broadcast(ws.OperationFailed, resp.Error)
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

type BridgeResponse struct {
	Data        map[string]interface{} `json:"data"`
	Status      string                 `json:"status"`
	Error       string                 `json:"error"`
	Transaction string                 `json:"transaction"`
}

func NewBridgeResponse() *BridgeResponse {
	return &BridgeResponse{
		Data:        map[string]interface{}{},
		Status:      "",
		Error:       "",
		Transaction: "",
	}
}

type bridgeLoggingResponse struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func (b *bridgeDeviceRenameResponseHandler) ProcessPayload(id string, connType string, payload []byte) error {

	utils.LogDebugf("bridge/response/device/rename: %s", string(payload))
	if !strings.HasPrefix(id, b.topic) {
		return nil
	}

	resp := new(BridgeResponse)
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

			// emit back to clients updated device name. not tested to see if it works!!!
			b.ws.EmitDevice(resp.Data["to"].(string))
			//b.ws.Broadcast(ws.OperationSuccess, fmt.Sprintf("Device %s renamed to %s", oldName, resp.Data["to"].(string)))
		}
	} else {
		b.ws.Broadcast(ws.OperationFailed, resp.Status)
	}

	return nil
}

type bridgeLoggingHandler struct {
	ws    ws.EventHub
	topic string
}

func newBridgeLoggingHandler(ws ws.EventHub) *bridgeLoggingHandler {
	return &bridgeLoggingHandler{topic: "bridge/logging", ws: ws}
}

func (b *bridgeLoggingHandler) ProcessPayload(id string, connType string, payload []byte) error {
	if !strings.HasPrefix(id, b.topic) {
		return nil
	}

	resp := new(bridgeLoggingResponse)
	err := json.Unmarshal(payload, &resp)
	if err != nil {
		return err
	}

	if resp.Level == logging.LogLevelError {
		b.ws.Broadcast(ws.OperationFailed, resp.Message)
		utils.LogError(resp.Message)
	}

	return nil
}

type deviceHandler struct {
	deviceProcessor *services.DeviceProcessor
}

func newDeviceHandler(deviceProcessor *services.DeviceProcessor) *deviceHandler {

	return &deviceHandler{
		deviceProcessor: deviceProcessor,
	}
}

func (c *deviceHandler) ProcessPayload(friendlyName string, connType string, payload []byte) error {

	dataMap, err := convertToMap(payload)
	if err != nil {
		utils.LogErrorf("error converting payload to map %s", err.Error())
		return nil
	}

	return c.deviceProcessor.CreateOrUpdateDevice(friendlyName, connType, dataMap)
}
