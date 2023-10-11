package main

import (
	"github.com/B-Urb/KubeVoyage/internal/app"
	"github.com/B-Urb/KubeVoyage/internal/handlers"
	"github.com/rs/cors"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	// or "gorm.io/driver/postgres" for PostgreSQL
)

var db *gorm.DB

func main() {
	app, err := application.NewApp() // Assuming NewApp is in the same package
	if err != nil {
		// Handle error
	}

	handler := handlers.NewHandler(app.DB)
	app.Migrate()

	mux := setupServer(handler)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
func setupServer(handle *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	handler := cors.Default().Handler(mux)

	// Serve static files
	fs := http.FileServer(http.Dir("../public/")) // Adjust the path based on your directory structure
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if it's an API route first
		if isAPIRoute(r.URL.Path) {
			// Handle API routes separately
			return
		}

		path := "../public" + r.URL.Path
		log.Println(path)
		_, err := os.Stat(path)

		// If the file exists, serve it
		if !os.IsNotExist(err) {
			fs.ServeHTTP(w, r)
			return
		}

		// Otherwise, serve index.html
		http.ServeFile(w, r, "../public/index.html")
	})

	mux.HandleFunc("/api/requests", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleRequests(w, r, db)
	})
	mux.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		handle.HandleRegister(w, r)
	})
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		handle.HandleLogin(w, r)
	})
	mux.HandleFunc("/api/authenticate", func(w http.ResponseWriter, r *http.Request) {
		handle.HandleAuthenticate(w, r)
	})
	mux.HandleFunc("/api/request", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleRequestSite(w, r, db)
	})

	return handler
}

func isAPIRoute(path string) bool {
	return len(path) >= 4 && path[0:4] == "/api"
}
