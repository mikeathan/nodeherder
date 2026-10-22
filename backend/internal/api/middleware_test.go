package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLimitRequestBodyPreservesAcceptedInputAndRejectsOversize(t *testing.T) {
	for _, size := range []int{0, MaxRequestBodyBytes, MaxRequestBodyBytes + 1} {
		called := false
		handler := LimitRequestBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			body, err := io.ReadAll(r.Body)
			if err != nil || len(body) != size {
				t.Errorf("body length=%d error=%v", len(body), err)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		request := httptest.NewRequest("POST", "/", strings.NewReader(strings.Repeat("x", size)))
		request.ContentLength = -1 // Also bound chunked/unknown-length requests.
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if size > MaxRequestBodyBytes {
			if called || response.Code != 413 {
				t.Fatalf("oversize called=%v status=%d", called, response.Code)
			}
		} else if !called || response.Code != 204 {
			t.Fatalf("valid body called=%v status=%d", called, response.Code)
		}
	}
}
