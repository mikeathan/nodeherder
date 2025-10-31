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

func GetMQTTBrokerURL() (string, error) {
	mqttURL := os.Getenv("MQTT_URL")
	if mqttURL == "" {
		env := os.Getenv("APP_ENV")
		return "", fmt.Errorf("MQTT_URL is not set for APP_ENV=%q", env)
	}
	return mqttURL, nil
}

func LoadEnviromentConfig() error {

	// loads .env if present
	_ = godotenv.Load()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	envFile := ".env." + env

	// Load env file explicitly
	err := godotenv.Load(envFile)
	if err != nil {
		return err
	}
	LogInfo("loaded env file: " + envFile)
	return nil
}
