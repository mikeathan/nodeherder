package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuth interface {
	HandleLogin(http.ResponseWriter, *http.Request)
	HandleCallback(http.ResponseWriter, *http.Request)
	RegisterRoutes(r *Router)
}

type googleOAuth struct {
	config *oauth2.Config
}

func NewGoogleOAuth() OAuth {
	return &googleOAuth{
		config: buildConfig(),
	}
}

func buildConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  "http://localhost:4110/api/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (g *googleOAuth) RegisterRoutes(r *Router) {
	base := "/api/auth/google"
	r.GET(base+"/login", http.HandlerFunc(g.HandleLogin))
	r.GET(base+"/callback", http.HandlerFunc(g.HandleCallback))
	r.GET("/api/auth/me", http.HandlerFunc(handleMe))
}

func (o *googleOAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {

	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:  "oauthstate",
		Value: state,
		Path:  "/",
	})

	url := o.config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (o *googleOAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	token, err := o.config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	// Later you can validate session/JWT etc.
	w.Write([]byte("You are logged in"))

	// 	 user, err := getUserFromSessionOrJWT(r)
	// if err != nil {
	//     http.Error(w, "unauthorized", http.StatusUnauthorized)
	//     return
	// }
	// json.NewEncoder(w).Encode(user)
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
