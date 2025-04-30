package services

import (
	"bytes"
	"cafe_backend/models"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

func connectSSH(computer models.Computer) (*ssh.Client, error) {
	log.Println("[info] Attempting SSH connection to %s@%s:%d", computer.SSHUsername, computer.IPAddress, computer.SSHPort)

	signer, err := ssh.ParsePrivateKey([]byte(computer.SSHPrivateKey))
	if err != nil {
		log.Println("[warn] Failed to parse private key: %v", err)
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: computer.SSHUsername,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // not safe for production
		Timeout:         5 * time.Second,
	}

	address := net.JoinHostPort(computer.IPAddress, fmt.Sprintf("%d", computer.SSHPort))
	log.Println("[info] Dialing SSH at %s", address)

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Println("[warn] SSH dial failed: %v", err)
		return nil, err
	}

	log.Println("[success] SSH connection established")
	return client, nil
}

func UnlockComputer(computer models.Computer) error {
	log.Println("[info] Unlocking computer: %s (%s)", computer.Hostname, computer.IPAddress)

	client, err := connectSSH(computer)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Println("[error] Failed to create SSH session: %v", err)
		return err
	}
	defer session.Close()

	log.Println("[info] Running unlock command via SSH...")
	var b bytes.Buffer
	session.Stdout = &b
	err = session.Run("sudo loginctl unlock-sessions")
	if err != nil {
		log.Println("[error] Failed to run unlock command: %v", err)
		return err
	}

	log.Println("[success] Unlock output: %s", b.String())
	return nil
}

func LogoutComputer(computer models.Computer, newPassword string) error {
	log.Println("[info] Logging out user on: %s (%s)", computer.Hostname, computer.IPAddress)

	client, err := connectSSH(computer)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Println("[error] Failed to create SSH session: %v", err)
		return err
	}
	defer session.Close()

	log.Println("[info] Running logout command...")
	cmd := `xfce4-screensaver-command --lock`

	// Logout user and cleanup
	// cmd := `
	// 	echo 'Changing password...'
	// 	echo '` + newPassword + `' | passwd --stdin ` + computer.SSHUsername + `
	// 	loginctl terminate-user ` + computer.SSHUsername + `
	// `

	err = session.Run(cmd)
	if err != nil {
		log.Println("[error] Failed to run logout command: %v", err)
		return err
	}

	log.Println("[success] Logout command executed")
	return nil
}
