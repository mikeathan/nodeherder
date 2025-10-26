package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
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
	oauth     OAuth
	blacklist *TokenBlacklist
}

func NewProvider(cfg JWTConfig, oauthCallbackURL string) *Provider {
	j := NewJWTService(cfg)
	b := NewTokenBlacklist()
	o := NewGoogleOAuth(j, b, oauthCallbackURL)
	return &Provider{jwt: j, oauth: o, blacklist: b}
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

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// PKCE helpers
func generatePKCEVerifier(length int) string {
	if length < 43 {
		length = 43
	}
	if length > 128 {
		length = 128
	}
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	b := make([]byte, length)
	rb := make([]byte, length)
	if _, err := rand.Read(rb); err != nil {
		// fallback
		return base64.RawURLEncoding.EncodeToString(rb)
	}
	for i := 0; i < length; i++ {
		b[i] = charset[int(rb[i])%len(charset)]
	}
	return string(b)
}

func generatePKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
