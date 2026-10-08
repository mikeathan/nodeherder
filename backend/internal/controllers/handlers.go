package controllers

import (
	"encoding/json"
	"errors"
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
	if !strings.HasPrefix(id, b.topic) {
		return nil
	}
	resp, err := decodeBridgeResponse[bridgeDeviceResponseData](b.topic, payload)
	if err != nil {
		return err
	}
	if resp.Status != "ok" {
		return broadcastBridgeResult(b.ws, b.topic, ws.OperationFailed, resp.Error)
	}
	if resp.Data.ID == "" {
		return fmt.Errorf("%s: missing device id", b.topic)
	}
	if err := b.registrar.RemoveDevice(resp.Data.ID); err != nil {
		notifyErr := broadcastBridgeResult(b.ws, b.topic, ws.OperationFailed, fmt.Sprintf("Device %s failed to remove from store", resp.Data.ID))
		return errors.Join(fmt.Errorf("%s: remove device: %w", b.topic, err), notifyErr)
	}
	return broadcastBridgeResult(b.ws, b.topic, ws.OperationSuccess, fmt.Sprintf("Device %s removed", resp.Data.ID))
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
	if !strings.HasPrefix(id, b.topic) {
		return nil
	}
	resp, err := decodeBridgeResponse[bridgeDeviceResponseData](b.topic, payload)
	if err != nil {
		return err
	}
	if resp.Status != "ok" {
		return broadcastBridgeResult(b.ws, b.topic, ws.OperationFailed, resp.Error)
	}
	if resp.Data.ID == "" {
		return fmt.Errorf("%s: missing device id", b.topic)
	}
	return broadcastBridgeResult(b.ws, b.topic, ws.OperationSuccess, fmt.Sprintf("Device %s interview successful", resp.Data.ID))
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
	if !strings.HasPrefix(id, b.topic) {
		return nil
	}
	// Permit-join confirmation is correlated by transaction, not response data.
	resp, err := decodeBridgeResponse[json.RawMessage](b.topic, payload)
	if err != nil {
		return err
	}
	if resp.Status != "ok" {
		return broadcastBridgeResult(b.ws, b.topic, ws.OperationFailed, resp.Error)
	}
	if resp.Transaction == "" {
		return nil
	}
	if err := b.ws.Context().Process(resp.Transaction); err != nil {
		notifyErr := broadcastBridgeResult(b.ws, b.topic, ws.OperationFailed, fmt.Sprintf("error starting permit join %s", err.Error()))
		return errors.Join(fmt.Errorf("%s: process transaction: %w", b.topic, err), notifyErr)
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
	if !strings.HasPrefix(id, b.topic) {
		return nil
	}
	resp, err := decodeBridgeResponse[bridgeRenameResponseData](b.topic, payload)
	if err != nil {
		return err
	}
	if resp.Status != "ok" {
		// Preserve the existing rename failure notification payload.
		return broadcastBridgeResult(b.ws, b.topic, ws.OperationFailed, resp.Status)
	}
	if resp.Data.From == "" || resp.Data.To == "" {
		return fmt.Errorf("%s: missing rename from or to", b.topic)
	}
	if err := b.mqtt.RemoveTopic(resp.Data.From); err != nil {
		return fmt.Errorf("%s: remove old topic: %w", b.topic, err)
	}
	if err := b.ws.EmitDevice(resp.Data.To); err != nil {
		return fmt.Errorf("%s: emit renamed device: %w", b.topic, err)
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

	var resp bridgeLoggingResponse
	err := json.Unmarshal(payload, &resp)
	if err != nil {
		return err
	}

	if resp.Level == logging.LogLevelError {
		err := broadcastBridgeResult(b.ws, b.topic, ws.OperationFailed, resp.Message)
		utils.LogError(resp.Message)
		return err
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
		return fmt.Errorf("device payload: %w", err)
	}

	return c.deviceProcessor.CreateOrUpdateDevice(friendlyName, connType, dataMap)
}
