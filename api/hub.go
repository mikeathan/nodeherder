package api

import (
	"log"
	"node-herder/api/hub"
)

var _wsServer *hub.Server
var _z2mClient *Z2MClient

type HubConfig struct {
	MqttConfig MqttConfig
	WSPath     string
}

func Init(config HubConfig) {

	if _wsServer != nil {
		log.Fatal("Init - wsServer is already initialized")
		return
	}

	_wsServer = hub.NewServer()
	go _wsServer.Run()

	_z2mClient = NewZ2MClient(config.MqttConfig)
	_z2mClient.Connect()
}

func Broadcast(eventName string, payload interface{}) {

	if _wsServer == nil {
		log.Fatal("Broadcast - eventhub is not initialized")
		return
	}

	_wsServer.Broadcast(eventName, payload)
}
