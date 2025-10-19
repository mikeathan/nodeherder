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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing code or state\n"))
		return
	}

	cookie, err := r.Cookie(AuthStateCookie)
	if err != nil || cookie.Value != state {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid state\n"))
		return
	}
	token, err := o.config.Exchange(context.Background(), code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to exchange token: " + err.Error() + "\n"))
		return
	}

	if token.AccessToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty access token received from Google\n"))
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get user info: " + err.Error() + "\n"))
		return
	}
	defer resp.Body.Close()

	var user map[string]any
	json.NewDecoder(resp.Body).Decode(&user)

	userID, okID := user["id"].(string)
	userName, okName := user["name"].(string)
	if !okID || !okName {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Invalid user info from Google\n"))
		return
	}

	// create JWT
	jwt, err := o.jwtService.GenerateJWT(userID, userName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create JWT: " + err.Error() + "\n"))
		return
	}

	// Set auth cookie (for same-origin requests to backend)
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookie,
		Value:    jwt,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	fmt.Printf("User authenticated: %v %v\n", userID, userName)

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
		// Don't use http.Error - it overwrites CORS headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	claims, err := o.jwtService.ValidateJWT(cookie.Value)
	if err != nil {
		// Don't use http.Error - it overwrites CORS headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
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
