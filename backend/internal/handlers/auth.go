package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"time"
)

var jwtKey = []byte("your_secret_key")

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
	storedHash, err := base64.StdEncoding.DecodeString(dbUser.Password)
	if err != nil {
		sendJSONError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	inputHash, err := scrypt.Key([]byte(inputUser.Password), nil, 16384, 8, 1, 32)
	if err != nil || !bytes.Equal(storedHash, inputHash) {
		sendJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": inputUser.Email,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	domain, err := extractMainDomain(r.URL.String())
	// Set the token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,   // Set this to true if using HTTPS
		Domain:   domain, // Adjust to your domain
		Path:     "/",
	})
	siteURL := r.URL.Query().Get("redirect")
	if siteURL == "" {
		http.Error(w, "Redirect URL missing", http.StatusBadRequest)
		return
	} else {
		http.Redirect(w, r, url.QueryEscape(siteURL), http.StatusSeeOther)
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
	// Hash the password using scrypt
	hash, err := scrypt.Key([]byte(user.Password), nil, 16384, 8, 1, 32)
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
func HandleAuthenticate(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// 1. Extract the user's email from the session or JWT token.
	userEmail, err := getUserEmailFromToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// 2. Extract the redirect parameter from the request to get the site URL.
	siteURL := r.URL.Query().Get("redirect")
	if siteURL == "" {
		http.Error(w, "Redirect URL missing", http.StatusBadRequest)
		return
	}

	// 3. Query the database to check if the user has an "authorized" state for the given site.
	var userSite models.UserSite
	err = db.Joins("JOIN users ON users.id = user_sites.user_id").
		Joins("JOIN sites ON sites.id = user_sites.site_id").
		Where("users.email = ? AND sites.url = ? AND user_sites.state = ?", userEmail, siteURL, "authorized").
		First(&userSite).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Redirect to /request if no authorized site is found for the user
			http.Redirect(w, r, "/request?redirect="+url.QueryEscape(siteURL), http.StatusSeeOther)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, siteURL, http.StatusSeeOther)

	// If everything is fine, return a success message.
	//w.Write([]byte("Access granted"))
}

func getUserEmailFromToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return "", fmt.Errorf("Authentication cookie missing")
	}

	tokenStr := cookie.Value
	claims := &jwt.MapClaims{}

	_, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("Invalid token")
	}

	userEmail, ok := (*claims)["user"].(string)
	if !ok {
		return "", fmt.Errorf("Invalid token claims")
	}

	return userEmail, nil
}
