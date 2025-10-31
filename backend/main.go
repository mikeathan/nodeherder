package main

import (
	"context"
	"flag"
	"fmt"
	hub "node-herder/internal"
	"node-herder/internal/mqtt"
	"node-herder/store"
	"node-herder/utils"
	"os"
	"os/signal"
	"syscall"
)

type cmdArgs struct {
	port      int
	buildType string
	logLevel  string
}

func readArgs() *cmdArgs {

	port := flag.Int("port", 4110, "port number")
	buildType := flag.String("buildType", "", "client build type")
	logLevel := flag.String("logLevel", "info", "logging level")

	flag.Parse()
	if *port <= 0 {

		fmt.Print("Invalid port number")
		os.Exit(-1)
	}

	return &cmdArgs{port: *port, buildType: *buildType, logLevel: *logLevel}
}

func main() {

	args := readArgs()

	utils.InitFileLogger()
	utils.SetLogLevel(args.logLevel)

	err := utils.LoadEnviromentConfig()
	if err != nil {
		utils.LogErrorf("error loading enviroment config: %v", err.Error())
		os.Exit(-1)
	}

	ctx, cancelCtx := context.WithCancel(context.Background())

	utils.LogInfo("starting up server")
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer close(c)
		<-c
		utils.LogWarn("system termination signal received.")
		cancelCtx()
	}()

	mqttBrokerURL, err := utils.GetMQTTBrokerURL()
	if err != nil {
		utils.LogErrorf("error getting MQTT broker URL: %v", err.Error())
		cancelCtx()
	}

	mqttConfig := mqtt.MqttConfig{
		Username:   "sinkhole",
		Password:   "mqtt2023",
		Broker:     mqttBrokerURL,
		ClientType: args.buildType,
	}

	store, err := store.Create(ctx)
	if err != nil {
		utils.LogErrorf("error creating store: %v", err.Error())
		cancelCtx()
	}
	h := hub.Register(args.port, store, mqttConfig, ctx)
	h.Listen()
	utils.LogInfo("exit")
}
