package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func extractMainDomain(u string) (string, error) {
	//TODO: This function currently does not work with double tlds like .co.uk
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	parts := strings.Split(parsedURL.Hostname(), ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain")
	}

	// Extract the main domain and TLD
	domain := parts[len(parts)-2] // Second to last part is the main domain
	tld := parts[len(parts)-1]    // Last part is the TLD

	return fmt.Sprintf(".%s.%s", domain, tld), nil
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
