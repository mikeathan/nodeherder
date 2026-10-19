package api

import (
	"encoding/json"
	"net/http"
	"node-herder/internal/metrics/query"
	"node-herder/internal/metrics/services"
)

func (h *MetricsQueryHandler) serveHistory(w http.ResponseWriter, r *http.Request, req query.MetricsQueryRequest) {
	if err := services.ValidateHistory(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := h.querier.QueryHistory(r.Context(), req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "History could not be read. Narrow the range or try again.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
