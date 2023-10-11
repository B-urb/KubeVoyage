package database

import (
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/util"
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

	dbName, err := util.GetEnvOrDefault("DB_NAME", "kubevoyage")
	if err != nil {
		return nil, err
	}

	switch dbType {
	case "mysql", "postgres":
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

		if dbType == "mysql" {
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		} else {
			dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		}

	case "sqlite":
		dsn = dbName // For SQLite, dbName would be the path to the .db file
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})

	default:
		return nil, fmt.Errorf("Unsupported DB_TYPE: %s", dbType)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %v", err)
	}

	return db, nil
}
