package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func (h *Handler) HandleRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userEmail, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	isAdmin, err := IsUserAdmin(h.db, userEmail)
	if !isAdmin {
		http.Error(w, "Only Admins can view this", http.StatusUnauthorized)
		return
	}
	var results []models.UserSiteResponse
	err = h.db.Table("user_sites").
		Select("users.email as user, sites.url as site, user_sites.state as state").
		Joins("JOIN users ON users.id = user_sites.user_id").
		Joins("JOIN sites ON sites.id = user_sites.site_id").
		Scan(&results).Error

	if err != nil {
		log.Printf("Database error: %v", err)
		sendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, results, http.StatusOK)
}

func (h *Handler) HandleRequestSite(w http.ResponseWriter, r *http.Request) {
	var redirect models.Redirect

	userEmail, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Query the User table to get the unique ID associated with the email
	var user models.User
	if err := h.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// Parse the request body
	err = json.NewDecoder(r.Body).Decode(&redirect)
	if err != nil {
		sendJSONError(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Check if site already exists
	var site models.Site
	if err := h.db.Where("url = ?", redirect.Redirect).First(&site).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		// If not, create a new site entry
		site = models.Site{URL: redirect.Redirect}
		h.db.Create(&site)
	}

	// Create a new UserSite entry with state "requested"
	userSite := models.UserSite{
		UserID: user.ID, // Use the ID from the user query
		SiteID: site.ID,
		State:  models.Requested,
	}
	h.db.Create(&userSite)

	w.Write([]byte("Request submitted"))
}
func (h *Handler) HandleUpdateSiteState(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		UserEmail string `json:"userEmail"`
		SiteURL   string `json:"siteURL"`
		NewState  string `json:"newState"`
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body RequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	state := models.State(body.NewState)
	if !state.IsValid() {
		http.Error(w, "Invalid state value", http.StatusBadRequest)
		return
	}
	userEmail, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	isAdmin, err := IsUserAdmin(h.db, userEmail)
	if !isAdmin {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var userID uint
	if err := h.db.Model(&models.User{}).Where("email = ?", body.UserEmail).Select("id").First(&userID).Error; err != nil {
		http.Error(w, fmt.Errorf("failed to find user: %w", err).Error(), http.StatusBadRequest)
		return
	}

	// 2. Find the site ID
	var siteID uint
	if err := h.db.Model(&models.Site{}).Where("url = ?", body.SiteURL).Select("id").First(&siteID).Error; err != nil {
		http.Error(w, fmt.Errorf("failed to find site: %w", err).Error(), http.StatusBadRequest)
		return
	}

	// 3. Update the UserSite record
	if err := h.db.Model(&models.UserSite{}).Where("user_id = ? AND site_id = ?", userID, siteID).Update("state", state).Error; err != nil {
		http.Error(w, fmt.Errorf("failed to find and update request: %w", err).Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("State updated successfully"))
}
