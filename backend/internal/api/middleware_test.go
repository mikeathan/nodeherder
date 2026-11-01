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

func TestCORSESPHomePreflight(t *testing.T) {
	t.Setenv("FRONTEND_ALLOWED_ORIGINS", "https://ui.example.test")
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	for _, path := range []string{esphomeDevicesPath, esphomeDevicesPath + "/fixture", esphomeDiscoveryPath} {
		for _, origin := range []string{"https://ui.example.test", "https://evil.example.test"} {
			req := httptest.NewRequest(http.MethodOptions, path, nil)
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", "POST")
			req.Header.Set("Access-Control-Request-Headers", "content-type,x-nodeherder-request")
			result := httptest.NewRecorder()
			handler.ServeHTTP(result, req)
			allowed := strings.Contains(result.Header().Get("Access-Control-Allow-Headers"), "X-NodeHerder-Request")
			if allowed != (origin == "https://ui.example.test") {
				t.Errorf("path %s origin %s headers %v", path, origin, result.Header())
			}
		}
	}
	for _, path := range []string{"/api/esphome-unrelated", esphomeDevicesPath + "-unrelated", esphomeDiscoveryPath + "/unrelated"} {
		req := httptest.NewRequest(http.MethodOptions, path, nil)
		req.Header.Set("Origin", "https://ui.example.test")
		result := httptest.NewRecorder()
		handler.ServeHTTP(result, req)
		if strings.Contains(result.Header().Get("Access-Control-Allow-Headers"), "X-NodeHerder-Request") {
			t.Errorf("unrelated route %q was broadened", path)
		}
	}
}
