package util

import (
	"fmt"
	"os"
)

func GetEnvOrError(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not set", key)
	}
	return value, nil
}

func GetEnvOrDefault(key string, defaultValue string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}
