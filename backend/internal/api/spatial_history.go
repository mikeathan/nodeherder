package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"node-herder/internal/spatialhistory"
)

// SpatialHistorySource is the read side of the spatial recorder.
type SpatialHistorySource interface {
	Sessions(device string, from, to time.Time) ([]spatialhistory.Session, error)
	Points(device, session string, from, to time.Time) ([]spatialhistory.Point, error)
	Heatmap(device string, from, to time.Time, cellMM int) ([]spatialhistory.Cell, error)
}

const (
	maxSpatialSessions = 500
	maxSpatialPoints   = 5000
	maxSpatialCells    = 5000
	maxSpatialDevice   = 256
)

// NewSpatialHistoryHandler serves the authenticated recordings API. One endpoint keeps routing
// and the client contract small: mode=sessions|points|heatmap.
func NewSpatialHistoryHandler(source func() SpatialHistorySource) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		history := source()
		if history == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "Spatial history is unavailable.")
			return
		}
		device := r.URL.Query().Get("device")
		if device == "" || len(device) > maxSpatialDevice || !utf8.ValidString(device) {
			writeJSONError(w, http.StatusBadRequest, "A valid device is required.")
			return
		}
		from, to := historyRange(r)
		mode := r.URL.Query().Get("mode")
		if mode == "" {
			mode = "sessions"
		}
		var payload any
		var err error
		switch mode {
		case "sessions":
			payload, err = history.Sessions(device, from, to)
			if sessions, ok := payload.([]spatialhistory.Session); ok && len(sessions) > maxSpatialSessions {
				payload = sessions[:maxSpatialSessions]
			}
		case "points":
			session := r.URL.Query().Get("session")
			if session == "" {
				writeJSONError(w, http.StatusBadRequest, "A session is required.")
				return
			}
			payload, err = history.Points(device, session, from, to)
			if points, ok := payload.([]spatialhistory.Point); ok && len(points) > maxSpatialPoints {
				payload = points[:maxSpatialPoints]
			}
		case "heatmap":
			cell, _ := strconv.Atoi(r.URL.Query().Get("cell"))
			payload, err = history.Heatmap(device, from, to, cell)
			if cells, ok := payload.([]spatialhistory.Cell); ok && len(cells) > maxSpatialCells {
				payload = cells[:maxSpatialCells]
			}
		default:
			writeJSONError(w, http.StatusBadRequest, "Unknown history mode.")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Spatial history could not be read.")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
}

func historyRange(r *http.Request) (time.Time, time.Time) {
	to := time.Now().UTC()
	from := to.Add(-24 * time.Hour)
	if raw := r.URL.Query().Get("to"); raw != "" {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil && value > 0 {
			to = time.UnixMilli(value).UTC()
		}
	}
	if raw := r.URL.Query().Get("from"); raw != "" {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil && value > 0 {
			from = time.UnixMilli(value).UTC()
		}
	}
	if to.Sub(from) > 366*24*time.Hour {
		from = to.Add(-366 * 24 * time.Hour)
	}
	return from, to
}
