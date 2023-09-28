package main

import (
	"github.com/B-Urb/KubeVoyage/internal/handlers"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"github.com/rs/cors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	// or "gorm.io/driver/postgres" for PostgreSQL
)

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	mux := http.NewServeMux()

	// Migrate the schema
	db.AutoMigrate(&models.User{}, &models.Site{}, &models.UserSite{})
	generateTestData()
	handler := cors.Default().Handler(mux)
	mux.HandleFunc("/api/requests", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleRequests(w, r, db)
	})
	mux.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleRegister(w, r, db)
	})
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleRegister(w, r, db)
	})
	// Start the server on port 8081
	log.Fatal(http.ListenAndServe(":8081", handler))

	// ... setup your routes and start your server
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
