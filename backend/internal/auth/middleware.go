package auth

import (
	"context"
	"net/http"
)

type contextKey string

const userKey contextKey = "user"

func Auth(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// exlude public path
			// to refactor

			needs fixing
			https://chatgpt.com/c/68f4d89d-3aa8-8330-8aea-9ac471ab86e6
			path := r.URL.Path
			if path == "/api/auth/login" ||
				path == "/api/auth/callback" ||
				path == "/api/auth/logout" ||
				path == "/api/auth/me" {
				next.ServeHTTP(w, r)
				return
			}
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			user, err := jwtService.ValidateJWT(authHeader)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(r *http.Request) (*Claims, bool) {
	user, ok := r.Context().Value(userKey).(*Claims)
	return user, ok
}
