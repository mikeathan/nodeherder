package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"node-herder/internal/auth"
	"node-herder/internal/controllers"
	"node-herder/internal/fs"
	metricsquery "node-herder/internal/metrics/query"
	metrics "node-herder/internal/metrics/services"
	"node-herder/internal/ratelimiter"
	"node-herder/internal/ws"
	"node-herder/models/assistant"
	"node-herder/models/automations"
	"node-herder/models/logging"
	"node-herder/store"
	"node-herder/utils"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const ParametirsedSuffix = ":"

type contextKey int

const paramsKey contextKey = iota

// Param extracts a named route parameter from the request context.
func Param(r *http.Request, key string) string {
	if params, ok := r.Context().Value(paramsKey).(map[string]string); ok {
		return params[key]
	}
	return ""
}

type segment struct {
	value   string
	isParam bool
}

type route struct {
	method  string
	handler http.Handler
	public  bool
}

type paramRoute struct {
	route
	segments []segment
}

type Router struct {
	static               map[string]route // "METHOD /path" -> route
	knownPaths           map[string]bool  // tracks all registered paths for OPTIONS
	paramRoutes          []paramRoute
	protectedMiddlewares []func(http.Handler) http.Handler
	globalMiddlewares    []func(http.Handler) http.Handler
}

func NewRouter() *Router {
	return &Router{
		static:               make(map[string]route),
		knownPaths:           make(map[string]bool),
		paramRoutes:          []paramRoute{},
		protectedMiddlewares: []func(http.Handler) http.Handler{},
		globalMiddlewares:    []func(http.Handler) http.Handler{},
	}
}

func (r *Router) UseProtected(fn func(http.Handler) http.Handler) {
	r.protectedMiddlewares = append(r.protectedMiddlewares, fn)
}

func (r *Router) UseGlobal(fn func(http.Handler) http.Handler) {
	r.globalMiddlewares = append(r.globalMiddlewares, fn)
}

func (r *Router) PublicGET(path string, handler http.Handler) {
	r.addRoute(http.MethodGet, path, handler, true)
}

func (r *Router) PublicPOST(path string, handler http.Handler) {
	r.addRoute(http.MethodPost, path, handler, true)
}

func (r *Router) GET(path string, handler http.Handler) {
	r.addRoute(http.MethodGet, path, handler, false)
}

func (r *Router) POST(path string, handler http.Handler) {
	r.addRoute(http.MethodPost, path, handler, false)
}

func (r *Router) PUT(path string, handler http.Handler) {
	r.addRoute(http.MethodPut, path, handler, false)
}

func (r *Router) DELETE(path string, handler http.Handler) {
	r.addRoute(http.MethodDelete, path, handler, false)
}

func routeKey(method, path string) string {
	return fmt.Sprintf("%s %s", method, path)
}

