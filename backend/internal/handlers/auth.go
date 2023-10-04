package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"github.com/B-Urb/KubeVoyage/internal/util"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Handler struct {
	db      *gorm.DB
	JWTKey  []byte
	BaseURL string
}

func NewHandler(db *gorm.DB) *Handler {
	jwtKey, err := util.GetEnvOrError("JWT_SECRET_KEY")
	if err != nil {
		log.Fatalf("Error reading JWT_SECRET_KEY: %v", err)
	}

	baseURL, err := util.GetEnvOrError("BASE_URL")
	if err != nil {
		log.Fatalf("Error reading BASE_URL: %v", err)
	}
	return &Handler{db: db, JWTKey: []byte(jwtKey), BaseURL: baseURL}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var inputUser models.User
	var dbUser models.User

	// Parse the request body
	err := json.NewDecoder(r.Body).Decode(&inputUser)
	if err != nil {
		sendJSONError(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Fetch the user from the database
	result := h.db.Where("email = ?", inputUser.Email).First(&dbUser)
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
	tokenString, err := token.SignedString(h.JWTKey)
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
	_, err = w.Write([]byte("Login successful"))
	if err != nil {
		return
	}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
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
	if err := h.db.Where("email = ?", user.Email).First(&existingUser).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			sendJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		sendJSONError(w, "User already exists", http.StatusConflict)
		return
	}
	// Save the user to the database
	result := h.db.Create(&user)
	if result.Error != nil {
		sendJSONError(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONSuccess(w, "", http.StatusCreated)
}
func (h *Handler) HandleAuthenticate(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the user's email from the session or JWT token.
	userEmail, err := h.getUserEmailFromToken(r)
	if err != nil {
		h.logError(w, "Failed to get user email from token", err, http.StatusUnauthorized)
		return
	}

	// 2. Extract the redirect parameter from the request to get the site URL.
	siteURL := r.Header.Get("X-Forwarded-Uri")
	if siteURL == "" {
		siteURL = r.URL.Query().Get("redirect")
		if siteURL == "" {
			h.logError(w, "Redirect URL missing from both header and URL parameter", nil, http.StatusBadRequest)
			return
		}
	}

	// 3. Query the database to check if the user has an "authorized" state for the given site.
	var userSite models.UserSite
	err = h.db.Joins("JOIN users ON users.id = user_sites.user_id").
		Joins("JOIN sites ON sites.id = user_sites.site_id").
		Where("users.email = ? AND sites.url = ? AND user_sites.state = ?", userEmail, siteURL, "authorized").
		First(&userSite).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Redirect to /request if no authorized site is found for the user
			http.Redirect(w, r, "/request?redirect="+url.QueryEscape(siteURL), http.StatusSeeOther)
			return
		}
		h.logError(w, "Database error while checking user authorization", err, http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, siteURL, http.StatusSeeOther)
}

func (h *Handler) logError(w http.ResponseWriter, message string, err error, statusCode int) {
	logMessage := message
	if err != nil {
		logMessage = fmt.Sprintf("%s: %v", message, err)
	}
	log.Println(logMessage)
	http.Error(w, message, statusCode)
}

func (h *Handler) getUserEmailFromToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return "", fmt.Errorf("Authentication cookie missing")
	}

	tokenStr := cookie.Value
	claims := &jwt.MapClaims{}

	_, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return h.JWTKey, nil
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
