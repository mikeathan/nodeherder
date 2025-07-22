package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func GetFrontendBaseURL() (string, error) {
	env := os.Getenv("APP_ENV")

	switch env {
	case "production":
		url := os.Getenv("FRONTEND_BASE_URL")
		if url == "" {
			return "", errors.New("FRONTEND_BASE_URL must be set in production")
		}
		return url, nil

	case "staging":
		url := os.Getenv("FRONTEND_BASE_URL")
		if url == "" {
			return "", errors.New("FRONTEND_BASE_URL must be set in staging")
		}
		return url, nil

	case "development", "":
		url := os.Getenv("FRONTEND_BASE_URL")
		if url == "" {
			return "", errors.New("FRONTEND_BASE_URL must be set in development")
		}
		return url, nil

	default:
		return "", fmt.Errorf("unknown APP_ENV value: %q", env)
	}
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

	return nil
}
