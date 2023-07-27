package routes

import (
	"fmt"
	"log"
	"net/http"
	"node-herder/transport/ws"
	"regexp"

	"github.com/gorilla/websocket"
)

type Route struct {
	pattern string
	method  string
	handler http.Handler
}

type Router struct {
	routes      []*Route
	middlewares []func(http.Handler) http.Handler
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Use(fn func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, fn)
}

func (r *Router) GET(path string, handler http.Handler) {
	r.addRoute(http.MethodGet, path, handler)
}

func (r *Router) POST(path string, handler http.Handler) {
	r.addRoute(http.MethodPost, path, handler)
}

func (r *Router) PUT(path string, handler http.Handler) {
	r.addRoute(http.MethodPut, path, handler)
}

func (r *Router) DELETE(path string, handler http.Handler) {
	r.addRoute(http.MethodDelete, path, handler)
}

func (r *Router) addRoute(method string, path string, handler http.Handler) {
	r.routes = append(r.routes, &Route{method: method, pattern: path, handler: handler})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	handler := r.getHandler(method, path)

	handler.ServeHTTP(w, req)
}

func (r *Router) getHandler(method, path string) http.Handler {
	for _, route := range r.routes {
		re := regexp.MustCompile(route.pattern)
		if route.method == method && re.MatchString(path) {

			handler := route.handler

			// chain handler with middleware
			for _, mw := range r.middlewares {
				handler = mw(handler)
			}

			return handler
		}
	}

	return http.NotFoundHandler()
}

type WsHandler struct {
	hub ws.EventHub
}

func NewWsHandler(hub ws.EventHub) *WsHandler {
	return &WsHandler{
		hub: hub,
	}
}

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // for debug only ??
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			http.Error(w, reason.Error(), status)
		},
	}
)

func (h *WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("websocketUpgrader error :", err)
		return
	}

	h.hub.RegisterNewClient(conn)

	fmt.Printf("WsHandler: client connected\n")
}
