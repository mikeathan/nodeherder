package main

import (
	"context"

	hub "node-herder/internal"
	"node-herder/store"
	"node-herder/utils"
	"os"
	"os/signal"
	"syscall"
)

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer close(c)
		<-c
		utils.LogWarn("system termination signal received.")
		cancelCtx()
	}()

	appStore, err := store.Create(ctx)
	if err != nil {
		utils.LogErrorf("error creating store: %v", err.Error())
		cancelCtx()
		os.Exit(-1)
	}

	utils.LogInfo("starting up server")
	h := hub.Register(args.port, appStore, ctx, args.enableMCP)
	h.Listen()
	utils.LogInfo("exit")
}
