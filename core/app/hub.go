package app

import (
	"context"
	"net/http"
	"node-herder/models/devices"
	"node-herder/transport/api"
	"node-herder/transport/controllers"
	"node-herder/transport/mqtt"
	"node-herder/transport/routes"
	"node-herder/transport/ws"
)

func registerApi(port int, ws ws.EventHub, ctx context.Context) *api.ApiServer {

	router := routes.NewRouter()
	router.GET("/ws", routes.NewWsHandler(ws))

	router.GET("/", http.FileServer(http.Dir("../frontend/dist")))

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

	// TODO: see if i can remove and do it when newing it - to check onmessage hander gets used correctly
	mqtt.Connect()

	controllers.RegisterHubController(ws, mqtt, repo, ctx)

	return registerApi(port, ws, ctx)
}
