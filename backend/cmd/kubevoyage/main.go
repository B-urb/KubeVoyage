package main

import (
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/app"
	"github.com/B-Urb/KubeVoyage/internal/handlers"
	"github.com/B-Urb/KubeVoyage/internal/util"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	// or "gorm.io/driver/postgres" for PostgreSQL
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	length     int
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.statusCode == 0 {
		// Default status code is 200 OK.
		lrw.statusCode = http.StatusOK
	}
	size, err := lrw.ResponseWriter.Write(b)
	lrw.length += size
	return size, err
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w}
		start := time.Now()

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf(
			"Method: %s, Path: %s, RemoteAddr: %s, Duration: %s, StatusCode: %d, ResponseSize: %d bytes\n",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			duration,
			lrw.statusCode,
			lrw.length,
		)
	})
}
func main() {
	app, err := application.NewApp() // Assuming NewApp is in the same package
	if err != nil {
		log.Fatalf(err.Error())
	}

	handler := handlers.NewHandler(app.DB)
	err = app.Init()
	if err != nil {
		log.Fatalf(err.Error())
	}

	mux := setupServer(handler)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
func setupServer(handle *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	handler := cors.Default().Handler(mux)
	frontendPathLocal, _ := util.GetEnvOrDefault("FRONTEND_PATH", "./public")
	log.Printf("Serving frontend from %s", frontendPathLocal)

	// Serve static files
	fs := http.FileServer(http.Dir(frontendPathLocal)) // Adjust the path based on your directory structure
	mux.Handle("/", logMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if it's an API route first
		if isAPIRoute(r.URL.Path) {
			// Handle API routes separately
			return
		}

		path := frontendPathLocal + r.URL.Path
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			fmt.Println("Error getting absolute path:", err)
			return
		}
		fmt.Println("Absolute Path:", absolutePath)
		_, err = os.Stat(path)

		// If the file exists, serve it
		if !os.IsNotExist(err) {
			fs.ServeHTTP(w, r)
			return
		} else {
			log.Println(err)
		}
		// Otherwise, serve index.html
		http.ServeFile(w, r, frontendPathLocal+"/index.html")
	})))

	mux.Handle("/api/requests", logMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle.HandleRequests(w, r)
	})))
	mux.Handle("/api/requests/update", logMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle.HandleUpdateSiteState(w, r)
	})))
	mux.Handle("/api/register", logMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle.HandleRegister(w, r)
	})))
	mux.Handle("/api/login", logMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle.HandleLogin(w, r)
	})))
	mux.Handle("/api/authenticate", logMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle.HandleAuthenticate(w, r)
	})))
	mux.Handle("/api/redirect", logMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle.HandleRedirect(w, r)
	})))
	mux.Handle("/api/request", logMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle.HandleRequestSite(w, r)
	})))

	return handler
}

func isAPIRoute(path string) bool {
	return len(path) >= 4 && path[0:4] == "/api"
}
