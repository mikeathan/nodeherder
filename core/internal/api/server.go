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

// TODO:
// r.HandleFunc("/", index)

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "build/index.html")
}

func Redirect(routePath string, filePath string) {

}
func (s *ApiServer) Listen() {

	done := make(chan bool)
	errors := make(chan error)

	defer close(done)
	defer close(errors)

	go func() {
		<-s.ctx.Done()
		if err := s.httpServer.Close(); err != nil {
			errors <- fmt.Errorf("HTTP close error: %v", err)
		}
		utils.LogInfof("HTTP Server Closed")
		done <- true
	}()

	go func() {
		utils.LogInfof("HTTP Server Listening %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			errors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	go func() {
		err := <-errors
		utils.LogErrorf("Finished with error %s", err.Error())
		done <- true
	}()

	<-done
}
