package handlers

import (
	"cafe_backend/services"
	"cafe_backend/utils"
	"encoding/json"
	"net/http"
	"time"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	valid, err := services.Authenticate(creds.Username, creds.Password)
	if err != nil || !valid {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, expirationTime, err := utils.GenerateJWT(creds.Username)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      token,
		"expires":    expirationTime.Unix(),               // Unix timestamp
		"expiresUTC": expirationTime.Format(time.RFC3339), // RFC3339 UTC
	})
}
