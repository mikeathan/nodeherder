package hub

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	Mqtt MqttConfig
}

type Hub interface {
	Repository() Repository
}

type HubConnector struct {
	ws   *WsServer
	mqtt *Z2MClient
	repo Repository
}

func (h *HubConnector) messageHandler() func(client mqtt.Client, msg mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		var name = SanitizeTopic(msg.Topic())
		var payload = msg.Payload()
		fmt.Printf("DEBUG - [Hub connector] mqtt message => Topic: %s, Payload; %s\n", name, payload)

		h.DeviceUpdated(name, payload)
	}
}

func WithRepository(repo Repository) func(h *HubConnector) {
	return func(h *HubConnector) { h.repo = repo }
}

func WithWsServer(wsServer *WsServer) func(h *HubConnector) {
	return func(h *HubConnector) { h.ws = wsServer }
}

func WithMqttClient(mqtt *Z2MClient) func(h *HubConnector) {
	return func(h *HubConnector) { h.mqtt = mqtt }
}

func Create(config Config, opts ...func(h *HubConnector)) Hub {

	var h = &HubConnector{}
	for _, opt := range opts {
		opt(h)
	}

	// Temporary
	if h.repo == nil {
		h.repo = NewMemoryRepository()
	}

	// TODO: cant new ws and mqtt because we cant mock them
	// temporay
	// needs to init an empty mock object instead
	if h.ws == nil {
		h.ws = NewWsServer()
		go h.ws.Run()
	}

	// temporay
	// needs to init an empty mock object instead
	if h.mqtt == nil {
		h.mqtt = NewMqttClient(config.Mqtt)
		h.mqtt.WithMessageHandler(h.messageHandler())
		h.mqtt.Connect()
	}

	return h
}

func (h *HubConnector) DeviceUpdated(name string, payload []byte) {

	h.repo.Store(name, payload)

	h.ws.Broadcast(DeviceUpdated, NewWsMessage(name, payload))
}

func (h *HubConnector) Repository() Repository {
	return h.repo
}
