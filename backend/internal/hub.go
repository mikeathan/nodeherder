package hub

import (
	"context"
	"node-herder/internal/api"
	"node-herder/internal/controllers"
	"node-herder/internal/fs"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/store"
	"node-herder/utils"
	"time"
)

func registerApi(port int, ws ws.EventHub, hub *controllers.HubController, store store.AppStore, ctx context.Context) *api.ApiServer {

	router := api.NewRouter()
	fservice := fs.NewFileSystem()

	//websocket routing
	router.GET("/ws", api.NewWsHandler(ws))

	// api routing
	router.POST("/collect", api.NewDataCollectorHandler(hub))

	router.POST("/logfile", api.NewLogFileHandler(fservice))
	router.GET("/listlogs", api.NewListFileLogsHandler(fservice))
	router.GET("/hubstate", api.NewHubStateHandler(store, 15*time.Minute))

	// file routing
	// fs := api.NewFileServer("../frontend/dist")
	// router.GET("/consoleviewer", fs.Resolve(false))
	// router.GET("/deviceDashboard", fs.Resolve(false))
	// router.GET("/settings", fs.Resolve(false))
	// router.GET("/viewer", fs.Resolve(false))
	// router.GET("/creator", fs.Resolve(false))
	// router.GET("/devicepage", fs.Resolve(false))
	// router.GET("/editor", fs.Resolve(false))
	// router.GET("/", fs.Resolve(true))

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

	return registerApi(port, ws, hub, store, ctx)
}
