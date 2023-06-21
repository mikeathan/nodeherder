package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"node-herder/api"
	"node-herder/hub"
	"os"
	"os/signal"
	"syscall"
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

	cfg := hub.Config{
		Mqtt: hub.MqttConfig{
			Username: "sinkhole",
			Password: "mqtt2023",
			Broker:   "192.168.50.179:1883",
			Topics: []string{
				"TH1",
			},
		}}

	hub := hub.Create(cfg)

	router := api.NewRouter()
	router.GET("/ws", api.NewWsHandler(hub))
	router.GET("/", http.FileServer(http.Dir("../../frontend/dist")))

	apiServer := api.NewServer(
		port,
		api.WithRouter(router),
		api.WithContext(ctx),
	)

	apiServer.Listen()
	fmt.Println("Exited")

}
