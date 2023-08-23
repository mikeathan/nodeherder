package main

import (
	"context"
	"flag"
	"fmt"
	"node-herder/internal/mqtt"
	repository "node-herder/repository/devices"
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
	}

	hub := RegisterHub(port, repo, mqttConfig, ctx)
	hub.Listen()

	fmt.Println("Exited")
}
