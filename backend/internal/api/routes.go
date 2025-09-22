package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"node-herder/internal/controllers"
	"node-herder/internal/fs"
	"node-herder/internal/ratelimiter"
	"node-herder/internal/ws"
	"node-herder/models/logging"
	"node-herder/store"
	"node-herder/utils"
	"regexp"
	"sync"
	"time"
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
	if handler == nil {
		utils.LogErrorf("Failed to find handler for path: %s", path)
		http.NotFound(w, req)
		return
	}
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

// Web socket
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

// List file logs
type ListFileLogsHandler struct {
	fs fs.FileSystem
}

func NewListFileLogsHandler(fs fs.FileSystem) *ListFileLogsHandler {
	return &ListFileLogsHandler{fs: fs}
}

func (h *ListFileLogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	files, err := h.fs.ListFilesWithExtension(utils.LogsPath, utils.LogExtension)
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

// Log file
type LogFileHandler struct {
	fs fs.FileSystem
}

func NewLogFileHandler(fs fs.FileSystem) *LogFileHandler {
	return &LogFileHandler{fs: fs}
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
	bytes, err := h.fs.Load(request.File)
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

// File
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

// Hub State
type HubStateHandler struct {
	store            store.AppStore
	cachedHubState   []byte
	hubStateIsDirty  bool
	cacheLastUpdated time.Time
	cacheExpiration  time.Duration
	mu               sync.RWMutex
}

func NewHubStateHandler(appStore store.AppStore, cacheExpiration time.Duration) *HubStateHandler {
	sh := &HubStateHandler{
		cachedHubState:  []byte{},
		hubStateIsDirty: false,
		store:           appStore,
		cacheExpiration: cacheExpiration,
		mu:              sync.RWMutex{},
	}

	sh.store.RegisterIsDirtyCallback(func() {
		sh.setDirty()
	})
	return sh
}

func (h *HubStateHandler) setDirty() {
	defer h.mu.Unlock()
	h.mu.Lock()
	h.hubStateIsDirty = true

}

func (h *HubStateHandler) isStaleOrDirty() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.hubStateIsDirty || time.Since(h.cacheLastUpdated) > h.cacheExpiration || len(h.cachedHubState) == 0
}

func (h *HubStateHandler) refreshCacheIfNeeded() error {
	if !h.isStaleOrDirty() {
		return nil
	}

	state, err := h.store.LoadHubState()
	if err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	h.mu.Lock()
	h.cachedHubState = data
	h.cacheLastUpdated = time.Now()
	h.hubStateIsDirty = false
	h.mu.Unlock()

	return nil
}

func (h *HubStateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := h.refreshCacheIfNeeded(); err != nil {
		utils.LogErrorf("HubStateHandler: Failed to load hub state %s", err.Error())
		http.Error(w, "Failed to load hub state", http.StatusInternalServerError)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	w.Write(h.cachedHubState)
}

// Automation Trigger
type AutomationTriggerHandler struct {
	hub       *controllers.HubController
	limiter   *ratelimiter.RateLimiter
	rateLimit time.Duration
}

func NewAutomationTriggerHandler(hub *controllers.HubController, rateLimit time.Duration) *AutomationTriggerHandler {
	sh := &AutomationTriggerHandler{
		hub:       hub,
		limiter:   ratelimiter.NewRateLimiter(),
		rateLimit: rateLimit,
	}
	return sh
}

func (h *AutomationTriggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	automationId := r.URL.Query().Get("automationId")
	triggerName := r.URL.Query().Get("triggerName")

	if automationId == "" || triggerName == "" {
		http.Error(w, "missing automationId or triggerName", http.StatusBadRequest)
		return
	}

	if !h.limiter.AllowWrite(automationId, h.rateLimit) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	err := h.hub.TriggerManual(automationId, triggerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
