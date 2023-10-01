package main

import (
	"github.com/B-Urb/KubeVoyage/internal/models"
)

func generateTestData() {
	// Insert test data for Users
	users := []models.User{
		{Email: "user1@example.com", Password: "password1", Role: "admin"},
		{Email: "user2@example.com", Password: "password2", Role: "user"},
		{Email: "user3@example.com", Password: "password3", Role: "user"},
	}
	for _, user := range users {
		db.Create(&user)
	}

	// Insert test data for Sites
	sites := []models.Site{
		{URL: "https://site1.com"},
		{URL: "https://site2.com"},
		{URL: "https://site3.com"},
	}
	for _, site := range sites {
		db.Create(&site)
	}

	// Insert test data for UserSite
	userSites := []models.UserSite{
		{UserID: 1, SiteID: 1, State: "authorized"},
		{UserID: 2, SiteID: 2, State: "requested"},
		{UserID: 3, SiteID: 3, State: "authorized"},
	}
	for _, userSite := range userSites {
		db.Create(&userSite)
	}
}
