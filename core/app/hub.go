package app

import (
	"context"
	"fmt"
	"net/http"
	"node-herder/models/devices"
	"node-herder/transport/api"
	"node-herder/transport/controllers"
	"node-herder/transport/mqtt"
	"node-herder/transport/routes"
	"node-herder/transport/ws"
	"os"
)

func registerApi(port int, ws ws.EventHub, ctx context.Context) *api.ApiServer {

	router := routes.NewRouter()
	router.GET("/ws", routes.NewWsHandler(ws))

	if err := os.Mkdir("../../../frontend/dist", 0755); os.IsExist(err) {
		fmt.Printf("frontend dir doesnt exist")
	}
	router.GET("/", http.FileServer(http.Dir("../../../frontend/dist")))

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
