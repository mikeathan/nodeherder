package hub

type Config struct {
	Mqtt MqttConfig
}

type Hub interface {
	Broadcast(eventName string, payload interface{})
	Repository() Repository
}

type HubConnector struct {
	ws   *WsServer
	mqtt *Z2MClient
	repo Repository
}

func Create(config Config) Hub {

	var h = &HubConnector{}
	h.repo = NewMemoryRepository()

	h.ws = NewWsServer()
	go h.ws.Run()

	h.mqtt = NewMqttClient(config.Mqtt)
	h.mqtt.Connect()

	return h
}

func (h *HubConnector) Repository() Repository {
	return h.repo
}

func (h *HubConnector) Broadcast(eventName string, payload interface{}) {

	h.ws.Broadcast(eventName, payload)
}