func splitPath(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

func (r *Router) addRoute(method, path string, handler http.Handler, public bool) {
	rt := route{method: method, handler: handler, public: public}
	r.knownPaths[path] = true

	if !strings.Contains(path, ParametirsedSuffix) {
		r.static[routeKey(method, path)] = rt
		return
	}

	parts := splitPath(path)
	segments := make([]segment, len(parts))
	for i, p := range parts {
		if strings.HasPrefix(p, ParametirsedSuffix) {
			segments[i] = segment{value: p[1:], isParam: true}
		} else {
			segments[i] = segment{value: p}
		}
	}

	r.paramRoutes = append(r.paramRoutes, paramRoute{route: rt, segments: segments})
}

func (r *Router) AddAuthentication(provider auth.AuthProvider) {
	provider.RegisterRoutes(r)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, params := r.match(req.Method, req.URL.Path)

	if len(params) > 0 {
		// add params to context
		ctx := context.WithValue(req.Context(), paramsKey, params)
		req = req.WithContext(ctx)
	}

	for _, mw := range r.globalMiddlewares {
		handler = mw(handler)
	}

	handler.ServeHTTP(w, req)
}

var noopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func (r *Router) match(method, path string) (http.Handler, map[string]string) {
	// Fast path: static route lookup
	if rt, ok := r.static[routeKey(method, path)]; ok {
		return r.applyProtected(rt), nil
	}

	// Static OPTIONS fallback: just return noop so CORS can work
	if method == http.MethodOptions && r.knownPaths[path] {
		return noopHandler, nil
	}

	// Slow path: parameterized routes
	handler, params := r.matchParam(method, path)

	// Parameterized OPTIONS fallback: skip auth
	if method == http.MethodOptions && handler != nil {
		return noopHandler, params
	}

	if handler == nil {
		return http.NotFoundHandler(), nil
	}

	return handler, params
}

func (r *Router) matchParam(method, path string) (http.Handler, map[string]string) {
	pathParts := splitPath(path)
	for _, pr := range r.paramRoutes {
		// Method must match (skip check for OPTIONS preflight)
		if method != http.MethodOptions && pr.method != method {
			continue
		}

		if len(pr.segments) != len(pathParts) {
			continue
		}

		params := make(map[string]string)
		matched := true
		for i, seg := range pr.segments {
			if seg.isParam {
				params[seg.value] = pathParts[i]
			} else if seg.value != pathParts[i] {
				matched = false
				break
			}
		}

		if matched {
			return r.applyProtected(pr.route), params
		}
	}

	return nil, nil
}

func (r *Router) applyProtected(rt route) http.Handler {
	handler := rt.handler
	if !rt.public {
		for _, mw := range r.protectedMiddlewares {
			handler = mw(handler)
		}
	}
	return handler
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
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(files); err != nil {
		utils.LogErrorf("ListFileLogsHandler: Failed to encode response %s", err.Error())
		writeJSONError(w, http.StatusInternalServerError, "Failed to encode response")
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

	cleanPath := filepath.Clean(request.File)
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "..") || filepath.Ext(cleanPath) != utils.LogExtension {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// we only support loading files for now
	bytes, err := h.fs.Load(cleanPath)
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
		writeJSONError(w, http.StatusInternalServerError, "Failed to load hub state")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	w.Write(h.cachedHubState)
}

// Device Context Handler
type DeviceContextHandler struct {
	store     store.AppStore
	limiter   *ratelimiter.DeviceRateLimiter
	rateLimit time.Duration
}

func NewDeviceContextHandler(store store.AppStore, rateLimit time.Duration) *DeviceContextHandler {
	return &DeviceContextHandler{
		store:     store,
		limiter:   ratelimiter.NewDeviceRateLimiter(),
		rateLimit: rateLimit,
	}
}

func (h *DeviceContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if !h.limiter.AllowWrite("DeviceContext", h.rateLimit) {
		writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	state, err := h.store.LoadHubState()
	if err != nil {
		utils.LogErrorf("DeviceContextHandler: Failed to load hub state %s", err.Error())
		writeJSONError(w, http.StatusInternalServerError, "Failed to load hub state")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := CreateDeviceContextResponse(state.Devices)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.LogErrorf("DeviceContextHandler: Failed to encode response %s", err.Error())
		writeJSONError(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
}

// Automation Trigger
type AutomationTriggerHandler struct {
	hub       automations.AutomationTrigger
	limiter   *ratelimiter.DeviceRateLimiter
	rateLimit time.Duration
}

func NewAutomationTriggerHandler(hub automations.AutomationTrigger, rateLimit time.Duration) *AutomationTriggerHandler {
	sh := &AutomationTriggerHandler{
		hub:       hub,
		limiter:   ratelimiter.NewDeviceRateLimiter(),
		rateLimit: rateLimit,
	}
	return sh
}

func (h *AutomationTriggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		writeJSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	// Read body
	var payload struct {
		AutomationId string `json:"automationId"`
		TriggerName  string `json:"triggerName"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	automationId := payload.AutomationId
	triggerName := payload.TriggerName
	if automationId == "" || triggerName == "" {
		writeJSONError(w, http.StatusBadRequest, "missing automationId or triggerName")
		return
	}

	if !h.limiter.AllowWrite(automationId, h.rateLimit) {
		writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	err = h.hub.TriggerManual(automationId, triggerName)
	if err != nil {
		writeJSONError(w, http.StatusPreconditionFailed, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Metrics query
type MetricsQueryHandler struct {
	limiter   *ratelimiter.WindowRateLimiter
	rateLimit time.Duration
	querier   *metrics.QueryService
}

func NewMetricsQueryHandler(limiter *ratelimiter.WindowRateLimiter, store store.AppStore) *MetricsQueryHandler {
	return &MetricsQueryHandler{
		limiter: limiter,
		querier: metrics.NewQueryService(store),
	}
}

func (h *MetricsQueryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		writeJSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var req metricsquery.MetricsQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if len(req.Exposes) == 0 || len(req.DeviceIds) == 0 {
		writeJSONError(w, http.StatusBadRequest, "missing deviceIds or expose")
		return
	}

	if !h.limiter.Allow() {
		writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	response, err := h.querier.Query(r.Context(), req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Assistant Message Proxy
type AssistantMessageHandler struct {
	store store.AppStore
}

func NewAssistantMessageHandler(store store.AppStore) *AssistantMessageHandler {
	return &AssistantMessageHandler{
		store: store,
	}
}

func (h *AssistantMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		writeJSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	config, err := h.store.AppConfig().LoadAppConfig()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load config")
		return
	}

	url := config.Hub.Assistant.Url
	if url == "" {
		writeJSONError(w, http.StatusBadRequest, "Assistant URL not configured")
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	var reqPayload struct {
		ConversationID string `json:"conversation_id"`
		Message        string `json:"message"`
	}
	if err := json.Unmarshal(bodyBytes, &reqPayload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Store user message
	if reqPayload.ConversationID != "" && reqPayload.Message != "" {
		if err := h.store.AppendAssistantMessage(reqPayload.ConversationID, assistant.RoleUser, reqPayload.Message); err != nil {
			utils.LogErrorf("AssistantMessageHandler: failed to save user message: %v", err)
		}
	}

	// Proxy to LLM
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "Failed to connect to assistant: "+err.Error())
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "Failed to read assistant response")
		return
	}

	// Store assistant reply
	if reqPayload.ConversationID != "" && resp.StatusCode == http.StatusOK {
		var respPayload struct {
			Reply string `json:"reply"`
		}
		if err := json.Unmarshal(respBytes, &respPayload); err == nil && respPayload.Reply != "" {
			if err := h.store.AppendAssistantMessage(reqPayload.ConversationID, assistant.RoleAssistant, respPayload.Reply); err != nil {
				utils.LogErrorf("AssistantMessageHandler: failed to save assistant reply: %v", err)
			}
		}
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	w.Write(respBytes)
}

// AssistantConversationsHandler handles GET /api/assistant/conversations
type AssistantConversationsHandler struct {
	store store.AppStore
}

func NewAssistantConversationsHandler(store store.AppStore) *AssistantConversationsHandler {
	return &AssistantConversationsHandler{store: store}
}

func (h *AssistantConversationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conversations, err := h.store.ListAssistantConversations()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list conversations")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"conversations": conversations})
}

// AssistantHistoryHandler handles GET /api/assistant/history/{id}
type AssistantHistoryHandler struct {
	store store.AppStore
}

func NewAssistantHistoryHandler(store store.AppStore) *AssistantHistoryHandler {
	return &AssistantHistoryHandler{store: store}
}

func (h *AssistantHistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conversationID := Param(r, "id")
	if conversationID == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing conversation ID")
		return
	}

	history, err := h.store.LoadAssistantHistory(conversationID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Conversation not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(history)
}

// AssistantDeleteHandler handles DELETE /api/assistant/history/{id}
type AssistantDeleteHandler struct {
	store store.AppStore
}

func NewAssistantDeleteHandler(store store.AppStore) *AssistantDeleteHandler {
	return &AssistantDeleteHandler{store: store}
}

func (h *AssistantDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conversationID := Param(r, "id")
	if conversationID == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing conversation ID")
		return
	}

	if err := h.store.DeleteAssistantConversation(conversationID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete conversation")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
