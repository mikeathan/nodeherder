package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"node-herder/internal/api"
	"node-herder/internal/spatialhistory"
)

type fakeSpatialHistory struct{ calls int }

func (f *fakeSpatialHistory) Sessions(device string, from, to time.Time) ([]spatialhistory.Session, error) {
	f.calls++
	if device != "radar" {
		return nil, nil
	}
	return []spatialhistory.Session{{ID: "s1", From: from.UnixMilli(), To: to.UnixMilli(), Points: 2, Slots: []int{1}, Samples: []int64{from.UnixMilli()}}}, nil
}
func (f *fakeSpatialHistory) Points(string, string, time.Time, time.Time) ([]spatialhistory.Point, error) {
	return []spatialhistory.Point{{At: 1, Slot: 1, X: 100, Y: 200}}, nil
}
func (f *fakeSpatialHistory) Heatmap(string, time.Time, time.Time, int) ([]spatialhistory.Cell, error) {
	return []spatialhistory.Cell{{X: 0, Y: 0, Count: 2}}, nil
}

func TestSpatialHistoryHandlerRequiresDeviceAndServesModes(t *testing.T) {
	source := &fakeSpatialHistory{}
	handler := api.NewSpatialHistoryHandler(func() api.SpatialHistorySource { return source })

	missing := httptest.NewRecorder()
	handler.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/api/spatial/history", nil))
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing device status = %d", missing.Code)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/spatial/history?device=radar&mode=sessions&from=1&to=2", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	// Decode into the wire shape, not the domain struct: decoding into spatialhistory.Session
	// would pass even if the `samples` json tag were renamed or dropped, because the field would
	// simply stay empty. The key must be present and carry the additive timestamps.
	var sessions []map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &sessions); err != nil || len(sessions) != 1 {
		t.Fatalf("sessions = %s err=%v", response.Body.String(), err)
	}
	var id string
	if err := json.Unmarshal(sessions[0]["id"], &id); err != nil || id != "s1" {
		t.Fatalf("id = %q err=%v", id, err)
	}
	rawSamples, ok := sessions[0]["samples"]
	if !ok {
		t.Fatalf("samples key missing from wire payload: %s", response.Body.String())
	}
	var samples []int64
	if err := json.Unmarshal(rawSamples, &samples); err != nil || len(samples) != 1 || samples[0] != 1 {
		t.Fatalf("samples = %s err=%v", rawSamples, err)
	}

	for _, mode := range []string{"points", "heatmap"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/spatial/history?device=radar&mode="+mode+"&session=s1&from=1&to=2", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d", mode, recorder.Code)
		}
	}
}
