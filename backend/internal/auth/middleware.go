package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const userKey contextKey = "user"

func Auth(jwtService *JWTService, blacklist *TokenBlacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			// FIRST: Try to get token from cookie (preferred for web clients)
			if c, err := r.Cookie(AuthCookie); err == nil {
				token = c.Value
			}

			// FALLBACK: Check Authorization header (for API clients/backward compatibility)
			if token == "" {
				token = extractBearer(r.Header.Get("Authorization"))
			}

			if token == "" {
				writeJSONAuthError(w, http.StatusUnauthorized, "missing token")
				return
			}

			// Check if token is blacklisted
			if blacklist != nil && blacklist.IsBlacklisted(token) {
				writeJSONAuthError(w, http.StatusUnauthorized, "token revoked")
				return
			}

			user, err := jwtService.ValidateJWT(token)
			if err != nil {
				writeJSONAuthError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearer(h string) string {
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	// allow passing plain token too
	return h
}

func writeJSONAuthError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func GetUserFromContext(r *http.Request) (*Claims, bool) {
	user, ok := r.Context().Value(userKey).(*Claims)
	return user, ok
}
