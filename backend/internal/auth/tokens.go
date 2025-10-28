package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"node-herder/utils"

	"github.com/golang-jwt/jwt/v5"
)

// =========================
// JWT service and claims
// =========================

const Issuer = "nodeherder"

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTConfig struct {
	Secret     []byte
	Expiration time.Duration
}

type JWTService struct {
	cfg *JWTConfig
}

func WithDefaultJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:     []byte("JWT_SECRET_KEY"),
		Expiration: 24 * time.Hour,
	}
}

func NewJWTService(cfg JWTConfig) *JWTService {
	return &JWTService{
		cfg: &cfg,
	}
}

func (s *JWTService) GenerateJWT(userID, username string) (string, error) {

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.cfg.Secret)
}

func (s *JWTService) ValidateJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.cfg.Secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// =========================
// Cookie helpers
// =========================

// setStateCookie stores the OAuth 2.0 state parameter in an HttpOnly cookie for CSRF protection.
func setStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameOAuthState,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes expiry
	})
}

// setPKCECookie stores the PKCE code_verifier bound to a state value (format: state.verifier).
func setPKCECookie(w http.ResponseWriter, state, verifier string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameOAuthPKCE,
		Value:    state + "." + verifier,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes expiry
	})
}

// setSessionCookie persists the user's session (JWT) with HttpOnly and SameSite=Lax.
// Note: set Secure=true in production when served over HTTPS.
func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameSession,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

// clearCookie expires a cookie immediately by name.
func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

// validateStateCookie ensures the received state matches the stored cookie value.
func validateStateCookie(r *http.Request, expectedState string) error {
	c, err := r.Cookie(CookieNameOAuthState)
	if err != nil {
		return errors.New("state cookie missing")
	}
	if c.Value != expectedState {
		return errors.New("state mismatch")
	}
	return nil
}

// extractPKCEVerifierFromCookie parses and validates the PKCE cookie binding and returns the verifier.
func extractPKCEVerifierFromCookie(r *http.Request, expectedState string) (string, error) {
	c, err := r.Cookie(CookieNameOAuthPKCE)
	if err != nil || c.Value == "" {
		return "", errors.New("missing PKCE verifier")
	}
	// Expect format: state.verifier
	if dot := strings.IndexByte(c.Value, '.'); dot > 0 {
		cookieState := c.Value[:dot]
		verifier := c.Value[dot+1:]
		if cookieState != expectedState || verifier == "" {
			return "", errors.New("invalid PKCE binding")
		}
		return verifier, nil
	}
	return "", errors.New("invalid PKCE format")
}

// frontendAuthSuccessURL returns the absolute URL to redirect to after successful auth.
func frontendAuthSuccessURL() (string, error) {
	base, err := utils.GetFrontendBaseURL()
	if err != nil {
		return "", err
	}
	return base + "/?auth=success", nil
}

// =========================
// PKCE + OAuth helpers
// =========================

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

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

// =========================
// Minimal HTML snippets for auth UX
// =========================

// OfflineAuthSuccessHTML returns a tiny HTML page used in the offline flow when a popup
// is opened to create a local session. It will try to notify the opener via postMessage
// and then close itself. If no opener exists, it shows a small message so the user can
// close the window manually.
func OfflineAuthSuccessHTML() string {
	return "<!doctype html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>Authenticated</title><style>html,body{height:100%;margin:0;font-family:system-ui,-apple-system,Segoe UI,Roboto,Ubuntu,\"Helvetica Neue\",Arial,sans-serif;background:#0b1020;color:#e8ecf1}main{min-height:100%;display:flex;align-items:center;justify-content:center}section{background:#111733;border:1px solid #1b254b;border-radius:10px;padding:20px 24px;box-shadow:0 10px 30px rgba(0,0,0,.4);max-width:420px}h1{font-size:18px;margin:0 0 8px}p{margin:0 0 10px;opacity:.9}small{opacity:.7}</style></head><body><main><section><h1>You're signed in (offline)</h1><p>You can close this window.</p><small>We'll close it automatically.</small></section></main><script>try{if(window.opener){window.opener.postMessage({type:'auth-success'},'*')}}catch(e){}try{window.close()}catch(e){}</script></body></html>"
}
