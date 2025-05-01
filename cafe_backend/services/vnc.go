package services

import (
	"fmt"
	"log"
	"net"
	"time"

	"cafe_backend/utils"
	"cafe_backend/models"
)

//@TODO remove hard code vncpassword and device password (text)
func UnlockComputer(computer models.Computer) error {
	vncPassword := "unlock1"
	text := "lantabletxp"
	address := fmt.Sprintf("%s:5900", computer.IPAddress)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("[error] Failed to connect to %s: %v", address, err)
	}
	defer conn.Close()

	log.Printf("[info] Connected to VNC server at %s\n", address)

	if err := utils.PerformHandshake(conn, vncPassword); err != nil {
		return fmt.Errorf("[error] Handshake failed: %v", err)
	}

	log.Println("[info] Sending keystrokes to unlock...")

	// Send each character
	for _, ch := range text {
		if err := utils.SendVNCKey(conn, true, uint32(ch)); err != nil {
			return fmt.Errorf("[error] Failed to send key press: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
		if err := utils.SendVNCKey(conn, false, uint32(ch)); err != nil {
			return fmt.Errorf("[error] Failed to send key release: %v", err)
		}
	}

	if err := utils.SendVNCKey(conn, true, 0xFF0D); err != nil { // XK_Return
		return fmt.Errorf("[error] Failed to press Enter: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	if err := utils.SendVNCKey(conn, false, 0xFF0D); err != nil {
		return fmt.Errorf("[error] Failed to release Enter: %v", err)
	}

	log.Println("[success] Unlock sequence sent successfully.")
	return nil
}
