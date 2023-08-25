package main

import (
	"context"
	"flag"
	"fmt"
	hub "node-herder/internal"
	"node-herder/internal/mqtt"
	repository "node-herder/repository/devices"
	logger "node-herder/utils"
	"os"
	"os/signal"
	"syscall"
)

type cmdArgs struct {
	port      int
	buildType string
}

func readArgs() *cmdArgs {

	port := flag.Int("port", 4100, "port number")
	buildType := flag.String("buildType", "", "client build type")

	flag.Parse()
	if *port <= 0 {

		fmt.Print("Invalid port number")
		os.Exit(-1)
	}

	return &cmdArgs{port: *port, buildType: *buildType}
}

func main() {

	args := readArgs()

	ctx, cancelCtx := context.WithCancel(context.Background())

	logger.GetInstance().Info("starting up server")
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer close(c)
		<-c
		logger.GetInstance().Warning("IGTERM signal notified")
		cancelCtx()
	}()

	repo := repository.NewMemoryDeviceRepo()
	mqttConfig := mqtt.MqttConfig{
		Username:   "sinkhole",
		Password:   "mqtt2023",
		Broker:     "192.168.50.179:1883",
		ClientType: args.buildType,
	}

	h := hub.Register(args.port, repo, mqttConfig, ctx)
	h.Listen()
	logger.GetInstance().Info("exit")
}
