package api

import (
	"fmt"
	"log"
	"net/http"
	"node-herder/utils"
	"strings"
)

func CORS(next http.Handler) http.Handler {
	frontendURL, err := utils.GetFrontendBaseURL()
	if err != nil {
		log.Fatalf("CORS middleware failed to get frontend base URL: %v", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		w.Header().Set("Vary", "Origin")

		fmt.Printf("CORS Middleware - Origin: %v Frontend URL: %v\n", origin, frontendURL)
		// Always send CORS headers for frontend origin
		if strings.EqualFold(origin, frontendURL) {
			w.Header().Set("Access-Control-Allow-Origin", frontendURL)
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
