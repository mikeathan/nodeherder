package main

import (
	"context"
	"flag"
	"fmt"
	hub "node-herder/internal"
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
	enableMCP bool
}

func readArgs() *cmdArgs {

	port := flag.Int("port", 4110, "port number")
	buildType := flag.String("buildType", "", "client build type")
	logLevel := flag.String("logLevel", "info", "logging level")
	enableMCP := flag.Bool("mcp", true, "enable MCP server for LLM clients")

	flag.Parse()
	if *port <= 0 {
		fmt.Print("Invalid port number")
		os.Exit(-1)
	}

	return &cmdArgs{port: *port, buildType: *buildType, logLevel: *logLevel, enableMCP: *enableMCP}
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
	}

	// Start HTTP server (and optionally MCP server)
	utils.LogInfo("starting up server")
	h := hub.Register(args.port, appStore, ctx, args.enableMCP)
	h.Listen()
	utils.LogInfo("exit")
}
