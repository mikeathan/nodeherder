package auth

import (
	"errors"
	"time"

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
