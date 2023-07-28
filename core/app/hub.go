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

type HubConnector struct {
	eventHub ws.EventHub
	mqtt     mqtt.MqttClient
	repo     devices.Repository
	ctx      context.Context
	wp       pool.WorkerPool
	procFunc func(t pool.Task) error
}

func NewHubConnector(opts ...func(h *HubConnector)) *HubConnector {
	h := &HubConnector{
		eventHub: &mocks.NopWsServer{},
		mqtt:     &mocks.NopMqttClient{},
		repo:     &mocks.NopRepository{},
		ctx:      context.Background(),
	}

	for _, opt := range opts {
		opt(h)
	}

	h.procFunc = func(t pool.Task) error {
		return h.processPayload(t)
	}

	h.wp = *pool.NewWorkerPool(1, h.ctx, h.procFunc)
	h.wp.Start()

	h.eventHub.OnConnected(func() interface{} {
		return h.repo.ListAllDevices()
	})

	h.mqtt.OnMessageHandler(func(name string, payload []byte) {
		h.wp.AddTask(&processorTask{Id: name, Payload: payload})
	})

	return h
}

func (c *HubConnector) processPayload(task pool.Task) error {
	pTask, ok := task.(*processorTask)
	if !ok {
		return errors.New("invalid task type")
	}
	id := pTask.Id
	payload := pTask.Payload
	data, err := convertToMap(payload)
	if err != nil {
		return errors.New("failed to convert to map")
	}

	device, _ := c.repo.FindDevice(id)
	if device == nil {
		devices.CreateNewDevice(id, data)
	} else {
		if !devices.TryUpdateDevice(device, data) {
			return nil
		}
	}

	c.repo.Store(id, device)
	c.eventHub.Broadcast(ws.DeviceUpdated, device)
	return nil
}

func (c *HubConnector) Connect() error {
	err := c.mqtt.Connect()
	if err != nil {
		return err
	}

	return nil
}

func WithRepository(repo devices.Repository) func(h *HubConnector) {
	return func(h *HubConnector) { h.repo = repo }
}

func WithEventHub(eventhub ws.EventHub) func(h *HubConnector) {
	return func(h *HubConnector) { h.eventHub = eventhub }
}

func WithMqtt(mqtt mqtt.MqttClient) func(h *HubConnector) {
	return func(h *HubConnector) { h.mqtt = mqtt }
}

func WithContext(ctx context.Context) func(h *HubConnector) {
	return func(h *HubConnector) { h.ctx = ctx }
}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
