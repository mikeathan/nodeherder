package hub

import (
	"node-herder/hub/mocks"

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
	repo Repository
}

func (h *HubConnector) RegisterConnection(conn *websocket.Conn) {
	client := h.ws.RegisterNewClient(conn)

	devices := h.repo.ListAllDevices()
	client.Broadcast(DeviceUpdated, devices)
}

func WithRepository(repo Repository) func(h *HubConnector) {
	return func(h *HubConnector) { h.repo = repo }
}

func WithWsServer(wsServer WsServer) func(h *HubConnector) {
	return func(h *HubConnector) { h.ws = wsServer }
}

func Create(config Config, opts ...func(h *HubConnector)) Hub {

	var h = &HubConnector{
		ws:   &mocks.NopWsServer{},
		repo: &mocks.NopRepository{}}

	for _, opt := range opts {
		opt(h)
	}

	return h
}
