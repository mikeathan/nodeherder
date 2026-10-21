package api

import (
	"context"
	"fmt"
	"net/http"
	"node-herder/utils"
	"time"
)

type ApiServer struct {
	httpServer http.Server
	ctx        context.Context
	router     *Router
}

func WithContext(ctx context.Context) func(s *ApiServer) {
	return func(s *ApiServer) { s.ctx = ctx }
}

func WithRouter(router *Router) func(s *ApiServer) {
	return func(s *ApiServer) { s.router = router }
}

func NewHttpServer(port int, opts ...func(s *ApiServer)) *ApiServer {

	api := &ApiServer{
		router: NewRouter(),
		ctx:    context.Background(),
	}

	for _, opt := range opts {
		opt(api)
	}

	api.httpServer = http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           api.router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return api
}

func (s *ApiServer) Listen() {
	stop := context.AfterFunc(s.ctx, func() { _ = s.httpServer.Close() })
	defer stop()
	utils.LogInfof("HTTP Server Listening %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.LogErrorf("HTTP server failed: %v", err)
	}
}
