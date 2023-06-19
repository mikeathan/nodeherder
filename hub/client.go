package hub

import (
	"encoding/json"
	"log"
	"net/http"
	"node-herder/hub/ws"
)

var _wsServer *ws.Server
var _z2mClient *Z2MClient

type HubConfig struct {
	MqttConfig MqttConfig
	WSPath     string
}

func Init(config HubConfig) *ws.WsHandler {

	if _wsServer != nil {
		log.Fatal("Init - wsServer is already initialized")
		return nil
	}

	_wsServer = ws.NewServer()
	go _wsServer.Run()

	_z2mClient := NewZ2MClient(config.MqttConfig)
	_z2mClient.Connect()

	handler := ws.NewHandler(config.WSPath)

	// TODO: configure in http package
	http.Handle(config.WSPath, handler)
	return handler
}

func Broadcast(event string, data interface{}) {

	if _wsServer == nil {
		log.Fatal("Broadcast - eventhub is not initialized")
		return
	}

	var wsData = ws.WsEvent{Name: event, Data: data}
	bytes, err := json.Marshal(wsData)
	if err != nil {
		panic(err)
	}
	_wsServer.Broadcast(bytes)
}
