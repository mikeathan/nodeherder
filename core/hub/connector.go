package hub

import (
	"fmt"
	"node-herder/hub/mocks"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

type Config struct {
	Mqtt MqttConfig
}

type Hub interface {
	RegisterConnection(conn *websocket.Conn)
}

type HubConnector struct {
	ws   WsServer
	mqtt MqttClient
	repo Repository
}

func (h *HubConnector) RegisterConnection(conn *websocket.Conn) {
	client := h.ws.RegisterNewClient(conn)

	devices := h.repo.ListAllDevices()
	client.Broadcast(DeviceUpdated, devices)
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

func WithWsServer(wsServer WsServer) func(h *HubConnector) {
	return func(h *HubConnector) { h.ws = wsServer }
}

func WithMqttClient(mqtt MqttClient) func(h *HubConnector) {
	return func(h *HubConnector) { h.mqtt = mqtt }
}

func Create(config Config, opts ...func(h *HubConnector)) Hub {

	var h = &HubConnector{
		ws:   &mocks.NopWsServer{},
		mqtt: &mocks.NopMqttClient{},
		repo: &mocks.NopRepository{}}

	for _, opt := range opts {
		opt(h)
	}

	// TODO
	// will need refactoring
	// configure  mqtt
	//h.mqtt.WithMessageHandler(h.messageHandler())
	h.mqtt.Connect()
	return h
}

func (h *HubConnector) DeviceUpdated(name string, payload []byte) {

	h.repo.Store(name, payload)

	h.ws.Broadcast(DeviceUpdated, NewHubMessage(name, payload))
}
