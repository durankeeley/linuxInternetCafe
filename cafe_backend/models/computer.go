package models

import (
	"cafe_backend/database"
	"time"
)

type Computer struct {
	ID               int        `json:"id"`
	Hostname         string     `json:"hostname"`
	IPAddress        string     `json:"ip_address"`
	SSHPort          int        `json:"ssh_port"`
	SSHUsername      string     `json:"ssh_username"`
	SSHPrivateKey    string     `json:"ssh_private_key"`
	VNCPort			 int        `json:"vnc_port"`
	VNCPassword		 string		`json:"vnc_password"`
	CurrentPassword  string    `json:"current_password,omitempty"`
	SessionExpiresAt *time.Time `json:"session_expires_at,omitempty"`
	Assigned         *string    `json:"assigned,omitempty"`
}

func GetComputerByID(id int) (Computer, error) {
	var c Computer
	row := database.DB.QueryRow(`
		SELECT id, hostname, ip_address, ssh_port, ssh_username, ssh_private_key, vnc_port, vnc_password, current_password, session_expires_at, assigned
		FROM computers WHERE id = ?`, id)
	err := row.Scan(&c.ID, &c.Hostname, &c.IPAddress, &c.SSHPort, &c.SSHUsername, &c.SSHPrivateKey, &c.VNCPort, &c.VNCPassword, &c.CurrentPassword, &c.SessionExpiresAt, &c.Assigned)
	return c, err
}
