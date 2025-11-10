package auth

import (
	"encoding/json"
	"net/http"
)

// LogoutHandler returns a handler that logs out the current user by
// blacklisting the JWT (when valid) and clearing the session cookie.
func LogoutHandler(jwtService *JWTService, blacklist *TokenBlacklist) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var token string
		if c, err := r.Cookie(CookieNameSession); err == nil {
			token = c.Value
		}

		if token == "" {
			writeJSONAuthError(w, http.StatusInternalServerError, "no auth token found")
			return
		}

		// Blacklist the token if found and valid
		if claims, err := jwtService.ValidateJWT(token); err == nil {
			blacklist.Add(token, claims.ExpiresAt.Time)
		}

		clearCookie(w, CookieNameSession)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success"})
	}
}

// MeHandler returns a handler that returns the current authenticated user's
// ID and username, or 401 if unauthorized.
func MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":       user.UserID,
			"username": user.Username,
		})
	}
}

// OfflineStartProxyHandler forwards the offline popup start request
// to the provided offline implementation if it supports HandleOfflineStart.
func OfflineStartProxyHandler(offline OAuth) http.HandlerFunc {
	type offlineStarter interface {
		HandleOfflineStart(http.ResponseWriter, *http.Request)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if h, ok := offline.(offlineStarter); ok {
			h.HandleOfflineStart(w, r)
			return
		}
		writeJSONAuthError(w, http.StatusNotFound, "offline start not available")
	}
}
