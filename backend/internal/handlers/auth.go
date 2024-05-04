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

var store = sessions.NewCookieStore([]byte("your-very-secret-key"))
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": inputUser.Email,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	_, err = token.SignedString(h.JWTKey)
	if err != nil {
		sendJSONError(w, "Internal Server Error", http.StatusInternalServerError)
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
		MaxAge:   3600,                  // Expires after 1 hour
		HttpOnly: true,                  // Not accessible via JavaScript
		Secure:   true,                  // Only sent over HTTPS
		SameSite: http.SameSiteNoneMode, // Controls cross-site request behavior
		Domain:   tld,
	}
	session.Values["authenticated"] = true
	session.Values["user"] = inputUser.Email
	session.Save(r, w)

	var domain string
	siteURL, siteUrlErr := h.getRedirectUrl(r, w)
	if siteUrlErr != nil {
		// If there was an error getting the redirect URL, use the request's host as the domain
		log.Println("Site URl could not be determined: " + siteURL)
		domain = r.Host
	} else {
		// If the redirect URL was obtained successfully, extract the main domain
		h.setRedirectCookie(siteURL, r, w)
		var err error
		domain, err = extractMainDomain(r.Host)
		if err != nil {
			sendJSONError(w, "Invalid Redirect URL", http.StatusBadRequest)
			return
		}
	}
	if err != nil {
		slog.Error("could not extract Main Domain", err)
		return
	}
	slog.Info("Domain: ", domain)
	// Set the token as a cookie
	//http.SetCookie(w, &http.Cookie{
	//	Name:     "X-Auth-Token",
	//	Value:    tokenString,
	//	Expires:  time.Now().Add(24 * time.Hour),
	//	Secure:   true,                  // Set this to true if using HTTPS
	//	SameSite: http.SameSiteNoneMode, // Set this to true if using HTTPS
	//	Domain:   r.Host,                // Adjust to your domain
	//	Path:     "/",
	//})

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
	siteURL, err := h.getRedirectFromCookie(r, w, true)
	if err != nil {

	}
	if siteURL == "" {
		siteURL = r.Host
	}

	redirect := r.Header.Get("X-Auth-Site")
	log.Println(redirect)
	log.Println(siteURL)
	http.Redirect(w, r, siteURL, http.StatusSeeOther)

}
func (h *Handler) HandleAuthenticate(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the user's email from the session or JWT token.
	siteURL, err := h.getRedirectUrl(r, w)
	if err != nil {
		log.Println(err.Error())
		//h.logError(w, err.Error(), nil, http.StatusBadRequest)
		//return
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
		h.setRedirectCookie(siteURL, r, w) //Fixme: improve domain handling
		http.Redirect(w, r, "/login?redirect="+strings.TrimSuffix(siteURL, "/")+"&token="+oneTimeToken, http.StatusSeeOther)
		return
	}
	if tokenAuthenticated {
		session.Values["authenticated"] = true
		session.Values["user"] = oneTimeStore[token].user
		session.Save(r, w)
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

func (h *Handler) logError(w http.ResponseWriter, message string, err error, statusCode int) {
	logMessage := message
	if err != nil {
		logMessage = fmt.Sprintf("%s: %v", message, err)
	}
	log.Println(logMessage)
	http.Error(w, message, statusCode)
}

func (h *Handler) getUserEmailFromToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("X-Auth-Token")
	if err != nil {
		slog.Error("Authentication Cookie missing", err)
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
func (h *Handler) getRedirectFromCookie(r *http.Request, w http.ResponseWriter, clear bool) (string, error) {
	cookie, err := r.Cookie("X-Auth-Site")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			// No cookie found
			return "", nil
		}
		return "", err
	}

	// Clear the cookie once it's read
	//http.SetCookie(w, &http.Cookie{
	//	Name:    "X-Auth-Site",
	//	Value:   "",
	//	Expires: time.Unix(0, 0),
	//	Path:    "/",
	//})

	return cookie.Value, nil
}
func (h *Handler) getRedirectUrl(r *http.Request, w http.ResponseWriter) (string, error) {
	// Extract the redirect parameter from the request to get the site URL.

	siteURL := r.Header.Get("X-Forwarded-Uri")
	if siteURL == "" {
		siteURL = r.Header.Get("X-Auth-Site")
		if siteURL == "" {
			siteURL = r.URL.Query().Get("redirect")
			if siteURL == "" {
				surl, err := h.getRedirectFromCookie(r, w, false)
				if err != nil {
					return "", fmt.Errorf("Redirect URL missing from both header and URL parameter")
				}
				siteURL = surl
			}
		}
	}
	if siteURL == "" {
		return "", fmt.Errorf("Redirect URL missing from both header and URL parameter")
	} else {
		return siteURL, nil
	}
}

// generateSessionID generates a secure, random session ID.
func generateSessionID() string {
	size := 32 // This size can be adjusted according to your security needs
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatalf("Failed to generate random session ID: %v", err)
	}
	return hex.EncodeToString(bytes)
}
func logCookies(r *http.Request) {
	cookies := r.Cookies() // Step 1: Retrieve all cookies
	var cookieStrings []string

	for _, cookie := range cookies {
		// Step 2: Format each cookie as a string
		cookieStr := fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
		cookieStrings = append(cookieStrings, cookieStr)
	}

	// Join all cookie strings into a single string
	allCookies := strings.Join(cookieStrings, "; ")

	// Step 3: Log the complete cookie string
	slog.Info("Cookies:", allCookies)
}
func printHeaders(r *http.Request) {
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}
}
