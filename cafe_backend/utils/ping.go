package utils

import (
	"net"
	"strconv"
	"time"
)

func Ping(ip string, port int) string {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", ip+":"+strconv.Itoa(port), timeout)
	if err != nil {
		return "offline"
	}
	conn.Close()
	return "online"
}
