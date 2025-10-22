package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const AuthStateCookie = "oauthstate"
const AuthCookie = "session"

type Provider struct {
	jwt       *JWTService
	oauth     OAuth
	blacklist *TokenBlacklist
}

func NewProvider(cfg JWTConfig) *Provider {
	j := NewJWTService(cfg)
	b := NewTokenBlacklist()
	o := NewGoogleOAuth(j, b)
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

func NewGoogleOAuth(jwtService *JWTService, blacklist *TokenBlacklist) OAuth {
	config := &oauth2.Config{
		RedirectURL:  "http://localhost:4110/api/auth/callback",
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
	http.SetCookie(w, &http.Cookie{
		Name:  AuthStateCookie,
		Value: state,
		Path:  "/",
	})

	url := o.config.AuthCodeURL(state)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})

}

func (o *googleOAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	fmt.Println("Callback hit, code:", code, "state:", state)

	if code == "" || state == "" {
		writeJSONAuthError(w, http.StatusBadRequest, "Missing code or state")
		return
	}

	cookie, err := r.Cookie(AuthStateCookie)
	if err != nil || cookie.Value != state {
		writeJSONAuthError(w, http.StatusBadRequest, "Invalid state")
		return
	}
	token, err := o.config.Exchange(context.Background(), code)
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

	// Set auth cookie with SameSite=Lax (same domain, different ports = same-site)
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookie,
		Value:    jwt,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})


	// Return small HTML page that posts message to popup opener
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
        <script>
            try {
                if (window.opener) {
                    window.opener.postMessage({ 
                        status: 'success', 
                        user: { id: '%s', username: '%s' }
                    }, "*");
                }
            } catch (e) {}
            window.close();
        </script>
    `, userID, userName)
}

func (o *googleOAuth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie or header
	var token string
	if c, err := r.Cookie(AuthCookie); err == nil {
		token = c.Value
	}

	if token == "" {
		writeJSONAuthError(w, http.StatusInternalServerError, "no auth token found")
		return
	}

	// if token == "" {
	// 	// Fallback to Authorization header
	// 	auth := r.Header.Get("Authorization")
	// 	if auth != "" {
	// 		parts := strings.SplitN(auth, " ", 2)
	// 		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
	// 			token = parts[1]
	// 		}
	// 	}
	// }

	// Blacklist the token if found and valid
	if token != "" {
		if claims, err := o.jwtService.ValidateJWT(token); err == nil {
			o.blacklist.Add(token, claims.ExpiresAt.Time)
		}
	}

	// Clear the auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

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
