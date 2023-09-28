package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func HandleLogin(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var inputUser models.User
	var dbUser models.User

	// Parse the request body
	err := json.NewDecoder(r.Body).Decode(&inputUser)
	if err != nil {
		sendJSONError(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Fetch the user from the database
	result := db.Where("email = ?", inputUser.Email).First(&dbUser)
	if result.Error != nil {
		sendJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	// Compare the password hash
	hash, err := base64.StdEncoding.DecodeString(dbUser.Password)
	if err != nil {
		sendJSONError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	inputHash, err := scrypt.Key([]byte(inputUser.Password), hash[:8], 16384, 8, 1, 32)
	if err != nil || !bytes.Equal(hash, inputHash) {
		sendJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Here, you'd typically generate a JWT or session token and send it back to the client.
	// For simplicity, we'll just send a success message.
	w.Write([]byte("Login successful"))
}

func HandleRegister(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var user models.User

	// Parse the request body
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Hash the password using scrypt
	salt := make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hash, err := scrypt.Key([]byte(user.Password), salt, 16384, 8, 1, 32)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.Password = base64.StdEncoding.EncodeToString(hash)
	var existingUser models.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			sendJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		sendJSONError(w, "User already exists", http.StatusConflict)
		return
	}
	// Save the user to the database
	result := db.Create(&user)
	if result.Error != nil {
		sendJSONError(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONSuccess(w, "", http.StatusCreated)
}

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
