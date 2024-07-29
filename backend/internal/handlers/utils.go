package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"gorm.io/gorm"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func extractMainDomain(u string) (string, error) {
	log.Println("Parsing input: " + u)
	// Prepend http:// if no scheme is provided, this ensures url.Parse succeeds
	if !strings.Contains(u, "//") {
		u = "http://" + u
	}

	// Parse the URL and validate it
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %v", err)
	}

	// Split the hostname into parts
	parts := strings.Split(parsedURL.Hostname(), ".")
	partsLength := len(parts)

	// Check if the URL has at least a domain and a TLD
	if partsLength < 2 {
		return "", fmt.Errorf("invalid domain: domain and TLD not found in URL")
	}

	// Handle second-level domains (SLDs) like ".co.uk", ".com.au", etc.
	if partsLength > 2 {
		// List of common SLDs
		secondLevelDomains := map[string]bool{
			"com.au": true,
			"co.uk":  true,
			"com.br": true,
			// ... add more second-level domains as needed
		}

		// Check if the last two parts match a known second-level domain
		if secondLevelDomains[parts[partsLength-2]+"."+parts[partsLength-1]] {
			if partsLength < 3 {
				return "", fmt.Errorf("invalid domain: missing main domain before second-level domain")
			}
			return fmt.Sprintf("%s.%s.%s", parts[partsLength-3], parts[partsLength-2], parts[partsLength-1]), nil
		}
	}

	// For non-SLDs, return the domain and TLD
	return fmt.Sprintf("%s.%s", parts[partsLength-2], parts[partsLength-1]), nil
}

func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
func sendJSONSuccess(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func IsUserAdmin(db *gorm.DB, email string) (bool, error) {
	var user models.User

	if email == "" {
		return false, errors.New("user is empty")
	}
	// Find the user by email
	result := db.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, errors.New("user not found")
	}
	if result.Error != nil {
		return false, result.Error
	}

	// Check if the user's role is "admin"
	return user.Role == "admin", nil
}
