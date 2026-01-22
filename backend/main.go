package main

import (
	"context"
	"flag"
	"fmt"
	hub "node-herder/internal"
	mcpserver "node-herder/internal/mcp/server"
	metrics "node-herder/internal/metrics/services"
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
	mcpOnly   bool // Run as MCP server only (no HTTP server)
}

func readArgs() *cmdArgs {

	port := flag.Int("port", 4110, "port number")
	buildType := flag.String("buildType", "", "client build type")
	logLevel := flag.String("logLevel", "info", "logging level")
	enableMCP := flag.Bool("mcp", true, "enable MCP server for LLM clients")
	mcpOnly := flag.Bool("mcp-only", false, "run as MCP server only (for MCP Inspector)")

	flag.Parse()
	if !*mcpOnly && *port <= 0 {
		fmt.Print("Invalid port number")
		os.Exit(-1)
	}

	return &cmdArgs{port: *port, buildType: *buildType, logLevel: *logLevel, enableMCP: *enableMCP, mcpOnly: *mcpOnly}
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

	// MCP-only mode: run just the MCP server (for MCP Inspector)
	if args.mcpOnly {
		utils.LogInfo("starting MCP server in stdio-only mode")
		querier := metrics.NewQueryService(appStore)
		srv := mcpserver.New(appStore, querier)
		if err := srv.ServeStdio(); err != nil {
			utils.LogErrorf("MCP server error: %v", err.Error())
			os.Exit(1)
		}
		return
	}

	// Start HTTP server (and optionally MCP server)
	utils.LogInfo("starting up server")
	h := hub.Register(args.port, appStore, ctx, args.enableMCP)
	h.Listen()
	utils.LogInfo("exit")
}
