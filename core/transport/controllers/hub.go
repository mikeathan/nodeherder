package controllers

import (
	"context"
	"errors"
	"fmt"
	"node-herder/common/pool"
	"node-herder/models/devices"
	"node-herder/transport/mqtt"
	"node-herder/transport/ws"
)

type processorTask struct {
	Id      string
	Payload interface{}
}

func (m *processorTask) OnFailure(err error) {
	fmt.Printf("Job: %s Error: %s", m.Id, err.Error())
}

type HubController struct {
	eventHub ws.EventHub
	mqtt     mqtt.MqttClient
	repo     devices.Repository
	ctx      context.Context
	pool     pool.WorkerPool
	procFunc func(t pool.Task) error
}

func RegisterHubController(ws ws.EventHub, mqtt mqtt.MqttClient, repo devices.Repository, ctx context.Context) *HubController {
	h := &HubController{
		eventHub: ws,
		mqtt:     mqtt,
		repo:     repo,
		ctx:      ctx,
	}

	// TODO:
	// do i need to store procFunc as member variable
	h.procFunc = func(task pool.Task) error {
		// TODO
		pTask, ok := task.(*processorTask)
		if !ok {
			return errors.New("invalid task type")
		}

		payload, err := convertToMap(pTask.Payload)
		if err != nil {
			return errors.New("failed to convert to map")
		}
		return h.processPayload(pTask.Id, payload)
	}

	h.pool = *pool.NewWorkerPool(1, h.ctx)
	h.pool.Run(h.procFunc)

	h.eventHub.OnConnected(func() interface{} {
		return h.repo.ListAllDevices()
	})

	h.mqtt.OnMessageHandler(func(name string, payload []byte) {
		h.pool.AddTask(&processorTask{Id: name, Payload: payload})
	})

	return h
}

func (c *HubController) processPayload(id string, payload map[string]interface{}) error {

	device, _ := c.repo.FindDevice(id)
	if device == nil {
		devices.CreateNewDevice(id, payload)
	} else {
		if !device.TryUpdateDevice(payload) {
			return nil
		}
	}

	c.repo.Store(id, device)
	c.eventHub.Broadcast(ws.DeviceUpdated, device)
	return nil
}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
