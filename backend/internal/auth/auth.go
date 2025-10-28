package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"node-herder/utils"
	"os"

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
	oauth     AuthProvider
	blacklist *TokenBlacklist
}

func NewProvider(cfg JWTConfig, oauthCallbackURL string) *Provider {
	j := NewJWTService(cfg)
	b := NewTokenBlacklist()
	online := NewGoogleOAuth(j, b, oauthCallbackURL)
	offline := NewOfflineOAuth(j, b, oauthCallbackURL)

	decide := func(r *http.Request) bool {
		if !utils.GetAuthLocalOfflineMode() {
			return true
		}

		return !utils.IsLocalRequest(r)
	}

	o := newAuthDelegator(online, offline, decide)

	utils.LogInfof("Auth hybrid enabled (default offline=%v)", utils.GetAuthLocalOfflineMode())
	return &Provider{jwt: j, oauth: o, blacklist: b}
}

func (m *Provider) Middleware() func(http.Handler) http.Handler {
	return Auth(m.jwt, m.blacklist)
}

func (m *Provider) OAuth() AuthProvider {
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
	HandleLogout(http.ResponseWriter, *http.Request)
	HandleMe(http.ResponseWriter, *http.Request)
}

// AuthProvider is the top-level interface for route registration
type AuthProvider interface {
	OAuth
	RegisterRoutes(registrar RouteRegistrar)
}

// Optional interface for offline-specific routes
type OfflineAuth interface {
	OAuth
	HandleOfflineStart(http.ResponseWriter, *http.Request)
}

// Offline Local OAuth implementation
type offlineOAuth struct {
	jwtService  *JWTService
	blacklist   *TokenBlacklist
	basePath    string
	RedirectURL string
}

func NewOfflineOAuth(jwtService *JWTService, blacklist *TokenBlacklist, callbackURL string) OAuth {
	return &offlineOAuth{
		jwtService:  jwtService,
		blacklist:   blacklist,
		basePath:    "/api/auth",
		RedirectURL: callbackURL,
	}
}

func (o *offlineOAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {

	// Return a URL to open in the popup that will create a local session and close the window.
	url := o.RedirectURL + o.basePath + "/offline/start"

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func (o *offlineOAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// Not implemented for offline mode
	writeJSONAuthError(w, http.StatusNotFound, "callback not supported in offline mode")
}

func (o *offlineOAuth) HandleOfflineStart(w http.ResponseWriter, r *http.Request) {

	// Create a local session for an offline user
	token, err := o.jwtService.GenerateJWT("local", "Offline User")
	if err != nil {
		writeJSONAuthError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	setSessionCookie(w, token)

	// Respond with a tiny page that notifies the opener and closes the popup
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, OfflineAuthSuccessHTML())
}

func (o *offlineOAuth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	LogoutHandler(o.jwtService, o.blacklist)(w, r)
}

func (o *offlineOAuth) HandleMe(w http.ResponseWriter, r *http.Request) {
	MeHandler()(w, r)
}

// Google OAuth implementation
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
	LogoutHandler(o.jwtService, o.blacklist)(w, r)
}

func (o *googleOAuth) HandleMe(w http.ResponseWriter, r *http.Request) {
	MeHandler()(w, r)
}

// authDelegator
// Delegator that routes to online or offline auth
// based on a simple predicate. If decide(r) returns true, offline is used;
// otherwise online is used.
type authDelegator struct {
	online  OAuth
	offline OAuth
	decide  func(*http.Request) bool
}

func newAuthDelegator(online, offline OAuth, decide func(*http.Request) bool) *authDelegator {
	return &authDelegator{online: online, offline: offline, decide: decide}
}

func (s *authDelegator) pick(r *http.Request) OAuth {
	if s.decide != nil && s.decide(r) {
		return s.offline
	}
	return s.online
}

func (s *authDelegator) HandleLogin(w http.ResponseWriter, r *http.Request) {
	s.pick(r).HandleLogin(w, r)
}

func (s *authDelegator) HandleCallback(w http.ResponseWriter, r *http.Request) {
	s.pick(r).HandleCallback(w, r)
}

func (s *authDelegator) HandleLogout(w http.ResponseWriter, r *http.Request) {
	s.pick(r).HandleLogout(w, r)
}

func (s *authDelegator) HandleMe(w http.ResponseWriter, r *http.Request) {
	s.pick(r).HandleMe(w, r)
}

func (s *authDelegator) HandleOfflineStart(w http.ResponseWriter, r *http.Request) {
	if h, ok := s.offline.(OfflineAuth); ok {
		h.HandleOfflineStart(w, r)
		return
	}
	writeJSONAuthError(w, http.StatusNotFound, "offline start not available")
}

func (s *authDelegator) RegisterRoutes(registrar RouteRegistrar) {
	base := "/api/auth"
	registrar.PublicPOST(base+"/login", http.HandlerFunc(s.HandleLogin))
	registrar.PublicPOST(base+"/logout", http.HandlerFunc(s.HandleLogout))
	registrar.PublicGET(base+"/callback", http.HandlerFunc(s.HandleCallback))
	registrar.PublicGET(base+"/offline/start", http.HandlerFunc(s.HandleOfflineStart))
	registrar.GET(base+"/me", http.HandlerFunc(s.HandleMe))
}
