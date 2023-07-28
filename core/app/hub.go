package app

import (
	"context"
	"errors"
	"fmt"
	"node-herder/common/pool"
	"node-herder/mocks"
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

type hubController struct {
	eventHub ws.EventHub
	mqtt     mqtt.MqttClient
	repo     devices.Repository
	ctx      context.Context
	pool     pool.WorkerPool
	procFunc func(t pool.Task) error
}

func newHubController(opts ...func(h *hubController)) *hubController {
	h := &hubController{
		eventHub: &mocks.NopWsServer{},
		mqtt:     &mocks.NopMqttClient{},
		repo:     &mocks.NopRepository{},
		ctx:      context.Background(),
	}

	for _, opt := range opts {
		opt(h)
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

func (c *hubController) processPayload(id string, payload map[string]interface{}) error {

	// REFACTOR !!!!!!!!!!!!!!!!!!!!
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

func (c *hubController) Connect() error {
	err := c.mqtt.Connect()
	if err != nil {
		return err
	}

	return nil
}

func WithRepository(repo devices.Repository) func(h *hubController) {
	return func(h *hubController) { h.repo = repo }
}

func WithEventHub(eventhub ws.EventHub) func(h *hubController) {
	return func(h *hubController) { h.eventHub = eventhub }
}

func WithMqtt(mqtt mqtt.MqttClient) func(h *hubController) {
	return func(h *hubController) { h.mqtt = mqtt }
}

func WithContext(ctx context.Context) func(h *hubController) {
	return func(h *hubController) { h.ctx = ctx }
}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
