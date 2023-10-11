package handlers

import (
	"encoding/json"
	"errors"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

func HandleRequests(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var results []models.UserSiteResponse

	log.Println("Incoming Request")
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db.Table("user_sites").
		Select("users.email as user, sites.url as site, user_sites.state as state").
		Joins("JOIN users ON users.id = user_sites.user_id").
		Joins("JOIN sites ON sites.id = user_sites.site_id").
		Scan(&results)
	w.Header().Set("Content-Type", "application/json")

	// Convert the results to JSON and send the response
	json.NewEncoder(w).Encode(results)

}

func (h *Handler) HandleRequestSite(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	userEmail, err := h.getUserEmailFromToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// Query the User table to get the unique ID associated with the email
	var user models.User
	if err := db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	siteURL := r.URL.Query().Get("redirect")

	// Check if site already exists
	var site models.Site
	if err := db.Where("url = ?", siteURL).First(&site).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		// If not, create a new site entry
		site = models.Site{URL: siteURL}
		db.Create(&site)
	}

	// Create a new UserSite entry with state "requested"
	userSite := models.UserSite{
		UserID: user.ID, // Use the ID from the user query
		SiteID: site.ID,
		State:  "requested",
	}
	db.Create(&userSite)

	w.Write([]byte("Request submitted"))
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true, // Set this to true if using HTTPS
		Domain:   "",   // Adjust to your domain
		Path:     "/",
	})

	w.Write([]byte("Logged out successfully"))
}
