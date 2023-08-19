package main

import (
	"context"
	"net/http"
	"node-herder/models/devices"
	"node-herder/transport/api"
	"node-herder/transport/controllers"
	"node-herder/transport/mqtt"
	"node-herder/transport/ws"
)

func registerApi(port int, ws ws.EventHub, hub *controllers.HubController, ctx context.Context) *api.ApiServer {

	router := api.NewRouter()
	router.GET("/ws", api.NewWsHandler(ws))
	router.GET("/", http.FileServer(http.Dir("../frontend/dist")))
	router.POST("/collect", api.NewDataCollectorHandler(hub))

	apiServer := api.NewHttpServer(
		port,
		api.WithContext(ctx),
		api.WithRouter(router),
	)

	return apiServer
}

func RegisterHub(port int, repo devices.Repository, config mqtt.MqttConfig, ctx context.Context) *api.ApiServer {

	ws := ws.NewWsHub()
	mqtt := mqtt.NewMqttClient(config)

	hub := controllers.RegisterHubController(ws, mqtt, repo, ctx)

	// TODO: see if i can remove and do it when newing it - to check onmessage hander gets used correctly
	// TODO: refactor
	mqtt.Connect()
	mqtt.Publish("zigbee2mqtt/bridge/devices", nil) // we publish so we get device data from bridge
	//

	return registerApi(port, ws, hub, ctx)
}
