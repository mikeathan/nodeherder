package api

// The observation history endpoint serves the stored evidence behind a class, so an operator can
// inspect a device's days rather than trust a bare label.
//
// It is deliberately narrow: one identity at a time, two modes (summaries and transitions), and a
// hard row ceiling. Browsing every identity's history at once would be a bulk export of the
// household's device activity, which nothing in the UI needs.

import (
	"encoding/json"
	"net/http"
	"strconv"

	model "node-herder/models/observations"
)

// ObservationHistorySource is the read side of the summary store.
type ObservationHistorySource interface {
	Summaries(key string, limit int) ([]model.DailySummary, error)
	Transitions(key string, limit int) ([]model.Transition, error)
}

const (
	// maxObservationHistoryRows matches the store's own ceiling, so the handler cannot ask for
	// more than the store will return and the contract is stated in one place.
	maxObservationHistoryRows = 30

	// maxObservationKeyLength bounds the identity parameter. Real keys are 17-40 characters; 256 is
	// generous while still refusing an unbounded string.
	maxObservationKeyLength = 256
)

// NewObservationHistoryHandler serves stored evidence for one identity.
//
// The identity key is a query parameter rather than a path segment because keys are opaque and
// contain characters that would need escaping. It is validated as non-empty and bounded, but not
// parsed: the key's structure is the resolver's business, and rejecting an unfamiliar shape here
// would break a future medium that keys differently.
func NewObservationHistoryHandler(source func() ObservationHistorySource) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		history := source()
		if history == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "Observation history is unavailable.")
			return
		}

		key := r.URL.Query().Get("key")
		if key == "" || len(key) > maxObservationKeyLength {
			writeJSONError(w, http.StatusBadRequest, "A valid identity key is required.")
			return
		}

		limit := maxObservationHistoryRows
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if value, err := strconv.Atoi(raw); err == nil && value > 0 && value < limit {
				limit = value
			}
		}

		payload := struct {
			Summaries   []model.DailySummary `json:"summaries"`
			Transitions []model.Transition   `json:"transitions"`
		}{}

		summaries, err := history.Summaries(key, limit)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Observation history could not be read.")
			return
		}
		payload.Summaries = summaries

		// Transitions are included in the same response because the UI always needs both: a class
		// timeline without the changes that produced it cannot be explained.
		transitions, err := history.Transitions(key, limit)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Observation history could not be read.")
			return
		}
		payload.Transitions = transitions

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
}
