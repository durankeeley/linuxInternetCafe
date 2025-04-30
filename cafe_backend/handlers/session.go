package handlers

import (
	"cafe_backend/database"
	"cafe_backend/models"
	"cafe_backend/services"
	"cafe_backend/utils"
	"encoding/json"
	"net/http"
	"time"
)

func StartSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ComputerID int `json:"computer_id"`
		Minutes    int `json:"minutes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Set assigned to "active" and calculate session expiry
	expiry := time.Now().Add(utils.MinutesToDuration(req.Minutes))

	_, err := database.DB.Exec(`
		UPDATE computers
		SET assigned = 'active', session_expires_at = ?
		WHERE id = ?
	`, expiry, req.ComputerID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Unlock session via SSH
	comp, err := models.GetComputerByID(req.ComputerID)
	if err != nil {
		http.Error(w, "Computer not found", http.StatusNotFound)
		return
	}
	err = services.UnlockComputer(comp)

	json.NewEncoder(w).Encode(map[string]string{"message": "Session started", "expires_at": expiry.Format(time.RFC3339)})
}

func EndSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ComputerID int `json:"computer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Logout and reset computer
	comp, err := models.GetComputerByID(req.ComputerID)
	if err != nil {
		http.Error(w, "Computer not found", http.StatusNotFound)
		return
	}

	_, err = database.DB.Exec(`
	UPDATE computers
	SET assigned = null
	WHERE id = ?
	`, req.ComputerID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	newPassword := utils.GenerateRandomPassword(12)
	err = services.LogoutComputer(comp, newPassword)

	json.NewEncoder(w).Encode(map[string]string{"message": "Session ended"})
}

func UnlockSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ComputerID int `json:"computer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	comp, err := models.GetComputerByID(req.ComputerID)

	err = services.UnlockComputer(comp)
	if err != nil {
		http.Error(w, "Failed to unlock: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Session unlocked"})
}

func ExtendSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ComputerID int `json:"computer_id"`
		Minutes    int `json:"minutes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`
		UPDATE computers
		SET session_expires_at = DATETIME(session_expires_at, ? || ' minutes')
		WHERE id = ?
	`, req.Minutes, req.ComputerID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Session extended"})
}

func NotifySession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ComputerID   int    `json:"computer_id"`
		Notification string `json:"notification"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// If no notification passed, set a default
	if req.Notification == "" {
		req.Notification = "⚠️ 1 minute remaining on your session."
	}

	comp, err := models.GetComputerByID(req.ComputerID)

	err = services.NotifyComputer(comp, req.Notification)
	if err != nil {
		http.Error(w, "Failed to send notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Notification sent"})
}
