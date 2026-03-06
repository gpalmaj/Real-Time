package network

import (
	"net"
	"strings"
)

var localIPCache string

func LocalIP() (string, error) {
	if localIPCache == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIPCache = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIPCache, nil
}
