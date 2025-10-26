package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Cookie names used in the OAuth + session flow.
// - CookieNameOAuthState: short-lived anti-CSRF state used during OAuth redirect
// - CookieNameOAuthPKCE: short-lived PKCE code_verifier (bound to state)
// - CookieNameSession: persisted session/JWT after successful auth
const CookieNameOAuthState = "oauthstate"
const CookieNameSession = "session"
const CookieNameOAuthPKCE = "oauthpkce"

type Provider struct {
	jwt       *JWTService
	oauth     OAuth
	blacklist *TokenBlacklist
}

func NewProvider(cfg JWTConfig, oauthCallbackURL string) *Provider {
	j := NewJWTService(cfg)
	b := NewTokenBlacklist()

	// OFFLINE_MODE modes:
	// - "offline"/"true": force offline auth
	// - "auto"/"dynamic": choose offline for local requests, Google otherwise
	// - default/anything else: Google OAuth
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("OFFLINE_MODE")))
	switch mode {
	case "true", "1", "on", "offline":
		o := NewOfflineOAuth(j, b, oauthCallbackURL)
		return &Provider{jwt: j, oauth: o, blacklist: b}
	case "auto", "dynamic":
		o := NewHybridOAuth(j, b, oauthCallbackURL)
		return &Provider{jwt: j, oauth: o, blacklist: b}
	default:
		o := NewGoogleOAuth(j, b, oauthCallbackURL)
		return &Provider{jwt: j, oauth: o, blacklist: b}
	}
}

func (m *Provider) Middleware() func(http.Handler) http.Handler {
	return Auth(m.jwt, m.blacklist)
}

func (m *Provider) OAuth() OAuth {
	return m.oauth
}

type RouteRegistrar interface {
	GET(path string, handler http.Handler)
	POST(path string, handler http.Handler)
	PublicGET(path string, handler http.Handler)
	PublicPOST(path string, handler http.Handler)
}

type OAuth interface {
	HandleLogin(http.ResponseWriter, *http.Request)
	HandleCallback(http.ResponseWriter, *http.Request)
	RegisterRoutes(registrar RouteRegistrar)
}

type googleOAuth struct {
	config     *oauth2.Config
	jwtService *JWTService
	blacklist  *TokenBlacklist
	basePath   string
}

func NewGoogleOAuth(jwtService *JWTService, blacklist *TokenBlacklist, callbackURL string) OAuth {
	config := &oauth2.Config{
		RedirectURL:  callbackURL,
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return &googleOAuth{
		config:     config,
		jwtService: jwtService,
		blacklist:  blacklist,
		basePath:   "/api/auth",
	}
}

func (g *googleOAuth) RegisterRoutes(registrar RouteRegistrar) {
	registrar.PublicPOST(g.basePath+"/login", http.HandlerFunc(g.HandleLogin))
	registrar.PublicPOST(g.basePath+"/logout", http.HandlerFunc(g.HandleLogout))
	registrar.PublicGET(g.basePath+"/callback", http.HandlerFunc(g.HandleCallback))
	registrar.GET(g.basePath+"/me", http.HandlerFunc(g.handleMe))
}

func (o *googleOAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {

	state := generateState()
	// Store an anti-CSRF state value. Google will return this value and we verify it in the callback.
	// SameSite=Lax allows redirect back across ports on the same host and protects against CSRF on top-level nav.
	setStateCookie(w, state)

	// PKCE: generate verifier and challenge
	verifier := generatePKCEVerifier(64)
	challenge := generatePKCEChallenge(verifier)

	// Store the PKCE code_verifier in an HttpOnly cookie, bound to the state (format: state.verifier).
	setPKCECookie(w, state, verifier)

	url := o.config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})

}

