package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/B-Urb/KubeVoyage/internal/models"
	"github.com/B-Urb/KubeVoyage/internal/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	db      *gorm.DB
	JWTKey  []byte
	BaseURL string
}

type TokenInfo struct {
	authenticated bool
	user          string
}

var secret, _ = util.GetEnvOrDefault("JWT_SECRET_KEY", "kubevoyage")
var store = sessions.NewCookieStore([]byte(secret))
var oneTimeStore = make(map[string]TokenInfo)

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

type LoginResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Redirect    bool   `json:"redirect"`
	RedirectURL string `json:"redirect_url,omitempty"`
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

	session, _ := store.Get(r, "session-cook")
	tld, err := extractMainDomain(r.Host)
	if err != nil {
		sendJSONError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:     "/",                   // Available across the entire domain
		MaxAge:   3600 * 24 * 7,         // Expires after 1 week TODO: make configurable
		HttpOnly: true,                  // Not accessible via JavaScript
		Secure:   true,                  // Only sent over HTTPS
		SameSite: http.SameSiteNoneMode, // Controls cross-site request behavior
		Domain:   tld,
	}
	session.Values["authenticated"] = true
	session.Values["user"] = inputUser.Email
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var domain string
	siteURL, siteUrlErr := h.getRedirectUrl(r)
	if siteUrlErr != nil {
		// If there was an error getting the redirect URL, use the request's host as the domain
		log.Println("Site URl could not be determined: " + siteURL)
		domain = r.Host
	} else {
		// If the redirect URL was obtained successfully, extract the main domain
		err := h.setRedirectCookie(siteURL, r, w)
		if err != nil {
			slog.Error("Failed to set redirect cookie", "error", err)
		}

		domain, err = extractMainDomain(r.Host)
		if err != nil {
			sendJSONError(w, "Invalid Redirect URL", http.StatusBadRequest)
			return
		}
	}
	if err != nil {
		slog.Error("could not extract Main Domain", "error", err)
		return
	}
	slog.Info("Domain: ", "value", domain)

	response := LoginResponse{
		Success:  true,
		Message:  "Login successful",
		Redirect: siteURL != "" && siteUrlErr == nil,
	}
	oneTimeToken := r.URL.Query().Get("token")
	if oneTimeToken == "" {
		return
	}
	oneTimeStore[oneTimeToken] = TokenInfo{true, inputUser.Email}
	sendJSONResponse(w, response, http.StatusOK)
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
func (h *Handler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	//FIXME: Not unchecked redirecting with parameter
	siteURL, err := h.getRedirectFromCookie(r, true)
	if err != nil {

	}
	if siteURL == "" {
		siteURL = r.Host
	}

	http.Redirect(w, r, siteURL, http.StatusSeeOther)

}
func (h *Handler) HandleAuthenticate(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the user's email from the session or JWT token.
	siteURL, err := h.getRedirectUrl(r)
	if err != nil {
	}
	tld, err := extractMainDomain(siteURL)
	session, err := store.Get(r, "session-cook")
	// Check if "authenticated" is set and true in the session
	auth, ok := session.Values["authenticated"].(bool)
	token, ok := session.Values["oneTimeToken"].(string)
	tokenAuthenticated := oneTimeStore[token].authenticated
	tokenUser := oneTimeStore[token].user

	if !ok || (!auth && !tokenAuthenticated) {
		session.Options = &sessions.Options{
			Path:     "/",                   // Available across the entire domain
			MaxAge:   3600,                  // Expires after 1 hour
			HttpOnly: true,                  // Not accessible via JavaScript
			Secure:   true,                  // Only sent over HTTPS
			SameSite: http.SameSiteNoneMode, // Controls cross-site request behavior
			Domain:   tld,
		}

		// Generate a new random session ID
		session.ID = generateSessionID()

		// Set some initial values
		session.Values["authenticated"] = false
		oneTimeToken := generateSessionID()
		oneTimeStore[oneTimeToken] = TokenInfo{false, ""}
		session.Values["oneTimeToken"] = oneTimeToken
		if err := session.Save(r, w); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// If the user cannot be read from the cookie, redirect to /login with the site URL as a parameter
		err = h.setRedirectCookie(siteURL, r, w) //Fixme: improve domain handling
		if err != nil {
			slog.Error("failed to set redirect cookie", "error", err)
		}
		http.Redirect(w, r, "/login?redirect="+strings.TrimSuffix(siteURL, "/")+"&token="+oneTimeToken, http.StatusSeeOther)
		return
	}
	if tokenAuthenticated {
		session.Values["authenticated"] = true
		session.Values["user"] = oneTimeStore[token].user
		err = session.Save(r, w)
		if err != nil {
			slog.Error("Failed to save session", "error", err)
		}
		delete(oneTimeStore, token)
	}
	slog.Debug("Incoming session is authenticated")
	sessionUser, ok := session.Values["user"].(string)
	if sessionUser == "" {
		sessionUser = tokenUser
	}
	if sessionUser == "" {
		h.logError(w, "error while fetching user details from session", err, http.StatusInternalServerError)
	}

	// Check if the user has the role "admin"
	var user models.User
	err = h.db.Where("email = ?", sessionUser).First(&user).Error
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
		Where("users.email = ? AND sites.url = ?", sessionUser, siteURL).
		First(&userSite).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return 401 if the user is not authorized for the requested siteURL
			http.Redirect(w, r, "/request?redirect="+siteURL, http.StatusSeeOther)
			return
		}
		h.logError(w, "Database error while checking user authorization", err, http.StatusInternalServerError)
		return
	}
	if userSite.State == models.Requested || userSite.State == models.Declined {
		w.WriteHeader(http.StatusUnauthorized)
	}
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-cook")
	if err != nil {
		sendJSONError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Clear session values
	session.Values["authenticated"] = false

	// Expire the cookie
	session.Options.MaxAge = -1
	tld, err := extractMainDomain(r.Host)
	session.Options.Domain = tld

	// Save the session
	err = session.Save(r, w)
	if err != nil {
		sendJSONError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	}
	sendJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) HandleValidateSession(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-cook")
	if err != nil {
		sendJSONError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	authenticated, ok := session.Values["authenticated"].(bool)
	if !ok || !authenticated {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, ok := session.Values["user"].(string)
	if !ok || user == "" {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Optionally, you can check if the user still exists in the database
	var dbUser models.User
	result := h.db.Where("email = ?", user).First(&dbUser)
	if result.Error != nil {
		sendJSONError(w, "User not found", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Session is valid",
		"user":    user,
	}
	sendJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) logError(w http.ResponseWriter, message string, err error, statusCode int) {
	logMessage := message
	if err != nil {
		logMessage = fmt.Sprintf("%s: %v", message, err)
	}
	log.Println(logMessage)
	http.Error(w, message, statusCode)
}

func (h *Handler) getUserFromSession(r *http.Request) (string, error) {
	session, err := store.Get(r, "session-cook")
	if err != nil {
		slog.Debug("Error retrieving user from session", "error", err)
		return "", err
	}
	user, success := session.Values["user"].(string)
	if success == false {
		slog.Debug("Error retrieving user from session", "error", err)
		return "", err
	}
	return user, nil
}

func (h *Handler) getUserEmailFromToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session-cook")
	if err != nil {
		slog.Error("Authentication Cookie missing", "error", err)
		return "", fmt.Errorf("authentication cookie missing")
	}

	tokenStr := cookie.Value
	claims := &jwt.MapClaims{}

	_, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return h.JWTKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token")
	}

	userEmail, ok := (*claims)["user"].(string)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	return userEmail, nil
}

func (h *Handler) setRedirectCookie(redirectUrl string, r *http.Request, w http.ResponseWriter) error {
	w.Header().Set("X-Auth-Site", redirectUrl)
	domain, err := extractMainDomain(redirectUrl)
	if err != nil {
		log.Println(err.Error())
		log.Println(domain)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "X-Auth-Site",
		Value:    redirectUrl,
		Expires:  time.Now().Add(15 * time.Minute), // Shorter duration
		HttpOnly: true,
		Secure:   true,                  // Set this to true if using HTTPS
		SameSite: http.SameSiteNoneMode, // Set this to true if using HTTPS
		Domain:   r.Host,                // Adjust to your domain
		Path:     "/",
	})
	return nil
}
func (h *Handler) getRedirectFromCookie(r *http.Request, clear bool) (string, error) {
	cookie, err := r.Cookie("X-Auth-Site")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			// No cookie found
			return "", nil
		}
		return "", err
	}
	if clear {
	}
	return cookie.Value, nil
}
func (h *Handler) getRedirectUrl(r *http.Request) (string, error) {
	// Extract the redirect parameter from the request to get the site URL.
	siteURL := r.URL.Query().Get("redirect")
	if siteURL == "" || siteURL == "null" {
		return "", fmt.Errorf("redirect URL missing from both header and URL parameter")
	} else {
		return siteURL, nil
	}
}

// generateSessionID generates a secure, random session ID.
func generateSessionID() string {
	size := 32 // This size can be adjusted according to your security needs
	randomBytes := make([]byte, size)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatalf("Failed to generate random session ID: %v", err)
	}
	return hex.EncodeToString(randomBytes)
}
