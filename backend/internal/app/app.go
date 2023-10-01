package main

import (
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"gorm.io/gorm"
	"log"
	"os"
)

type App struct {
	DB      *gorm.DB
	JWTKey  []byte
	BaseURL string
}

func NewApp() (*App, error) {
	db, err := initializeDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	jwtKey, err := getEnvOrError("JWT_SECRET_KEY")
	if err != nil {
		log.Fatalf("Error reading JWT_SECRET_KEY: %v", err)
	}

	baseURL, err := getEnvOrError("BASE_URL")
	if err != nil {
		log.Fatalf("Error reading BASE_URL: %v", err)
	}

	return &App{
		DB:      db,
		JWTKey:  []byte(jwtKey),
		BaseURL: baseURL,
	}, nil
}

func (app *App) Migrate() {
	app.DB.AutoMigrate(&models.User{}, &models.Site{}, &models.UserSite{})
}

func getEnvOrError(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not set", key)
	}
	return value, nil
}
