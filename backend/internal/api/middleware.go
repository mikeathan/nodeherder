package api

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"node-herder/utils"
)

// MaxRequestBodyBytes bounds JSON commands, collection reports and imports.
const MaxRequestBodyBytes = 1 << 20

// LimitRequestBody rejects oversized input before any handler changes domain state.
func LimitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.Body == http.NoBody {
			next.ServeHTTP(w, r)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxRequestBodyBytes))
		_ = r.Body.Close()
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				writeJSONError(w, http.StatusRequestEntityTooLarge, "request body exceeds 1 MiB")
			} else {
				writeJSONError(w, http.StatusBadRequest, "failed to read request body")
			}
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}

func CORS(next http.Handler) http.Handler {
	allowedOrigins := allowedOriginsMap()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		w.Header().Set("Vary", "Origin")

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			headers := "Content-Type, Authorization, Cache-Control, Connection"
			if integrationPath(r.URL.Path) {
				headers += ", X-NodeHerder-Request, If-Match, If-None-Match"
				w.Header().Set("Access-Control-Expose-Headers", "ETag, Retry-After, Location")
			}
			w.Header().Set("Access-Control-Allow-Headers", headers)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		} else if origin != "" {
			utils.LogWarnf("CORS: Blocked request from origin: %s", origin)
		}

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func allowedOriginsMap() map[string]bool {
	allowedOriginUrls := utils.GetFrontendAllowedOrigins()
	allowedOrigins := make(map[string]bool, len(allowedOriginUrls))
	for _, url := range allowedOriginUrls {
		allowedOrigins[url] = true
	}

	return allowedOrigins
}
