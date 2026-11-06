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
	return []spatialhistory.Session{{ID: "s1", From: from.UnixMilli(), To: to.UnixMilli(), Points: 2, Slots: []int{1}}}, nil
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
	var sessions []spatialhistory.Session
	if err := json.Unmarshal(response.Body.Bytes(), &sessions); err != nil || len(sessions) != 1 || sessions[0].ID != "s1" {
		t.Fatalf("sessions = %+v err=%v", sessions, err)
	}

	for _, mode := range []string{"points", "heatmap"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/spatial/history?device=radar&mode="+mode+"&session=s1&from=1&to=2", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d", mode, recorder.Code)
		}
	}
}
