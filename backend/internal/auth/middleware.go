package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const userKey contextKey = "user"
const sessionKey contextKey = "session"

type requestSession struct {
	token     string
	expires   time.Time
	blacklist *TokenBlacklist
}

// SessionValid rechecks the lifetime of an authenticated transport request.
// Unauthenticated transports retain their routing policy; Auth installs the lease.
func SessionValid(r *http.Request) bool {
	session, ok := r.Context().Value(sessionKey).(requestSession)
	if !ok {
		return true
	}
	return (session.expires.IsZero() || time.Now().Before(session.expires)) &&
		(session.blacklist == nil || !session.blacklist.IsBlacklisted(session.token))
}

// WatchSession closes long-lived transports on expiry/revocation. The transport
// must stop the watcher when it disconnects, including failed upgrades.
func WatchSession(r *http.Request, closeTransport func()) func() {
	session, ok := r.Context().Value(sessionKey).(requestSession)
	if !ok {
		return func() {}
	}
	var timer *time.Timer
	if !session.expires.IsZero() {
		timer = time.AfterFunc(time.Until(session.expires), closeTransport)
	}
	stop := func() {}
	if session.blacklist != nil {
		stop = session.blacklist.OnRevoke(session.token, closeTransport)
	}
	if !SessionValid(r) {
		closeTransport()
	}
	return func() {
		if timer != nil {
			timer.Stop()
		}
		stop()
	}
}

func Auth(jwtService *JWTService, blacklist *TokenBlacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			if c, err := r.Cookie(CookieNameSession); err == nil {
				token = c.Value
			}

			if token == "" {
				token = extractBearer(r.Header.Get("Authorization"))
			}

			if token == "" {
				writeJSONAuthRequestError(w, r, http.StatusUnauthorized, "missing token")
				return
			}

			// Check if token is blacklisted
			if blacklist != nil && blacklist.IsBlacklisted(token) {
				writeJSONAuthRequestError(w, r, http.StatusUnauthorized, "token revoked")
				return
			}

			user, err := jwtService.ValidateJWT(token)
			if err != nil {
				writeJSONAuthRequestError(w, r, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userKey, user)
			session := requestSession{token: token, blacklist: blacklist}
			if user.ExpiresAt != nil {
				session.expires = user.ExpiresAt.Time
			}
			ctx = context.WithValue(ctx, sessionKey, session)
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

// Keep the existing authentication decision and give Connections its documented envelope.
func writeJSONAuthRequestError(w http.ResponseWriter, r *http.Request, code int, message string) {
	if strings.HasPrefix(r.URL.Path, "/api/integrations/") {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, no-store")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": nil, "error": map[string]string{"code": "unauthorized", "message": "Sign in to use Connections."}})
		return
	}
	writeJSONAuthError(w, code, message)
}
