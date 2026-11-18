package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	model "node-herder/models/observations"
)

// historySourceStub serves canned evidence for the route tests.
type historySourceStub struct {
	summaries   []model.DailySummary
	transitions []model.Transition
	err         error
	lastKey     string
	lastLimit   int
}

func (s *historySourceStub) Summaries(key string, limit int) ([]model.DailySummary, error) {
	s.lastKey, s.lastLimit = key, limit
	if s.err != nil {
		return nil, s.err
	}
	return s.summaries, nil
}

func (s *historySourceStub) Transitions(string, int) ([]model.Transition, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.transitions, nil
}

// historyResponse mirrors the handler's payload for decoding in tests.
type historyResponse struct {
	Summaries   []model.DailySummary `json:"summaries"`
	Transitions []model.Transition   `json:"transitions"`
}

func serveHistory(t *testing.T, source ObservationHistorySource, query string) *httptest.ResponseRecorder {
	t.Helper()
	handler := NewObservationHistoryHandler(func() ObservationHistorySource { return source })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/observations/history"+query, nil))
	return recorder
}

func TestObservationHistoryReturnsEvidenceAndTransitions(t *testing.T) {
	// Both halves travel together: a class timeline without the changes that produced it cannot be
	// explained, so the UI would need two round trips to say anything.
	source := &historySourceStub{
		summaries: []model.DailySummary{{Key: "addr:aa:bb:cc:dd:ee:01", Day: "2026-11-13", Packets: 100}},
		transitions: []model.Transition{{
			Key: "addr:aa:bb:cc:dd:ee:01", Day: "2026-11-13", From: "personal", To: "infrastructure",
		}},
	}

	recorder := serveHistory(t, source, "?key=addr:aa:bb:cc:dd:ee:01")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var payload historyResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload.Summaries) != 1 || len(payload.Transitions) != 1 {
		t.Errorf("got %d summaries and %d transitions, want 1 and 1",
			len(payload.Summaries), len(payload.Transitions))
	}
}

func TestObservationHistoryRequiresAKey(t *testing.T) {
	// Without a key the route would be a bulk export of every device's activity, which nothing in
	// the UI needs.
	recorder := serveHistory(t, &historySourceStub{}, "")

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", recorder.Code)
	}
}

func TestObservationHistoryRejectsAnOversizedKey(t *testing.T) {
	recorder := serveHistory(t, &historySourceStub{}, "?key="+strings.Repeat("a", 300))

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", recorder.Code)
	}
}

func TestObservationHistoryBoundsTheLimit(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "default", query: "?key=k", want: maxObservationHistoryRows},
		{name: "smaller is honoured", query: "?key=k&limit=5", want: 5},
		{name: "larger is capped", query: "?key=k&limit=999", want: maxObservationHistoryRows},
		{name: "zero falls back", query: "?key=k&limit=0", want: maxObservationHistoryRows},
		{name: "negative falls back", query: "?key=k&limit=-3", want: maxObservationHistoryRows},
		{name: "garbage falls back", query: "?key=k&limit=many", want: maxObservationHistoryRows},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := &historySourceStub{}
			recorder := serveHistory(t, source, test.query)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", recorder.Code)
			}
			if source.lastLimit != test.want {
				t.Errorf("limit = %d, want %d", source.lastLimit, test.want)
			}
		})
	}
}

func TestObservationHistoryUnavailableWhenTheStoreIsMissing(t *testing.T) {
	// A nil source means the store failed to open. The route must say so rather than pretend the
	// identity has no history, which would look like data loss.
	recorder := serveHistory(t, nil, "?key=k")

	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", recorder.Code)
	}
}

func TestObservationHistoryReportsAReadFailure(t *testing.T) {
	recorder := serveHistory(t, &historySourceStub{err: errors.New("disk")}, "?key=k")

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", recorder.Code)
	}
}

func TestObservationHistoryIsEmptyRatherThanAbsent(t *testing.T) {
	// No evidence is a normal state, not an error: a device seen for the first time has no history
	// yet, and the UI must be able to render that without special-casing a 404.
	source := &historySourceStub{}

	recorder := serveHistory(t, source, "?key=k")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var payload historyResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload.Summaries) != 0 || len(payload.Transitions) != 0 {
		t.Errorf("got %d summaries, want none", len(payload.Summaries))
	}
}
