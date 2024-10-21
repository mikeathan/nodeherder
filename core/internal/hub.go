package hub

import (
	"context"
	"node-herder/internal/api"
	"node-herder/internal/controllers"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/store"
	"node-herder/utils"
)

func registerApi(port int, ws ws.EventHub, hub *controllers.HubController, ctx context.Context) *api.ApiServer {

	router := api.NewRouter()

	//websocket routing
	router.GET("/ws", api.NewWsHandler(ws))

	// api routing
	router.POST("/collect", api.NewDataCollectorHandler(hub))
	router.POST("/logfile",api.NewLogFileHandler(utils.FileSystemLoader{}))
	router.GET("/listlogs", api.NewListFileLogsHandler(utils.FileSystemWalker{}))

	// file routing
	fs := api.NewFileServer("../frontend/dist")
	router.GET("/consoleviewer", fs.Resolve(false))
	router.GET("/settings", fs.Resolve(false))
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

func Register(port int, store store.AppStore, config mqtt.MqttConfig, ctx context.Context) *api.ApiServer {

	ws := ws.NewWsHub()
	ws.Start()

	utils.RegisterRemoteLoggerHook(ws)

	mqtt := mqtt.NewMqttClient(config)
	hub := controllers.RegisterHubController(ws, store, mqtt, ctx)

	return registerApi(port, ws, hub, ctx)
}
