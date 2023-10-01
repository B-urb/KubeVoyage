package application

import (
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/database"
	"github.com/B-Urb/KubeVoyage/internal/models"
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

func (app *App) Migrate() {
	err := app.DB.AutoMigrate(models.User{}, models.Site{}, models.UserSite{})
	if err != nil {
		return
	}
}
