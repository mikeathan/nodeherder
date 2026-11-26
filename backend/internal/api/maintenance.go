package api

import (
	"encoding/json"
	"net/http"
	"node-herder/internal/ratelimiter"
	"node-herder/store"
	"node-herder/utils"
	"time"
)

// PurgeMetricsRequest is the explicit scope for a history delete. `Before` is RFC3339 and
// must not be in the future; the handler never accepts a relative "last N days" value.
type PurgeMetricsRequest struct {
	Before string `json:"before"`
}

// PurgeMetricsResponse reports the accepted cutoff so the UI can confirm what was deleted.
type PurgeMetricsResponse struct {
	PurgedBefore string `json:"purgedBefore"`
	PurgedAt     string `json:"purgedAt"`
}

// PurgeMetricsHandler deletes readings older than an explicit cutoff. It is a destructive
// operation: the route is authenticated, the cutoff is validated, and the result is logged.
type PurgeMetricsHandler struct {
	limiter *ratelimiter.WindowRateLimiter
	store   store.AppStore
}

// NewPurgeMetricsHandler builds the history purge handler around the application store.
func NewPurgeMetricsHandler(limiter *ratelimiter.WindowRateLimiter, appStore store.AppStore) *PurgeMetricsHandler {
	return &PurgeMetricsHandler{limiter: limiter, store: appStore}
}

func (h *PurgeMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.limiter.Allow() {
		http.Error(w, "Too many purge requests", http.StatusTooManyRequests)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
		return
	}
	var request PurgeMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	before, err := time.Parse(time.RFC3339, request.Before)
	if err != nil || before.IsZero() {
		http.Error(w, "before must be an RFC3339 timestamp", http.StatusBadRequest)
		return
	}
	now := time.Now()
	if before.After(now) {
		http.Error(w, "before must not be in the future", http.StatusBadRequest)
		return
	}
	if err := h.store.PurgeMetricsBefore(before); err != nil {
		utils.LogErrorf("PurgeMetricsHandler: purge before %s failed: %s", before.Format(time.RFC3339), err.Error())
		http.Error(w, "Could not delete history", http.StatusInternalServerError)
		return
	}
	utils.LogWarnf("PurgeMetricsHandler: deleted readings before %s", before.Format(time.RFC3339))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(PurgeMetricsResponse{
		PurgedBefore: before.UTC().Format(time.RFC3339),
		PurgedAt:     now.UTC().Format(time.RFC3339),
	})
}
