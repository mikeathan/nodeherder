package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var supporteEnvironments = []string{"production", "staging", "development"}

func GetFrontendBaseURL() (string, error) {
	env := os.Getenv("APP_ENV")

	for _, e := range supporteEnvironments {
		if e == env {
			url := os.Getenv("FRONTEND_BASE_URL")
			if url == "" {
				return "", fmt.Errorf("FRONTEND_BASE_URL is not set for APP_ENV=%q", env)
			}

			return url, nil
		}
	}

	return "", fmt.Errorf("unknown APP_ENV value: %q", env)
}

func GetJwtSecretKey() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET_KEY is not set")
	}
	return []byte(secret), nil
}

// ValidateServiceCredentials checks if the provided client ID and secret
// match any of the configured service clients in the environment variables.
func ValidateServiceCredentials(clientID, clientSecret string) bool {
	raw := os.Getenv("SERVICE_CLIENTS")
	if raw == "" {
		return false
	}

	for _, id := range strings.Split(raw, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}

		secret := os.Getenv("SERVICE_SECRET_" + id)
		if secret == "" {
			continue
		}

		if clientID == id && clientSecret == secret {
			return true
		}
	}

	return false
}

func GetAuthLocalOfflineMode() bool {
	mode := os.Getenv("OFFLINE_STRICT_LOCAL")
	return mode == "true"
}

func GetAuthCallbackURL(port int) (string, error) {
	env := os.Getenv("APP_ENV")

	for _, e := range supporteEnvironments {
		if e == env {
			url := os.Getenv("OAUTH_CALLBACK_URL")
			if url == "" {
				return "", fmt.Errorf("OAUTH_CALLBACK_URL is not set for APP_ENV=%q", env)
			}

			callbackURL := strings.ReplaceAll(url, "{PORT}", strconv.Itoa(port))
			return callbackURL, nil
		}
	}

	return "", fmt.Errorf("unknown APP_ENV value: %q", env)
}

func GetFrontendAllowedOrigins() []string {
	originsEnv := os.Getenv("FRONTEND_ALLOWED_ORIGINS")
	if originsEnv == "" {
		return []string{}
	}
	origins := strings.Split(originsEnv, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return origins
}

func GetMQTTBrokerURL() (string, error) {
	mqttURL := os.Getenv("MQTT_URL")
	if mqttURL == "" {
		env := os.Getenv("APP_ENV")
		return "", fmt.Errorf("MQTT_URL is not set for APP_ENV=%q", env)
	}
	return mqttURL, nil
}

func GetMQTTBrokerCredentials() (string, string) {
	username := os.Getenv("MQTT_USER")
	password := os.Getenv("MQTT_PASS")
	return username, password
}

func GetMQTTClientID() string {
	clientID := os.Getenv("MQTT_CLIENT_ID")
	return clientID
}

func LoadEnviromentConfig() error {
	// Local-only
	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load()
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Load env file explicitly
	envFile := ".env." + env
	if _, err := os.Stat(envFile); err == nil {
		_ = godotenv.Load(envFile)
		LogInfo("loaded env file: " + envFile)
	}
	return nil
}
