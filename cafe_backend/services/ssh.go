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
	fmt.Printf("[info] Attempting SSH connection to %s@%s:%d\n", computer.SSHUsername, computer.IPAddress, computer.SSHPort)

	signer, err := ssh.ParsePrivateKey([]byte(computer.SSHPrivateKey))
	if err != nil {
		fmt.Printf("[warn] Failed to parse private key: %v\n", err)
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
	fmt.Printf("[info] Dialing SSH at %s\n", address)

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		fmt.Printf("[warn] SSH dial failed: %v\n", err)
		return nil, err
	}

	fmt.Println("[success] SSH connection established")
	return client, nil
}

func UnlockComputer(computer models.Computer) error {
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

	var b bytes.Buffer
	session.Stdout = &b
	err = session.Run("loginctl unlock-sessions")
	if err != nil {
		return err
	}
	log.Println("Unlock output:", b.String())
	return nil
}

func LogoutComputer(computer models.Computer, newPassword string) error {
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

	// Logout user and cleanup
	cmd := `
		echo 'Changing password...'
		echo '` + newPassword + `' | passwd --stdin ` + computer.SSHUsername + `
		loginctl terminate-user ` + computer.SSHUsername + `
	`
	return session.Run(cmd)
}
