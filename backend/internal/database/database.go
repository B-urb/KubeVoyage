package database

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"github.com/B-Urb/KubeVoyage/internal/util"
	"golang.org/x/crypto/scrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitializeDatabase() (*gorm.DB, error) {
	// Read environment variables
	dbType, err := util.GetEnvOrDefault("DB_TYPE", "sqlite")
	if err != nil {
		return nil, err
	}

	var dsn string
	var db *gorm.DB

	switch dbType {
	case "mysql":
		dbHost, err := util.GetEnvOrError("DB_HOST")
		if err != nil {
			return nil, err
		}

		dbPort, err := util.GetEnvOrError("DB_PORT")
		if err != nil {
			return nil, err
		}

		dbUser, err := util.GetEnvOrError("DB_USER")
		if err != nil {
			return nil, err
		}

		dbPassword, err := util.GetEnvOrError("DB_PASSWORD")
		if err != nil {
			return nil, err
		}

		dbName, err := util.GetEnvOrDefault("DB_NAME", "kubevoyage")
		if err != nil {
			return nil, err
		}

		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	case "postgres":
		dbHost, err := util.GetEnvOrError("DB_HOST")
		if err != nil {
			return nil, err
		}

		dbPort, err := util.GetEnvOrError("DB_PORT")
		if err != nil {
			return nil, err
		}

		dbUser, err := util.GetEnvOrError("DB_USER")
		if err != nil {
			return nil, err
		}

		dbPassword, err := util.GetEnvOrError("DB_PASSWORD")
		if err != nil {
			return nil, err
		}

		dbName, err := util.GetEnvOrDefault("DB_NAME", "kubevoyage")
		if err != nil {
			return nil, err
		}

		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	case "sqlite":
		dbName, err := util.GetEnvOrDefault("DB_NAME", "kubevoyage")
		if err != nil {
			return nil, err
		}

		dsn = dbName // For SQLite, dbName would be the path to the .db file
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})

	default:
		return nil, fmt.Errorf("Unsupported DB_TYPE: %s", dbType)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %v", err)
	}

	// After successfully connecting to the database
	if err := createAdminUserIfNotExist(db); err != nil {
		return nil, fmt.Errorf("Failed to create admin user: %v", err)
	}

	return db, nil
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
