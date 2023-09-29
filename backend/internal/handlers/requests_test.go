package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleRegister(t *testing.T) {
	// Create a mock HTTP request
	body := bytes.NewBuffer([]byte(`{"email": "test@example.com", "password": "password123"}`))
	req, err := http.NewRequest("POST", "/api/register", body)
	if err != nil {
		t.Fatal(err)
	}

	// Record the HTTP response
	rr := httptest.NewRecorder()

	// Assuming you have a function to set up a test database
	db := setupTestDatabase()

	app := &App{DB: db}
	handler := http.HandlerFunc(app.HandleRegister)

	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201")
	// Add more assertions as needed
}
