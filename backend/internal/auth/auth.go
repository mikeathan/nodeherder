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
	jwt   *JWTService
	oauth OAuth
}

func NewProvider(cfg JWTConfig) *Provider {
	j := NewJWTService(cfg)
	o := NewGoogleOAuth(j)
	return &Provider{jwt: j, oauth: o}
}

func (m *Provider) Middleware() func(http.Handler) http.Handler {
	return Auth(m.jwt)
}

func (m *Provider) OAuth() OAuth {
	return m.oauth
}

type RouteRegistrar interface {
	GET(path string, handler http.Handler)
	POST(path string, handler http.Handler)
}

type OAuth interface {
	HandleLogin(http.ResponseWriter, *http.Request)
	HandleCallback(http.ResponseWriter, *http.Request)
	RegisterRoutes(registrar RouteRegistrar)
}

type googleOAuth struct {
	config     *oauth2.Config
	jwtService *JWTService
	basePath   string
}

func NewGoogleOAuth(jwtService *JWTService) OAuth {
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
		basePath:   "/api/auth",
	}
}

// http://localhost:4110/api/auth/callback
func (g *googleOAuth) RegisterRoutes(registrar RouteRegistrar) {
	registrar.POST(g.basePath+"/login", http.HandlerFunc(g.HandleLogin))
	registrar.POST(g.basePath+"/logout", http.HandlerFunc(g.HandleLogout))
	registrar.GET(g.basePath+"/callback", http.HandlerFunc(g.HandleCallback))
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
		http.Error(w, "Missing code or state", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie(AuthStateCookie)
	if err != nil || cookie.Value != state {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	token, err := o.config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusBadRequest)
		return
	}

	if token.AccessToken == "" {
		http.Error(w, "Empty access token received from Google", http.StatusBadRequest)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	var user map[string]any
	json.NewDecoder(resp.Body).Decode(&user)

	userID, okID := user["id"].(string)
	userName, okName := user["name"].(string)
	if !okID || !okName {
		http.Error(w, "Invalid user info from Google", http.StatusInternalServerError)
		return
	}

	// create JWT
	jwt, err := o.jwtService.GenerateJWT(userID, userName)
	if err != nil {
		http.Error(w, "Failed to create JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookie,
		Value:    jwt,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	fmt.Printf("%v %v ", userID, userName)

	// Return small HTML page that posts message to popup opener
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
        <script>
			console.log("Posting message to opener", window.opener);
			if (window.opener) {
				window.opener.postMessage({ status: 'success' }, "*");
				console.log("Message posted");
			}
			window.close();
        </script>
    `))
}

func (o *googleOAuth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "success"})
}

//https://chatgpt.com/c/68f363dc-ab8c-8328-9459-2ae8eb9e9de4

func (o *googleOAuth) handleMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(AuthCookie)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := o.jwtService.ValidateJWT(cookie.Value)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":       claims.UserID,
		"username": claims.Username,
	})
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
