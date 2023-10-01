package main

import (
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/handlers"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"github.com/rs/cors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	// or "gorm.io/driver/postgres" for PostgreSQL
)

var db *gorm.DB

func main() {
	// Read environment variables
	dbType := os.Getenv("DB_TYPE")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	var dsn string
	var err error
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
		log.Fatalf("Unsupported DB_TYPE: %s", dbType)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	mux := http.NewServeMux()

	// Migrate the schema
	db.AutoMigrate(&models.User{}, &models.Site{}, &models.UserSite{})
	//generateTestData()

	handler := cors.Default().Handler(mux)

	// Serve static files
	fs := http.FileServer(http.Dir("../frontend/public/")) // Adjust the path based on your directory structure
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if it's an API route first
		if isAPIRoute(r.URL.Path) {
			// Handle API routes separately
			return
		}

		path := "../frontend/public" + r.URL.Path
		log.Println(path)
		_, err := os.Stat(path)

		// If the file exists, serve it
		if !os.IsNotExist(err) {
			fs.ServeHTTP(w, r)
			return
		}

		// Otherwise, serve index.html
		http.ServeFile(w, r, "../frontend/public/index.html")
	})

	mux.HandleFunc("/api/requests", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleRequests(w, r, db)
	})
	mux.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleRegister(w, r, db)
	})
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleLogin(w, r, db)
	})
	mux.HandleFunc("/api/authenticate", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleAuthenticate(w, r, db)
	})
	mux.HandleFunc("/api/request", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleRequestSite(w, r, db)
	})
	// Start the server on port 8081
	log.Println("Starting server on :8080")

	log.Fatal(http.ListenAndServe(":8080", handler))

	// ... setup your routes and start your server
}
func isAPIRoute(path string) bool {
	return len(path) >= 4 && path[0:4] == "/api"
}
func generateTestData() {
	// Insert test data for Users
	users := []models.User{
		{Email: "user1@example.com", Password: "password1", Role: "admin"},
		{Email: "user2@example.com", Password: "password2", Role: "user"},
		{Email: "user3@example.com", Password: "password3", Role: "user"},
	}
	for _, user := range users {
		db.Create(&user)
	}

	// Insert test data for Sites
	sites := []models.Site{
		{URL: "https://site1.com"},
		{URL: "https://site2.com"},
		{URL: "https://site3.com"},
	}
	for _, site := range sites {
		db.Create(&site)
	}

	// Insert test data for UserSite
	userSites := []models.UserSite{
		{UserID: 1, SiteID: 1, State: "authorized"},
		{UserID: 2, SiteID: 2, State: "requested"},
		{UserID: 3, SiteID: 3, State: "authorized"},
	}
	for _, userSite := range userSites {
		db.Create(&userSite)
	}
}
