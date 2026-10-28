package hub

import (
	"context"
	"fmt"
	"node-herder/internal/api"
	"node-herder/internal/auth"
	"node-herder/internal/automations"
	"node-herder/internal/controllers"
	"node-herder/internal/fs"
	"node-herder/internal/integrations"
	mcpserver "node-herder/internal/mcp/server"
	mcphttp "node-herder/internal/mcp/transport/http"

	metrics "node-herder/internal/metrics/services"
	"node-herder/internal/mqtt"
	"node-herder/internal/ratelimiter"
	"node-herder/internal/ws"
	"node-herder/store"
	"node-herder/utils"
	"time"
)

func registerApi(port int, ws ws.EventHub, hub *controllers.HubController, store store.AppStore, mcpServer *mcpserver.Server, ctx context.Context, options ...Option) (*api.ApiServer, error) {
	router, err := newRouter(port, ws, hub, store, mcpServer, ctx, options...)
	if err != nil {
		return nil, err
	}
	return api.NewHttpServer(port, api.WithContext(ctx), api.WithRouter(router)), nil
}

func newRouter(port int, ws ws.EventHub, hub *controllers.HubController, store store.AppStore, mcpServer *mcpserver.Server, ctx context.Context, options ...Option) (*api.Router, error) {

	router := api.NewRouter()
	fservice := fs.NewFileSystem()

	// Build OAuth callback URL from env with port substitution
	callbackURL, err := utils.GetAuthCallbackURL(port)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth callback URL: %w", err)
	}

	authProvider := auth.NewProvider(auth.WithDefaultJWTConfig(), callbackURL)
	// middlewares

	router.UseGlobal(api.CORS)
	router.UseProtected(authProvider.Middleware())
	configuration := optionsConfig{}
	for _, option := range options {
		option(&configuration)
	}
	if configuration.connections != nil {
		api.RegisterIntegrationRoutes(router, configuration.connections)
	}
	var homeSources api.HomeIntegrations
	if configuration.connections != nil {
		homeSources = configuration.connections
	}
	homeLayout := api.NewHomeLayoutHandler(store.AppConfig(), store, homeSources)
	router.GET("/api/home/layout", homeLayout)
	router.PUT("/api/home/layout", homeLayout)

	// websocket routing
	router.GET("/ws", api.NewWsHandler(ws))

	// api routing
	// protected routes
	router.POST("/api/metrics/query", api.NewMetricsQueryHandler(ratelimiter.NewWindowRateLimiter(4, 1*time.Second), store))
	router.POST("/api/collect", api.NewDataCollectorHandler(hub))
	router.POST("/api/logfile", api.NewLogFileHandler(fservice))
	router.POST("/api/automation/trigger", api.NewAutomationTriggerHandler(hub, 1*time.Second))
	router.POST("/api/assistant/message", api.NewAssistantMessageHandler(store))
	router.GET("/api/listlogs", api.NewListFileLogsHandler(fservice))
	router.GET("/api/hubstate", api.NewHubStateHandler(store, 15*time.Minute))
	router.GET("/api/assistant/conversations", api.NewAssistantConversationsHandler(store))
	router.GET("/api/assistant/history/:id", api.NewAssistantHistoryHandler(store))
	router.DELETE("/api/assistant/history/:id", api.NewAssistantDeleteHandler(store))

	// public routes
	router.PublicGET("/api/context/devices", api.NewDeviceContextHandler(store, 1*time.Second))
	router.PublicPOST("/api/auth/token", authProvider.ServiceTokenHandler())

	// MCP routes
	if mcpServer != nil {
		sseHandler := mcphttp.NewSSEHandler(mcpServer)
		context.AfterFunc(ctx, sseHandler.Close)
		router.PublicPOST("/api/mcp", mcphttp.NewMCPHandler(mcpServer, sseHandler))
		router.PublicGET("/api/mcp", sseHandler)
		router.PublicGET("/api/mcp/events", sseHandler)
	}

	// authentication routes
	router.AddAuthentication(authProvider.OAuth())

	return router, nil
}

func Register(port int, store store.AppStore, ctx context.Context, options ...Option) (*api.ApiServer, error) {

	ws := ws.NewWsHub()
	ws.Start()
	context.AfterFunc(ctx, func() { _ = ws.Close() })

	utils.RegisterRemoteLoggerHook(ws)

	mqtt := mqtt.NewMqttClient(mqtt.WithDefaultMqttConfig())

	querier := metrics.NewQueryService(store)
	mcpServer := mcpserver.New(store, store.AppConfig(), querier)

	utils.LogInfo("Initializing MCP Server")
	if err := mcpServer.Initialize(); err != nil {
		utils.LogErrorf("Failed to initialize MCP server: %v", err)
	}

	automationHandlers := automations.DefaultAutomationHandlers(ctx)
	controllerOpts := []controllers.HubControllerOption{
		controllers.WithContext(ctx),
		controllers.WithAutomationHandlers(automationHandlers),
	}
	controllerOpts = append(controllerOpts, controllers.WithMCPServer(mcpServer))
	hub := controllers.RegisterHubController(ws, store, mqtt, controllerOpts...)
	if hub == nil {
		return nil, fmt.Errorf("failed to register hub controller")
	}

	return registerApi(port, ws, hub, store, mcpServer, ctx, options...)
}

type optionsConfig struct{ connections *integrations.Service }

// Option composes optional features without expanding AppStore or device test doubles.
type Option func(*optionsConfig)

// WithConnections adds protected connection routes without changing AppStore or device payloads.
func WithConnections(service *integrations.Service) Option {
	return func(c *optionsConfig) { c.connections = service }
}
