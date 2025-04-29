package services

import (
	"cafe_backend/database"
	"cafe_backend/models"
	"cafe_backend/utils"
	"log"
	"time"
)

func WatchSessions() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		rows, err := database.DB.Query(`
			SELECT id, session_expires_at, hostname, ip_address, ssh_port, ssh_username, ssh_private_key
			FROM computers
			WHERE assigned = 'active'
		`)
		if err != nil {
			log.Println("Error querying sessions:", err)
			continue
		}

		for rows.Next() {
			var id int
			var expiresAtStr string
			var hostname, ip, sshUsername, sshPrivateKey string
			var sshPort int

			err := rows.Scan(&id, &expiresAtStr, &hostname, &ip, &sshPort, &sshUsername, &sshPrivateKey)
			if err != nil {
				log.Println("Error scanning row:", err)
				continue
			}

			expiresAt, _ := time.Parse(time.RFC3339, expiresAtStr)
			remaining := time.Until(expiresAt)

			computer := models.Computer{
				ID:            id,
				Hostname:      hostname,
				IPAddress:     ip,
				SSHPort:       sshPort,
				SSHUsername:   sshUsername,
				SSHPrivateKey: sshPrivateKey,
			}

			if remaining <= time.Minute && remaining > 0 {
				err := NotifyComputer(computer, "1 minute remaining!")
				if err != nil {
					log.Println("Notify error:", err)
				}
			} else if remaining <= 0 {
				// Time expired - log out
				newPassword := utils.GenerateRandomPassword(12)
				err := LogoutComputer(computer, newPassword)
				if err != nil {
					log.Println("Logout error:", err)
				}

				// Reset DB
				_, err = database.DB.Exec(`UPDATE computers SET assigned = NULL, current_password = ? WHERE id = ?`, newPassword, id)
				if err != nil {
					log.Println("DB update error:", err)
				}
			}
		}
		rows.Close()
	}
}

func NotifyComputer(computer models.Computer, message string) error {
	client, err := connectSSH(computer)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := `DISPLAY=:0 DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus notify-send "Session Warning" "` + message + `"`
	return session.Run(cmd)
}
