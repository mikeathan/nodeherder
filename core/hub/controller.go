package hub

import (
	"context"
	"errors"
	"fmt"
	"node-herder/devices"
	"node-herder/hub/pool"
	"node-herder/mocks"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type processorTask struct {
	Id      string
	Payload interface{}
}

func (m *processorTask) OnFailure(err error) {
	fmt.Printf("Job: %s Error: %s", m.Id, err.Error())
}

type Controller struct {
	eventHub EventHub
	mqtt     MqttClient
	repo     devices.Repository
	ctx      context.Context
	wp       pool.WorkerPool
	procFunc func(t pool.Task) error
}

func NewController(opts ...func(h *Controller)) *Controller {
	h := &Controller{
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

	h.mqtt.OnMessageHandler(func(client mqtt.Client, msg mqtt.Message) {

		var name = SanitizeTopic(msg.Topic())
		var payload = msg.Payload()

		fmt.Printf("mqtt Message => Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())

		h.wp.AddTask(&processorTask{Id: name, Payload: payload})
	})

	return h
}

func (c *Controller) processPayload(task pool.Task) error {
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
	c.eventHub.Broadcast(DeviceUpdated, device)
	return nil
}

func (c *Controller) Connect() error {
	err := c.mqtt.Connect()
	if err != nil {
		return err
	}

	return nil
}

func WithRepository(repo devices.Repository) func(h *Controller) {
	return func(h *Controller) { h.repo = repo }
}

func WithEventHub(eventhub EventHub) func(h *Controller) {
	return func(h *Controller) { h.eventHub = eventhub }
}

func WithMqtt(mqtt MqttClient) func(h *Controller) {
	return func(h *Controller) { h.mqtt = mqtt }
}

func WithContext(ctx context.Context) func(h *Controller) {
	return func(h *Controller) { h.ctx = ctx }
}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
