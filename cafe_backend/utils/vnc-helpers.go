package utils

import (
	"crypto/des"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func PerformHandshake(conn net.Conn, password string) error {
	// Read server version
	serverVersion := make([]byte, 12)
	if _, err := conn.Read(serverVersion); err != nil {
		return fmt.Errorf("failed to read server version: %v", err)
	}
	clientVersion := []byte("RFB 003.008\n")
	if _, err := conn.Write(clientVersion); err != nil {
		return fmt.Errorf("failed to send client version: %v", err)
	}

	// Read available security types
	var numSecTypes uint8
	if err := binary.Read(conn, binary.BigEndian, &numSecTypes); err != nil {
		return fmt.Errorf("failed to read security types count: %v", err)
	}
	secTypes := make([]byte, numSecTypes)
	if _, err := conn.Read(secTypes); err != nil {
		return fmt.Errorf("failed to read security types: %v", err)
	}

	supported := false
	for _, t := range secTypes {
		if t == 2 {
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf("VNC authentication (type 2) not supported by server")
	}
	if _, err := conn.Write([]byte{2}); err != nil {
		return fmt.Errorf("failed to select VNC auth type: %v", err)
	}

	// Authenticate with password
	if err := vncAuthenticateWithPassword(conn, password); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	// Send ClientInit
	if _, err := conn.Write([]byte{1}); err != nil {
		return fmt.Errorf("failed to send ClientInit: %v", err)
	}

	// Skip reading ServerInit
	dummy := make([]byte, 24)
	if _, err := io.ReadFull(conn, dummy); err != nil {
		return fmt.Errorf("failed to read ServerInit: %v", err)
	}

	return nil
}

func vncAuthenticateWithPassword(conn net.Conn, password string) error {
	challenge := make([]byte, 16)
	if _, err := io.ReadFull(conn, challenge); err != nil {
		return fmt.Errorf("failed to read challenge: %v", err)
	}

	key := prepareVNCPasswordKey(password)
	block, err := des.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create DES cipher: %v", err)
	}

	response := make([]byte, 16)
	for i := 0; i < 2; i++ {
		block.Encrypt(response[i*8:(i+1)*8], challenge[i*8:(i+1)*8])
	}
	if _, err := conn.Write(response); err != nil {
		return fmt.Errorf("failed to send response: %v", err)
	}

	var result uint32
	if err := binary.Read(conn, binary.BigEndian, &result); err != nil {
		return fmt.Errorf("failed to read auth result: %v", err)
	}
	if result != 0 {
		return fmt.Errorf("authentication failed with code: %d", result)
	}
	return nil
}

func prepareVNCPasswordKey(password string) []byte {
	key := make([]byte, 8)
	copy(key, password)
	for i := range key {
		key[i] = reverseByteBitsForVNC(key[i])
	}
	return key
}

func reverseByteBitsForVNC(b byte) byte {
	var result byte
	for i := 0; i < 8; i++ {
		if b&(1<<i) != 0 {
			result |= 1 << (7 - i)
		}
	}
	return result
}

func SendVNCKey(conn net.Conn, down bool, key uint32) error {
	buf := make([]byte, 8)
	buf[0] = 4 // KeyEvent
	if down {
		buf[1] = 1
	}
	binary.BigEndian.PutUint32(buf[4:], key)
	_, err := conn.Write(buf)
	return err
}
