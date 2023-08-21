package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils/pool"
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
	handler  *messageHandler
	pool     *pool.WorkerPool
}

func RegisterHubController(ws ws.EventHub, mqtt mqtt.MqttClient, repo devices.Repository, ctx context.Context) *HubController {
	h := &HubController{
		eventHub: ws,
		mqtt:     mqtt,
		repo:     repo,
	}

	h.handler = newMessageHandler(repo, mqtt, ws, ctx)
	h.handler.Register()

	h.eventHub.OnConnected(func() interface{} {
		return h.repo.ListAllDevices()
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

	return c.handler.ProcessMessage(id, bytes, connType)
}
