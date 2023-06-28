package hub

import "node-herder/hub/mocks"

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

	// todo : configure mqtt hanlder
	// and connect
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
