package ip

import (
	"net"
	"strings"
)

func FindIP() (string, error) {
	var localIP string

	conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localIP = strings.Split(conn.LocalAddr().String(), ":")[0]

	return localIP, nil
}
