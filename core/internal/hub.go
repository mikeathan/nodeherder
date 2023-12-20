package hub

import (
	"context"
	"net/http"
	"node-herder/internal/api"
	"node-herder/internal/controllers"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
)

func registerApi(port int, ws ws.EventHub, hub *controllers.HubController, ctx context.Context) *api.ApiServer {

	router := api.NewRouter()
	router.GET("/ws", api.NewWsHandler(ws))

	// TODO: needs refactoring - need to use regex in the path and pass a http.serveFile
	rootRedirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := "../frontend/dist/index.html"
		http.ServeFile(w, r, name)
	})
	router.GET("/viewer", http.StripPrefix("/viewer", rootRedirectHandler))
	router.GET("/creator", http.StripPrefix("/creator", rootRedirectHandler))
	router.GET("/devicepage", http.StripPrefix("/devicepage", rootRedirectHandler))
	router.GET("/editor", http.StripPrefix("/editor", rootRedirectHandler))
	//router.GET("/", api.NewFileHandler("../frontend/dist/index.html"))

	router.GET("/", http.FileServer(http.Dir("../frontend/dist")))
	router.POST("/collect", api.NewDataCollectorHandler(hub))

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
