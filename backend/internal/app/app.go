package application

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/database"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"github.com/B-Urb/KubeVoyage/internal/util"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
)

type App struct {
	DB *gorm.DB
}

func NewApp() (*App, error) {
	db, err := database.InitializeDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return &App{
		DB: db,
	}, nil
}

func (app *App) Init() error {
	err := app.DB.AutoMigrate(models.User{}, models.Site{}, models.UserSite{})
	if err != nil {
		return err
	}
	err = createAdminUserIfNotExist(app.DB)
	if err != nil {
		return err
	}
	return nil
}

func createAdminUserIfNotExist(db *gorm.DB) error {
	adminEmail, _ := util.GetEnvOrDefault("ADMIN_USER", "admin@admin.de")
	adminPassword, _ := util.GetEnvOrDefault("ADMIN_PASSWORD", "test")

	var existingUser models.User
	if err := db.Where("email = ?", adminEmail).First(&existingUser).Error; err == nil {
		// Admin user already exists
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Hash the password using scrypt
	hash, err := hashPassword(adminPassword)
	if err != nil {
		return err
	}

	adminUser := models.User{
		Email:    adminEmail,
		Password: hash,
		Role:     "admin",
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return err
	}

	return nil
}

func hashPassword(password string) (string, error) {
	// Hash the password using scrypt
	hash, err := scrypt.Key([]byte(password), nil, 16384, 8, 1, 32)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash), nil
}
