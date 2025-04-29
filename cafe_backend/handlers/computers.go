package handlers

import (
	"cafe_backend/database"
	"cafe_backend/models"
	"cafe_backend/utils"
	"encoding/json"
	"log"
	"net/http"
)

func AddComputerHandler(w http.ResponseWriter, r *http.Request) {
	var c models.Computer
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`
		INSERT INTO computers (hostname, ip_address, ssh_port, ssh_username, ssh_private_key)
		VALUES (?, ?, ?, ?, ?)
	`, c.Hostname, c.IPAddress, c.SSHPort, c.SSHUsername, c.SSHPrivateKey)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Computer added successfully"})
}

func GetComputersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := database.DB.Query(`
		SELECT id, hostname, ip_address, ssh_port, ssh_username, current_password, assigned
		FROM computers
	`)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var computers []map[string]interface{}

	for rows.Next() {
		var id, sshPort int
		var hostname, ip, sshUser string
		var assigned *string
		var currentPassword *string

		if err := rows.Scan(&id, &hostname, &ip, &sshPort, &sshUser, &currentPassword, &assigned); err != nil {
			log.Println("Row scan error:", err)
			continue
		}

		status := utils.Ping(ip)

		computer := map[string]interface{}{
			"id":       id,
			"hostname": hostname,
			"ip":       ip,
			"ssh_port": sshPort,
			"ssh_user": sshUser,
			"assigned": assigned,
			"status":   status,
		}

		if currentPassword != nil && *currentPassword != "" {
			computer["current_password"] = *currentPassword
		}

		if assigned != nil && *assigned != "" {
			computer["assigned"] = *assigned
		}

		computers = append(computers, computer)
	}

	log.Printf("Returning %d computers\n", len(computers))
	json.NewEncoder(w).Encode(computers)
}
