package cmd

import (
	"context"
	"flag"
	"fmt"
	"node-herder/app"
	repository "node-herder/repository/devices"
	"node-herder/transport/mqtt"
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

	repo := repository.NewMemoryDeviceRepo()
	mqttConfig := mqtt.MqttConfig{
		Username: "sinkhole",
		Password: "mqtt2023",
		Broker:   "192.168.50.179:1883",
		Topics: []string{
			"TH1",
		},
	}

	server := app.NewHubServer(port, ctx, mqttConfig, repo)

	server.Listen()
	fmt.Println("Exited")
}
