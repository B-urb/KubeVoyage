package handlers

import (
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/url"
)

func handleUnauthenticated(w http.ResponseWriter, r *http.Request) {
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		redirectURL = "/" // default URL if no redirect was provided
	} else {
		redirectURL = "/?redirect=" + redirectURL
	}
	http.Redirect(w, r, url.QueryEscape(redirectURL), http.StatusSeeOther)
}

func (h *Handler) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "Authentication cookie missing", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value
		claims := &jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return h.JWTKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
