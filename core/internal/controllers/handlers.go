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
	"node-herder/store"
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
	AvailabilityTimeoutInSeconds int
	registrar                    *services.HubRegisterService
	eventHub                     ws.EventHub
	deviceServices               map[string]*services.DeviceLifetimeService
	automationEngine             automations.Engine
	store                        store.AppStore
}

func newDeviceHandler(registrar *services.HubRegisterService, eventHub ws.EventHub, store store.AppStore, automationEngine automations.Engine) *deviceHandler {

	return &deviceHandler{
		registrar:                    registrar,
		eventHub:                     eventHub,
		AvailabilityTimeoutInSeconds: 3600, // 1 Hour
		automationEngine:             automationEngine,
		store:                        store,
	}
}

func (c *deviceHandler) ProcessPayload(friendlyName string, connType string, payload []byte) error {

	dataMap, err := convertToMap(payload)
	if err != nil {
		utils.LogErrorf("error converting payload to map %s", err.Error())
		return nil
	}

	p := services.NewDeviceProcessor(c.registrar, nil, c.eventHub, 0)
	pu, err := p.CreateOrUpdateDevice(friendlyName, connType, dataMap)
	newDeviceEvent := func(device *devices.Device, data map[string]interface{}) {
		c.HandleDeviceAdded(device, data)
	}
	//p.OnNewDevice(newDeviceEvent)
	updateDeviceEvent := func(device *devices.Device, data map[string]interface{}) {
		c.HandleDeviceUpdated(device, data)
	}
	//p.OnDeviceUpdated(updateDeviceEvent)
	if err != nil {
		return err
	}

	device, _ := c.registrar.LookupByName(friendlyName)
	if device == nil {
		device, err = c.registrar.CreateNewDevice(friendlyName, connType, dataMap)

		if err != nil {
			return err
		}

		// WIP ##################
		appConfig := c.store.AppConfig()
		debouncer := services.NewDeviceDebouncer(device.Id, appConfig.GetDeviceConfigCache(device.Id), utils.NewRealClock())

		s := services.NewDeviceLifetimeService(device, debouncer)
		s.Monitor(c.AvailabilityTimeoutInSeconds, func(p any) {
			c.eventHub.Broadcast(ws.DeviceUpdated, p)
		})

		c.deviceServices[device.Id] = s
		//

		c.eventHub.Broadcast(ws.DeviceAdded, device)
		c.HandleDeviceAdded(device, dataMap)
	} else {

		updatedData := c.deviceServices[device.Id].Update(dataMap)

		//updatedData := device.Update(dataMap)
		if !updatedData.HasData() {
			return nil
		}

		// check to see if we have an automation for current device
		c.eventHub.Broadcast(ws.DeviceUpdated, updatedData)
		c.HandleDeviceUpdated(device, updatedData.Data)
	}

	return nil
}

func (d *deviceHandler) HandleDeviceAdded(device *devices.Device, data map[string]interface{}) error {
	// todo: execute in worker pool
	// 	action()
	// 	m.wp.AddTask(utils.NewWorkerTask(d.Id, action))

	if err := d.registrar.Register(device.FriendlyName, device); err != nil {
		return err
	}

	return d.registrar.StoreMetrics(device.FriendlyName, data)
}

func (d *deviceHandler) HandleDeviceUpdated(device *devices.Device, data map[string]interface{}) error {

	// todo: execute in worker pool
	// 	action()
	// 	m.wp.AddTask(utils.NewWorkerTask(d.Id, action))

	d.automationEngine.HandleDevice(device)
	if err := d.registrar.Register(device.FriendlyName, device); err != nil {
		return err
	}

	return d.registrar.StoreMetrics(device.FriendlyName, data)
}
