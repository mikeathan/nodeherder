package api

import (
	"net/http"

	"node-herder/utils"
)

func CORS(next http.Handler) http.Handler {
	allowedOrigins := allowedOriginsMap()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		w.Header().Set("Vary", "Origin")

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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

	utils.LogInfof("CORS: Allowing origins: %v", allowedOriginUrls)
	return allowedOrigins
}