func (o *googleOAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		writeJSONAuthError(w, http.StatusBadRequest, "Missing code or state")
		return
	}

	// Verify the anti-CSRF state value matches what we stored during login.
	if err := validateStateCookie(r, state); err != nil {
		writeJSONAuthError(w, http.StatusBadRequest, "Invalid state")
		return
	}

	// Retrieve and validate the PKCE verifier from the cookie, ensuring it's bound to the same state.
	verifier, err := extractPKCEVerifierFromCookie(r, state)
	if err != nil {
		writeJSONAuthError(w, http.StatusBadRequest, err.Error())
		return
	}
	token, err := o.config.Exchange(
		context.Background(),
		code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		writeJSONAuthError(w, http.StatusBadRequest, "Failed to exchange token: "+err.Error())
		return
	}

	if token.AccessToken == "" {
		writeJSONAuthError(w, http.StatusBadRequest, "Empty access token received from Google")
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		writeJSONAuthError(w, http.StatusBadRequest, "Failed to get user info: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var user map[string]any
	json.NewDecoder(resp.Body).Decode(&user)

	userID, okID := user["id"].(string)
	userName, okName := user["name"].(string)
	if !okID || !okName {
		writeJSONAuthError(w, http.StatusInternalServerError, "Invalid user info from Google")
		return
	}

	// create JWT
	jwt, err := o.jwtService.GenerateJWT(userID, userName)
	if err != nil {
		writeJSONAuthError(w, http.StatusInternalServerError, "Failed to create JWT: "+err.Error())
		return
	}

	// Persist the authenticated session (JWT) as an HttpOnly cookie; SameSite=Lax works with top-level redirects.
	setSessionCookie(w, jwt)

	// Build the frontend success redirect URL from environment.
	frontendURL, err := frontendAuthSuccessURL()
	if err != nil {
		writeJSONAuthError(w, http.StatusInternalServerError, "Failed to get URL for redirect: "+err.Error())
		return
	}

	// Clear the short-lived state and PKCE cookies; they're no longer needed after successful exchange.
	clearCookie(w, CookieNameOAuthPKCE)
	clearCookie(w, CookieNameOAuthState)

	// Always redirect back to frontend for full-page OAuth flow.
	http.Redirect(w, r, frontendURL, http.StatusFound)
}

func (o *googleOAuth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie or header
	var token string
	if c, err := r.Cookie(CookieNameSession); err == nil {
		token = c.Value
	}

	if token == "" {
		writeJSONAuthError(w, http.StatusInternalServerError, "no auth token found")
		return
	}

	// Blacklist the token if found and valid
	if token != "" {
		if claims, err := o.jwtService.ValidateJWT(token); err == nil {
			o.blacklist.Add(token, claims.ExpiresAt.Time)
		}
	}

	clearCookie(w, CookieNameSession)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "success"})
}

