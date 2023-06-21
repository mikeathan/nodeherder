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

func Create(config Config) Hub {

	var h = &HubConnector{}
	h.repo = NewMemoryRepository()

	h.ws = newWsServer()
	go h.ws.Run()

	h.mqtt = newMqttClient(config.Mqtt)
	h.mqtt.WithMessageHandler(h.messageHandler())
	h.mqtt.Connect()

	return h
}

func (h *HubConnector) DeviceUpdated(name string, payload []byte) {

	h.repo.Store(name, payload)

	h.ws.Broadcast(DeviceUpdated, NewWsMessage(name, payload))
}

func (h *HubConnector) Repository() Repository {
	return h.repo
}
