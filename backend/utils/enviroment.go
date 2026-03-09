package utils

import (
	"fmt"
	"os"

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

func LoadEnviromentConfig() error {
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
