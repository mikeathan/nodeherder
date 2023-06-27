package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"node-herder/hub"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func readPort() int {

	port := flag.Int("port", 4100, "port number")
	flag.Parse()
	if *port <= 0 {

		fmt.Print("Invalid port number")
		os.Exit(-1)
	}

	return *port
}

func main() {

	port := readPort()

	ctx, cancelCtx := context.WithCancel(context.Background())

	fmt.Println("Starting up server.")
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer close(c)
		<-c
		fmt.Println("SIGTERM signal notified")
		cancelCtx()
	}()

	repo := hub.NewMemoryRepository()
	wsConfig := hub.WsConfig{
		OnConnected: func() interface{} {
			devices := repo.ListAllDevices()
			bytes, err := hub.ToJson(devices)
			if err != nil {
				fmt.Printf("OnConnected error: %s \n", err)
				return []byte{}
			}
			return bytes
		}}

	ws := hub.NewWsHub(wsConfig)

	mqttConfig := hub.MqttConfig{
		Username: "sinkhole",
		Password: "mqtt2023",
		Broker:   "192.168.50.179:1883",
		Topics: []string{
			"TH1",
		},
		MessageHandler: func(client mqtt.Client, msg mqtt.Message) {
			var name = hub.SanitizeTopic(msg.Topic())
			var payload = msg.Payload()
			fmt.Printf("mqtt Message => Topic: %s, Payload; %s\n", msg.Topic(), msg.Payload())
			repo.Store(name, payload)
			ws.Broadcast(hub.DeviceUpdated, hub.NewHubMessage(name, payload))
		},
	}

	mqtt := hub.NewMqttClient(mqttConfig)
	mqtt.Connect()

	// todo:
	// ws service
	// on event call func
	// connector := hub.Create(config,
	// 	hub.WithRepository(repo),
	// 	hub.WithWsServer(ws))

	router := hub.NewRouter()
	router.GET("/ws", hub.NewWsHandler(ws))
	router.GET("/", http.FileServer(http.Dir("../../frontend/dist")))

	apiServer := hub.NewHttpServer(
		port,
		hub.WithRouter(router),
		hub.WithContext(ctx),
	)

	apiServer.Listen()
	fmt.Println("Exited")

}
