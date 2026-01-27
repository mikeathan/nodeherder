package hub

import (
	"context"
	"fmt"
	"node-herder/internal/api"
	"node-herder/internal/auth"
	"node-herder/internal/automations"
	"node-herder/internal/controllers"
	"node-herder/internal/fs"
	mcpserver "node-herder/internal/mcp/server"
	metrics "node-herder/internal/metrics/services"
	"node-herder/internal/mqtt"
	"node-herder/internal/ratelimiter"
	"node-herder/internal/ws"
	"node-herder/store"
	"node-herder/utils"
	"os"
	"time"
)

func registerApi(port int, ws ws.EventHub, hub *controllers.HubController, store store.AppStore, mcpServer *mcpserver.Server, ctx context.Context) *api.ApiServer {

	router := api.NewRouter()
	fservice := fs.NewFileSystem()

	// Build OAuth callback URL from env with port substitution
	callbackURL, err := utils.GetAuthCallbackURL(port)
	if err != nil {
		fatalErr := fmt.Errorf("failed to get OAuth callback URL: %w", err)
		fmt.Fprintln(os.Stderr, fatalErr)
		os.Exit(1)
	}

	authProvider := auth.NewProvider(auth.WithDefaultJWTConfig(), callbackURL)
	// middlewares

	router.UseGlobal(api.CORS)
	router.UseProtected(authProvider.Middleware())

	// websocket routing
	router.GET("/ws", api.NewWsHandler(ws))

	// api routing
	// protected routes
	router.POST("/api/metrics/query", api.NewMetricsQueryHandler(ratelimiter.NewWindowRateLimiter(4, 1*time.Second), store))
	router.POST("/api/collect", api.NewDataCollectorHandler(hub))
	router.POST("/api/logfile", api.NewLogFileHandler(fservice))
	router.POST("/api/automation/trigger", api.NewAutomationTriggerHandler(hub, 1*time.Second))
	router.GET("/api/listlogs", api.NewListFileLogsHandler(fservice))
	router.GET("/api/hubstate", api.NewHubStateHandler(store, 15*time.Minute))

	// public routes
	router.PublicGET("/api/context/devices", api.NewDeviceContextHandler(store, 1*time.Second))
	router.PublicPOST("/api/auth/token", authProvider.ServiceTokenHandler())

	// MCP routes 
	if mcpServer != nil {
		router.PublicPOST("/api/mcp", api.NewMCPHandler(mcpServer))
		router.PublicGET("/api/mcp/events", api.NewMCPEventsHandler(mcpServer))
	}

	// authentication routes
	router.AddAuthentication(authProvider.OAuth())

	apiServer := api.NewHttpServer(
		port,
		api.WithContext(ctx),
		api.WithRouter(router),
	)

	return apiServer
}

func Register(port int, store store.AppStore, ctx context.Context, enableMCP bool) *api.ApiServer {

	ws := ws.NewWsHub()
	ws.Start()

	utils.RegisterRemoteLoggerHook(ws)

	mqtt := mqtt.NewMqttClient(mqtt.WithDefaultMqttConfig())

	automationHandlers := automations.DefaultAutomationHandlers(ctx)
	hub := controllers.RegisterHubController(ws,
		store,
		mqtt,
		controllers.WithContext(ctx),
		controllers.WithAutomationHandlers(automationHandlers))

	// Create MCP server for HTTP transport
	var mcpServer *mcpserver.Server
	if enableMCP {
		utils.LogInfo("MCP HTTP transport enabled")
		querier := metrics.NewQueryService(store)
		mcpServer = mcpserver.New(store, querier)
	}

	return registerApi(port, ws, hub, store, mcpServer, ctx)
}
