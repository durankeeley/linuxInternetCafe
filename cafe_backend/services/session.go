package services

import (
	"cafe_backend/database"
	"cafe_backend/models"
	"cafe_backend/utils"
	"log"
	"time"
	"fmt"
)

func WatchSessions() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		rows, err := database.DB.Query(`
			SELECT id, session_expires_at, hostname, ip_address, ssh_port, ssh_username, ssh_private_key, current_password
			FROM computers
			WHERE assigned = 'active'
		`)
		if err != nil {
			log.Println("Error querying sessions:", err)
			continue
		}

		for rows.Next() {
			var id, sshPort int
			var hostname, ip, sshUsername, sshPrivateKey, expiresAtStr, currentPassword string

			err := rows.Scan(&id, &expiresAtStr, &hostname, &ip, &sshPort, &sshUsername, &sshPrivateKey, &currentPassword)
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
				CurrentPassword: currentPassword,
			}

			if remaining <= time.Minute && remaining > 0 {
				fmt.Printf("[info] %s 1 minute remaining\n", computer.Hostname)
				err := NotifyComputer(computer, "1 minute remaining!")
				if err != nil {
					log.Println("[warn] Notify error:", err)
				}
			} else if remaining <= 0 {
				// Time expired - log out
				newPassword := utils.GenerateRandomPassword(12)
				//newPassword := "lantabletxp"
				fmt.Printf("[info] %s session has ended\n", computer.Hostname)
				err := LogoutComputer(computer, newPassword)
				if err != nil {
					// terminate user will always error
					//log.Println("[error] Logout error:", err)
				}

				// Reset DB
				_, err = database.DB.Exec(`UPDATE computers SET assigned = NULL, session_expires_at = NULL, current_password = ? WHERE id = ?`, newPassword, id)
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
