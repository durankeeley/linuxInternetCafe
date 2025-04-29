package utils

import (
	"net"
	"time"
)

func Ping(ip string) string {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", ip+":8080", timeout)
	if err != nil {
		return "offline"
	}
	conn.Close()
	return "online"
}
