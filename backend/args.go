package main

import (
	"flag"
	"fmt"
	"os"
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
	enableMCP := flag.Bool("mcp", false, "enable MCP HTTP endpoints for LLM clients")

	flag.Parse()
	if *port <= 0 {
		fmt.Print("Invalid port number")
		os.Exit(-1)
	}

	return &cmdArgs{port: *port, buildType: *buildType, logLevel: *logLevel, enableMCP: *enableMCP}
}
