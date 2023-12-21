package hub

import (
	"context"
	"node-herder/internal/api"
	"node-herder/internal/controllers"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
)

func registerApi(port int, ws ws.EventHub, hub *controllers.HubController, ctx context.Context) *api.ApiServer {

	router := api.NewRouter()

	//websocket routing
	router.GET("/ws", api.NewWsHandler(ws))

	// api routing
	router.POST("/collect", api.NewDataCollectorHandler(hub))

	// file routing
	fs := api.NewFileServer("../frontend/dist")
	router.GET("/viewer", fs.Resolve(false))
	router.GET("/creator", fs.Resolve(false))
	router.GET("/devicepage", fs.Resolve(false))
	router.GET("/editor", fs.Resolve(false))
	router.GET("/", fs.Resolve(true))

	apiServer := api.NewHttpServer(
		port,
		api.WithContext(ctx),
		api.WithRouter(router),
	)

	return apiServer
}

func Register(port int, repo devices.Repository, config mqtt.MqttConfig, ctx context.Context) *api.ApiServer {

	ws := ws.NewWsHub()
	mqtt := mqtt.NewMqttClient(config)

	hub := controllers.RegisterHubController(ws, mqtt, repo, ctx)

	return registerApi(port, ws, hub, ctx)
}
