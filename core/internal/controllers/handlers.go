package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/models/automations"
	"node-herder/models/devices"
	"node-herder/utils/pool"
	"strings"
)

type messageTask struct {
	Id      string
	Payload []byte
	Type    string
	h       handler
}

func (m *messageTask) OnFailure(err error) {
	fmt.Printf("Job: %s Error: %s", m.Id, err.Error())
}

func (m *messageTask) Process() error {
	return m.h.ProcessPayload(m.Id, m.Type, m.Payload)
}

type handler interface {
	ProcessPayload(id string, connType string, payload []byte) error
}
type bridgeHandler struct {
	mqtt mqtt.MqttClient
}

func newBridgeHandler(mqtt mqtt.MqttClient) *bridgeHandler {
	return &bridgeHandler{
		mqtt: mqtt, //?????
	}
}

func (b *bridgeHandler) ProcessPayload(id string, connType string, payload []byte) error {

	if id == "bridge/devices" {

		devices, err := devices.LoadBridgeDevices(payload)
		if err != nil {
			fmt.Println("parsing devices error: ", err.Error())
			return err
		}
		if !automations.IsConfigured() {
			automations.Load(devices, b.mqtt)
			if err != nil {
				fmt.Println("loading automations error: ", err.Error())
				return err
			}
		}
	} else if id == "bridge/logging" {
		// todo: handle
		fmt.Println(string(payload))
	}
	return nil
}

type deviceHandler struct {
	AvailabilityTimeoutinSeconds int
	repo                         devices.Repository
	eventHub                     ws.EventHub
}

func newDeviceHandler(repo devices.Repository, eventHub ws.EventHub) *deviceHandler {
	return &deviceHandler{
		repo:                         repo,
		eventHub:                     eventHub,
		AvailabilityTimeoutinSeconds: 3600, // 1 Hour
	}
}

func (c *deviceHandler) ProcessPayload(id string, connType string, payload []byte) error {

	// TODO:
	// Hanlde API payload
	// if connType == "http" do differnt parsing
	//
	dataMap, err := convertToMap(payload)
	if err != nil {
		fmt.Println("error: failed to convert mqtt payload to map")
		return err
	}

	device, _ := c.repo.FindDevice(id)
	if device == nil {
		device, err = devices.CreateNewDevice(id, connType, dataMap)
		if err != nil {
			return err
		}
		device.StartAvailabilityTimer(c.AvailabilityTimeoutinSeconds)
	} else {
		if !device.TryUpdateDevice(dataMap) {
			return nil
		}
	}

	c.repo.Store(id, device)
	c.eventHub.Broadcast(ws.DeviceUpdated, device) // ????
	return nil
}

type messageHandler struct {
	mqtt     mqtt.MqttClient
	wp       *pool.WorkerPool
	ctx      context.Context
	repo     devices.Repository
	eventHub ws.EventHub
	handlers map[string]handler
}

func newMessageHandler(repo devices.Repository, mqtt mqtt.MqttClient, eventHub ws.EventHub, ctx context.Context) *messageHandler {
	h := &messageHandler{
		repo:     repo,
		mqtt:     mqtt,
		ctx:      ctx,
		handlers: map[string]handler{},
		wp:       &pool.WorkerPool{},
	}
	return h
}

func (m *messageHandler) ProcessMessage(id string, payload []byte, connType string) error {

	var h handler
	if _, ok := m.handlers[id]; !ok {

		if strings.HasPrefix(id, "bridge") {

			h = newBridgeHandler(m.mqtt)
			m.handlers[id] = h

		} else {

			h = newDeviceHandler(m.repo, m.eventHub)
			m.handlers[id] = h
		}
	}

	return m.wp.AddTask(&messageTask{Id: id, Type: connType, Payload: payload, h: h})
}

func (m *messageHandler) Register() error {

	m.wp = pool.NewWorkerPool(1, m.ctx)
	m.wp.Run()

	m.mqtt.OnMessageHandler(func(id string, payload []byte) {
		m.ProcessMessage(id, payload, "mqtt")
	})

	return nil
}

func convertToMap(payload []byte) (map[string]interface{}, error) {

	deviceMap := make(map[string]interface{})
	err := json.Unmarshal(payload, &deviceMap)
	if err != nil {
		return nil, errors.New("invalid device data")
	}
	return deviceMap, nil
}
