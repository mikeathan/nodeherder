package hub

import (
	"context"
	"fmt"
	"net/http"
	"node-herder/devices"
)

type apiServer struct {
	httpServer http.Server
	ctx        context.Context
	router     *Router
}

func withContext(ctx context.Context) func(s *apiServer) {
	return func(s *apiServer) { s.ctx = ctx }
}

func withRouter(router *Router) func(s *apiServer) {
	return func(s *apiServer) { s.router = router }
}

func newHttpServer(port int, opts ...func(s *apiServer)) *apiServer {

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
	api        *apiServer
	controller *Controller
	ctx        context.Context
}

func NewServer(port int, ctx context.Context, config MqttConfig, repo devices.Repository) *Server {

	s := &Server{ctx: ctx}
	eventHub := NewWsHub()
	mqtt := NewMqttClient(config)

	s.controller = NewController(
		WithRepository(repo),
		WithMqtt(mqtt),
		WithEventHub(eventHub),
		WithContext(ctx))

	s.api = registerApiServer(port, ctx, eventHub)
	return s
}

func registerApiServer(port int, ctx context.Context, eventHub EventHub) *apiServer {
	router := NewRouter()
	router.GET("/ws", NewWsHandler(eventHub))
	//router.GET("/", http.FileServer(http.Dir("../../frontend/dist")))

	return newHttpServer(
		port,
		withContext(ctx),
		withRouter(router),
	)
}

func (s *Server) Listen() {

	err := s.controller.Connect()

	if err != nil {
		fmt.Printf("Controller connection error: %s \n", err.Error())
		<-s.ctx.Done()
	}

	s.api.Listen()
}
