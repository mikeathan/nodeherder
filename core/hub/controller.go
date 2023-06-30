package hub

import (
	"fmt"
	"node-herder/hub/mocks"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Controller struct {
	eventHub EventHub
	mqtt     MqttClient
	repo     Repository
}

func NewController(opts ...func(h *Controller)) (*Controller, error) {
	h := &Controller{
		eventHub: &mocks.NopWsServer{},
		mqtt:     &mocks.NopMqttClient{},
		repo:     &mocks.NopRepository{},
	}

	for _, opt := range opts {
		opt(h)
	}

	h.eventHub.OnConnected(func() interface{} {
		return h.repo.ListAllDevices()
	})

	h.mqtt.OnMessageHandler(func(client mqtt.Client, msg mqtt.Message) {

		var name = SanitizeTopic(msg.Topic())
		var payload = msg.Payload()

		fmt.Printf("mqtt Message => Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
		h.repo.StoreJson(name, payload)

		device, _ := h.repo.FindDevice(name)
		h.eventHub.Broadcast(DeviceUpdated, device)
	})

	err := h.mqtt.Connect()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func WithRepository(repo Repository) func(h *Controller) {
	return func(h *Controller) { h.repo = repo }
}

func WithEventHub(eventhub EventHub) func(h *Controller) {
	return func(h *Controller) { h.eventHub = eventhub }
}

func WithMqtt(mqtt MqttClient) func(h *Controller) {
	return func(h *Controller) { h.mqtt = mqtt }
}
