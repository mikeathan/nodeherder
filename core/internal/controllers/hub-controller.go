package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils/pool"
	"strings"
)

type apiTask struct {
	Id      string
	Payload map[string]interface{}
	Type    string
}

func (m *apiTask) OnFailure(err error) {
	fmt.Printf("Job: %s Error: %s", m.Id, err.Error())
}

type HubController struct {
	eventHub ws.EventHub
	mqtt     mqtt.MqttClient
	repo     devices.Repository
	wp       *pool.WorkerPool
	handlers map[string]handler
}

func RegisterHubController(ws ws.EventHub, mqtt mqtt.MqttClient, repo devices.Repository, ctx context.Context) *HubController {
	h := &HubController{
		eventHub: ws,
		mqtt:     mqtt,
		repo:     repo,
		handlers: map[string]handler{},
	}

	h.wp = pool.NewWorkerPool(1, ctx)
	h.wp.Run()

	h.eventHub.OnConnected(func() interface{} {
		return h.repo.ListAllDevices()
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

func (m *HubController) ProcessMessage(id string, payload []byte, connType string) error {

	if _, ok := m.handlers[id]; !ok {

		if strings.HasPrefix(id, "bridge") {

			var h = newBridgeHandler(m.mqtt)
			m.handlers[id] = h

		} else {

			var h = newDeviceHandler(m.repo, m.eventHub)
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
