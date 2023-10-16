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
	siteURL, err := h.getRedirectFromCookie(r, w)
	if siteURL == "" {
		http.Error(w, "Redirect URL missing", http.StatusBadRequest)
		return
	} else {
		http.Redirect(w, r, siteURL, http.StatusSeeOther)
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
	siteURL, err := h.getRedirectUrl(r)
	if err != nil {
		h.logError(w, err.Error(), nil, http.StatusBadRequest)
		return
	}
	// 1. Extract the user's email from the session or JWT token.
	userEmail, err := h.getUserEmailFromToken(r)
	if err != nil {
		// If the user cannot be read from the cookie, redirect to /login with the site URL as a parameter
		h.setRedirectCookie(siteURL, r, w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Check if the user has the role "admin"
	var user models.User
	err = h.db.Where("email = ?", userEmail).First(&user).Error
	if err != nil {
		h.logError(w, "Database error while fetching user details", err, http.StatusInternalServerError)
		return
	}
	if user.Role == "admin" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. Query the database to check if the user has an "authorized" state for the given site.
	var userSite models.UserSite
	err = h.db.Joins("JOIN users ON users.id = user_sites.user_id").
		Joins("JOIN sites ON sites.id = user_sites.site_id").
		Where("users.email = ? AND sites.url = ? AND user_sites.state = ?", userEmail, siteURL, "authorized").
		First(&userSite).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return 401 if the user is not authorized for the requested siteURL
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.logError(w, "Database error while checking user authorization", err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
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

func (h *Handler) setRedirectCookie(redirectUrl string, r *http.Request, w http.ResponseWriter) error {
	domain, err := extractMainDomain(r.URL.String())
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(domain)
	http.SetCookie(w, &http.Cookie{
		Name:     "auth-site",
		Value:    redirectUrl,
		Expires:  time.Now().Add(15 * time.Minute), // Shorter duration
		HttpOnly: true,
		Secure:   true,                  // Set this to true if using HTTPS
		SameSite: http.SameSiteNoneMode, // Set this to true if using HTTPS
		Domain:   domain,                // Adjust to your domain
		Path:     "/",
	})
	return nil
}
func (h *Handler) getRedirectFromCookie(r *http.Request, w http.ResponseWriter) (string, error) {
	cookie, err := r.Cookie("auth-site")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			// No cookie found
			return "", nil
		}
		return "", err
	}

	// Clear the cookie once it's read
	//http.SetCookie(w, &http.Cookie{
	//	Name:    "auth-site",
	//	Value:   "",
	//	Expires: time.Unix(0, 0),
	//	Path:    "/",
	//})

	return cookie.Value, nil
}
func (h *Handler) getRedirectUrl(r *http.Request) (string, error) {
	// Extract the redirect parameter from the request to get the site URL.

	siteURL := r.Header.Get("X-Forwarded-Uri")
	if siteURL == "" {
		siteURL = r.Header.Get("Referer")
	}
	if siteURL == "" {
		siteURL = r.URL.Query().Get("redirect")
	}
	if siteURL == "" {
		printHeaders(r)
		return "", fmt.Errorf("Redirect URL missing from both header and URL parameter")
	}
	return siteURL, nil
}
func printHeaders(r *http.Request) {
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}
}
