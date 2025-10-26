package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"node-herder/utils"

	"github.com/golang-jwt/jwt/v5"
)

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
			Issuer:    "nodeherder",
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
