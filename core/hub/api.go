package hub

import (
	"context"
	"fmt"
	"net/http"
)

type apiServer struct {
	httpServer http.Server
	ctx        context.Context
	router     *Router
}

func WithContext(ctx context.Context) func(s *apiServer) {
	return func(s *apiServer) { s.ctx = ctx }
}

func WithRouter(router *Router) func(s *apiServer) {
	return func(s *apiServer) { s.router = router }
}

func NewHttpServer(port int, opts ...func(s *apiServer)) *apiServer {

	api := &apiServer{
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

func (s *apiServer) Listen() {

	done := make(chan bool)
	errors := make(chan error)

	defer close(done)
	defer close(errors)

	go func() {
		<-s.ctx.Done()
		if err := s.httpServer.Close(); err != nil {
			errors <- fmt.Errorf("HTTP close error: %v", err)
		}
		fmt.Println("HTTP Server Closed")
		done <- true
	}()

	go func() {
		fmt.Printf("HTTP Server Listening : %s \n", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			errors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	go func() {
		err := <-errors
		fmt.Println("Finished with error:", err.Error())
		done <- true
	}()

	<-done
}

type Server struct {
	api *apiServer
	ctx context.Context
}

func NewServer(port int, ctx context.Context, config MqttConfig, repo Repository) {

	s := Server{ctx: ctx}
	eventHub := NewWsHub()
	mqtt := NewMqttClient(config)

	_, err := NewController(
		WithRepository(repo),
		WithMqtt(mqtt),
		WithEventHub(eventHub))

	if err != nil {
		fmt.Printf("Hub connector error: %s \n", err.Error())
		<-ctx.Done()
	}

	s.api = registerApi(port, ctx, eventHub)
}

func registerApi(port int, ctx context.Context, eventHub EventHub) *apiServer {
	router := NewRouter()
	router.GET("/ws", NewWsHandler(eventHub))
	//router.GET("/", http.FileServer(http.Dir("../../frontend/dist")))

	return NewHttpServer(
		port,
		WithContext(ctx),
		WithRouter(router),
	)
}

func (s *Server) Listen() {
	s.api.Listen()
}
