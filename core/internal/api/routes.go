package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"node-herder/internal/controllers"
	"node-herder/internal/ws"
	"node-herder/models/logging"
	"node-herder/utils"
	"regexp"
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

func (h *WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.hub.HandleRequest(w, r)
	if err != nil {
		utils.LogErrorf("WsHandler: HandleRequest error %s", err.Error())
		return
	}

	utils.LogInfo("WsHandler: client connected")
}

type ListFileLogsHandler struct {
	walker utils.Walker
}

type FileSystemService interface {
	ListFiles(path string) ([]string, error)
	Load(file string) ([]byte, error)
}

TODO use filesystemservice

func NewListFileLogsHandler(walker utils.Walker) *ListFileLogsHandler {
	return &ListFileLogsHandler{walker: walker}
}

func (h *ListFileLogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	files, err := utils.ListFileLogs(h.walker)
	if err != nil {
		utils.LogErrorf("ListFileLogsHandler: ListFileLogs error %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(files); err != nil {
		utils.LogErrorf("ListFileLogsHandler: Failed to encode response %s", err.Error())

		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type LogFileHandler struct {
	loader utils.Loader
}

func NewLogFileHandler(loader utils.Loader) *LogFileHandler {
	return &LogFileHandler{loader: loader}
}

func (h *LogFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body := buf.String()
	request := logging.FileLogRequest{}
	if err = json.Unmarshal([]byte(body), &request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.Action != logging.LoadAction {
		http.Error(w, "Error unknown request action type", http.StatusBadRequest)
		return
	}

	// we only support loading files for now
	bytes, err := h.loader.Load(request.File)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err = w.Write(bytes); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}

type DataCollectorHandler struct {
	hub *controllers.HubController
}

func NewDataCollectorHandler(hub *controllers.HubController) *DataCollectorHandler {
	return &DataCollectorHandler{
		hub: hub,
	}
}

func (h *DataCollectorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body := buf.String()
	//unquote, _ := strconv.Unquote(body)
	var payload map[string]interface{}
	if err = json.Unmarshal([]byte(body), &payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, payload, err := utils.ParsePayload(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.hub.Enqueue(id, payload, "http")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}

type FileHandler struct {
	handlerFunc func(http.ResponseWriter, *http.Request)
}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handlerFunc(w, r)
}

func NewFileHandler(path string, redirectPath string) *FileHandler {
	fh := &FileHandler{}

	fh.handlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, redirectPath)
	})

	return fh
}