func (o *googleOAuth) handleMe(w http.ResponseWriter, r *http.Request) {

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

// =========================
// Offline OAuth implementation
// =========================

type offlineOAuth struct {
	jwtService *JWTService
	blacklist  *TokenBlacklist
	basePath   string
	// strictLocal: if true (default), only allow offline login from private/loopback IPs
	strictLocal bool
}

func NewOfflineOAuth(jwtService *JWTService, blacklist *TokenBlacklist, callbackURL string) OAuth {
	strict := true
	if v := os.Getenv("OFFLINE_STRICT_LOCAL"); v != "" {
		strict = !strings.EqualFold(v, "false")
	}
	return &offlineOAuth{
		jwtService:  jwtService,
		blacklist:   blacklist,
		basePath:    "/api/auth",
		strictLocal: strict,
	}
}

func (o *offlineOAuth) RegisterRoutes(registrar RouteRegistrar) {
	registrar.PublicPOST(o.basePath+"/login", http.HandlerFunc(o.HandleLogin))
	registrar.PublicPOST(o.basePath+"/logout", http.HandlerFunc(o.HandleLogout))
	registrar.PublicGET(o.basePath+"/offline/start", http.HandlerFunc(o.HandleOfflineStart))
	registrar.GET(o.basePath+"/me", http.HandlerFunc(o.handleMe))
}

func (o *offlineOAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if o.strictLocal && !isLocalRequest(r) {
		writeJSONAuthError(w, http.StatusForbidden, "offline login not allowed from non-local network")
		return
	}

	// Return a URL to open in the popup that will create a local session and close the window.
	backendBase, err := backendBaseURLFromCallback()
	if err != nil {
		writeJSONAuthError(w, http.StatusInternalServerError, "failed to resolve backend base URL")
		return
	}
	url := backendBase + o.basePath + "/offline/start"
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func (o *offlineOAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// Not used in offline mode
	writeJSONAuthError(w, http.StatusNotFound, "callback not supported in offline mode")
}

func (o *offlineOAuth) HandleOfflineStart(w http.ResponseWriter, r *http.Request) {
	if o.strictLocal && !isLocalRequest(r) {
		writeJSONAuthError(w, http.StatusForbidden, "offline login not allowed from non-local network")
		return
	}

	// Create a local session for an offline user
	token, err := o.jwtService.GenerateJWT("local", "Offline User")
	if err != nil {
		writeJSONAuthError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	setSessionCookie(w, token)

	// Respond with a tiny page that notifies the opener and closes the popup
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<!doctype html><html><body><script>
	  try { if (window.opener) { window.opener.postMessage({ type: 'auth-success' }, '*'); } } catch (e) {}
	  window.close();
	</script>Authenticated (offline). You can close this window.</body></html>`)
}

func (o *offlineOAuth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var token string
	if c, err := r.Cookie(CookieNameSession); err == nil {
		token = c.Value
	}

	if token == "" {
		writeJSONAuthError(w, http.StatusInternalServerError, "no auth token found")
		return
	}

	if claims, err := o.jwtService.ValidateJWT(token); err == nil {
		o.blacklist.Add(token, claims.ExpiresAt.Time)
	}

	clearCookie(w, CookieNameSession)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success"})
}

func (o *offlineOAuth) handleMe(w http.ResponseWriter, r *http.Request) {
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

// isLocalRequest returns true if the request appears to originate from a private or loopback address.
func isLocalRequest(r *http.Request) bool {
	host := r.RemoteAddr
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		host = xri
	} else if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// take the first IP in the list
		parts := strings.Split(xff, ",")
		host = strings.TrimSpace(parts[0])
	}
	// Strip port if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	// Loopback
	if ip.IsLoopback() {
		return true
	}
	// IPv4 private ranges
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 { // 10.0.0.0/8
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 { // 172.16.0.0/12
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 { // 192.168.0.0/16
			return true
		}
	}
	// IPv6 unique local addresses (fc00::/7) or link-local (fe80::/10)
	if ip.To16() != nil && ip.To4() == nil {
		// fc00::/7 -> first 7 bits are 1111110x, we can check prefix fc00::/7 approximately
		if strings.HasPrefix(ip.String(), "fc") || strings.HasPrefix(ip.String(), "fd") {
			return true
		}
		if strings.HasPrefix(ip.String(), "fe80:") {
			return true
		}
	}
	return false
}

// backendBaseURLFromCallback derives the backend origin (scheme://host) from OAUTH_CALLBACK_URL.
func backendBaseURLFromCallback() (string, error) {
	raw := os.Getenv("OAUTH_CALLBACK_URL")
	if raw == "" {
		return "", fmt.Errorf("OAUTH_CALLBACK_URL not set")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + u.Host, nil
}

// PKCE helpers moved to tokens.go

// =========================
// Hybrid OAuth (dynamic selection per request)
// =========================

type hybridOAuth struct {
	google     *googleOAuth
	offline    *offlineOAuth
	jwtService *JWTService
	blacklist  *TokenBlacklist
	basePath   string
}

func NewHybridOAuth(jwtService *JWTService, blacklist *TokenBlacklist, callbackURL string) OAuth {
	// Build underlying providers
	g := NewGoogleOAuth(jwtService, blacklist, callbackURL)
	o := NewOfflineOAuth(jwtService, blacklist, callbackURL)
	// Type assert to concrete types (same package)
	gg := g.(*googleOAuth)
	oo := o.(*offlineOAuth)
	return &hybridOAuth{
		google:     gg,
		offline:    oo,
		jwtService: jwtService,
		blacklist:  blacklist,
		basePath:   "/api/auth",
	}
}

func (h *hybridOAuth) RegisterRoutes(registrar RouteRegistrar) {
	// Single login endpoint dynamically chooses flow
	registrar.PublicPOST(h.basePath+"/login", http.HandlerFunc(h.HandleLogin))

	// Google callback (only used when not local)
	registrar.PublicGET(h.basePath+"/callback", http.HandlerFunc(h.google.HandleCallback))

	// Offline start (only used when local)
	registrar.PublicGET(h.basePath+"/offline/start", http.HandlerFunc(h.offline.HandleOfflineStart))

	// Common endpoints
	registrar.PublicPOST(h.basePath+"/logout", http.HandlerFunc(h.HandleLogout))
	registrar.GET(h.basePath+"/me", http.HandlerFunc(h.handleMe))
}

func (h *hybridOAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// If the request appears local and offline.strictLocal permits it, use offline
	if h.offline.strictLocal {
		if isLocalRequest(r) {
			h.offline.HandleLogin(w, r)
			return
		}
		// Not local, fall back to Google
		h.google.HandleLogin(w, r)
		return
	}
	// If not strict, still prefer offline for local requests, otherwise Google
	if isLocalRequest(r) {
		h.offline.HandleLogin(w, r)
	} else {
		h.google.HandleLogin(w, r)
	}
}

func (h *hybridOAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// Only Google uses callback flow
	h.google.HandleCallback(w, r)
}

func (h *hybridOAuth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var token string
	if c, err := r.Cookie(CookieNameSession); err == nil {
		token = c.Value
	}
	if token == "" {
		writeJSONAuthError(w, http.StatusInternalServerError, "no auth token found")
		return
	}
	if claims, err := h.jwtService.ValidateJWT(token); err == nil {
		h.blacklist.Add(token, claims.ExpiresAt.Time)
	}
	clearCookie(w, CookieNameSession)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success"})
}

func (h *hybridOAuth) handleMe(w http.ResponseWriter, r *http.Request) {
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
