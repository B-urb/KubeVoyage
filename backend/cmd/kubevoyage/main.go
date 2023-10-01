package main

import (
	"github.com/B-Urb/KubeVoyage/internal/handlers"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"github.com/rs/cors"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	// or "gorm.io/driver/postgres" for PostgreSQL
)

var db *gorm.DB

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	app.Migrate()

	handler := setupServer(app)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
func setupServer(app *App) http.Handler {
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

	handler := cors.Default().Handler(mux)
	return handler
}

func isAPIRoute(path string) bool {
	return len(path) >= 4 && path[0:4] == "/api"
}
