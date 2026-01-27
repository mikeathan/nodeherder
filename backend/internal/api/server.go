package api

import (
	"context"
	"fmt"
	"net/http"
	"node-herder/utils"
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
		router: &Router{},
	}

	for _, opt := range opts {
		opt(api)
	}

	api.httpServer = http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: api.router,
	}

	return api
}

func (s *ApiServer) Listen() {

	done := make(chan bool)
	errors := make(chan error)


	go func() {
		<-s.ctx.Done()
		if err := s.httpServer.Close(); err != nil {
			select {
			case errors <- fmt.Errorf("HTTP close error: %v", err):
			default:
			}
		}
		utils.LogInfof("HTTP Server Closed")
		select {
		case done <- true:
		default:
		}
	}()

	go func() {
		utils.LogInfof("HTTP Server Listening %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			select {
			case errors <- fmt.Errorf("HTTP server error: %v", err):
			default:
			}
		}
	}()

	go func() {
		err := <-errors
		if err != nil {
			utils.LogErrorf("Finished with error %s", err.Error())
		}
		select {
		case done <- true:
		default:
		}
	}()

	<-done
}
