package services

import (
	// "bytes"
	"cafe_backend/models"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

func connectSSH(computer models.Computer) (*ssh.Client, error) {
	log.Printf("[info] Attempting SSH connection to %s@%s:%d\n", computer.SSHUsername, computer.IPAddress, computer.SSHPort)

	signer, err := ssh.ParsePrivateKey([]byte(computer.SSHPrivateKey))
	if err != nil {
		log.Printf("[warn] Failed to parse private key: %v\n", err)
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
	log.Printf("[info] Dialing SSH at %s\n", address)

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Printf("[warn] SSH dial failed: %v\n", err)
		return nil, err
	}

	log.Println("[success] SSH connection established")
	return client, nil
}

// func UnlockComputer(computer models.Computer) error {
// 	log.Println("[info] Unlocking computer: %s (%s)", computer.Hostname, computer.IPAddress)

// 	client, err := connectSSH(computer)
// 	if err != nil {
// 		return err
// 	}
// 	defer client.Close()

// 	session, err := client.NewSession()
// 	if err != nil {
// 		log.Println("[error] Failed to create SSH session: %v", err)
// 		return err
// 	}
// 	defer session.Close()

// 	log.Println("[info] Running unlock command via SSH...")
// 	var b bytes.Buffer
// 	session.Stdout = &b
// 	err = session.Run("xfce4-screensaver-command --deactivate")
// 	if err != nil {
// 		log.Println("[error] Failed to run unlock command: %v", err)
// 		return err
// 	}

// 	log.Println("[success] Unlock output: %s", b.String())
// 	return nil
// }

func LogoutComputer(computer models.Computer, newPassword string) error {
	currentPassword := computer.CurrentPassword
	log.Printf(currentPassword)
	log.Printf("[info] Logging out user on: %s (%s)\n", computer.Hostname, computer.IPAddress)

	client, err := connectSSH(computer)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Printf("[error] Failed to create SSH session: %v\n", err)
		return err
	}
	defer session.Close()

	log.Println("[info] Running logout command...")
	// cmd := `xfce4-screensaver-command --lock`
	//cmd := `loginctl terminate-user ` + computer.SSHUsername

	// Logout user and cleanup
	cmd := fmt.Sprintf(`
	echo -e "%s\n%s\n%s" | (passwd %s) && \
	loginctl terminate-user %s
	`, currentPassword, newPassword, newPassword, computer.SSHUsername, computer.SSHUsername)

	log.Println(cmd)
	err = session.Run(cmd)
	// loginctl terminate-user will always error!
	//
	// if err != nil {
	// 	log.Printf("[error] Failed to run logout command: %v\n", err)
	// 	return err
	// }

	log.Println("[success] Logout command executed")
	return nil
}
