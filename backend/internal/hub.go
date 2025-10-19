package hub

import (
	"context"
	"node-herder/internal/api"
	"node-herder/internal/auth"
	"node-herder/internal/automations"
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

	authProvider := auth.NewProvider(auth.WithDefaultJWTConfig())

	// middlewares
	router.Use(authProvider.Middleware())
	router.Use(api.CORS) // that needs to execute first

	// websocket routing
	router.GET("/ws", api.NewWsHandler(ws))

	// api routing
	router.POST("/api/collect", api.NewDataCollectorHandler(hub))
	router.POST("/api/logfile", api.NewLogFileHandler(fservice))
	router.POST("/api/automation/trigger", api.NewAutomationTriggerHandler(hub, 1*time.Second))
	router.GET("/api/listlogs", api.NewListFileLogsHandler(fservice))
	router.GET("/api/hubstate", api.NewHubStateHandler(store, 15*time.Minute))

	// authentication routes
	router.AddAuthentication(authProvider.OAuth())

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

	automationHandlers := automations.DefaultAutomationHandlers(ctx)
	hub := controllers.RegisterHubController(ws,
		store,
		mqtt,
		controllers.WithContext(ctx),
		controllers.WithAutomationHandlers(automationHandlers))

	return registerApi(port, ws, hub, store, ctx)
}
