package hub

import (
	"fmt"
	"node-herder/hub/mocks"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Connector interface {
}

type HubConnector struct {
	ws   WsServer
	mqtt MqttClient
	repo Repository
}

func (h *HubConnector) NewHubConnector(opts ...func(h *HubConnector)) *HubConnector {
	connector := &HubConnector{
		ws:   &mocks.NopWsServer{},
		mqtt: &mocks.NopMqttClient{},
		repo: &mocks.NopRepository{},
	}
	for _, opt := range opts {
		opt(connector)
	}

	h.ws.WithOnConnected(func() interface{} { return h.repo.ListAllDevices() })

	h.mqtt.OnMessageHandler(func(client mqtt.Client, msg mqtt.Message) {

		var name = SanitizeTopic(msg.Topic())
		var payload = msg.Payload()
		fmt.Printf("mqtt Message => Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
		h.repo.StoreJson(name, payload)
		device, _ := h.repo.FindDevice(name)
		h.ws.Broadcast(DeviceUpdated, device)
	})

	return connector
}

func (h *HubConnector) WithRepository() {

}

func WithRepository(repo Repository) func(h *HubConnector) {
	return func(h *HubConnector) { h.repo = repo }
}

func WithWs(ws WsServer) func(h *HubConnector) {
	return func(h *HubConnector) { h.ws = ws }
}

func WithMqtt(mqtt MqttClient) func(h *HubConnector) {
	return func(h *HubConnector) { h.mqtt = mqtt }
}
