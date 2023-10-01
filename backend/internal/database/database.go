package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

func InitializeDatabase() (*gorm.DB, error) {
	// Read environment variables
	dbType, err := getEnvOrError("DB_TYPE")
	if err != nil {
		return nil, err
	}

	dbHost, err := getEnvOrError("DB_HOST")
	if err != nil {
		return nil, err
	}

	dbPort, err := getEnvOrError("DB_PORT")
	if err != nil {
		return nil, err
	}

	dbUser, err := getEnvOrError("DB_USER")
	if err != nil {
		return nil, err
	}

	dbPassword, err := getEnvOrError("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbName, err := getEnvOrError("DB_NAME")
	if err != nil {
		return nil, err
	}

	var dsn string
	var db *gorm.DB

	switch dbType {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

func getEnvOrError(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("Environment variable %s not set", key)
	}
	return value, nil
}
