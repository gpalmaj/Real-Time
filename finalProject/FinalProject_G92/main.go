package main

import (
	//"FinalProject_G92/hardware"
	"FinalProject_G92/hardware"
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/network"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var localIP string

func LocalIP() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func main() {
	ipStr, err := LocalIP()
	if err != nil {
		fmt.Println("Error finding IP: ", err)
	}

	ip := net.ParseIP(ipStr)
	fmt.Println(ip)

	//ID is used for testing with many instances on the same machine.
	var id int
	if len(os.Args) > 1 {
		id, _ = strconv.Atoi(os.Args[1])
	} else {
		id = 0
	}

	//handles incomming worldviews
	heartbeatCh := make(chan network.Heartbeat)

	//handles outgoing worldviews
	worldviewCh := make(chan network.Worldview)

	//handles new orders
	orderCh := make(chan network.Order)
	rmOrderCh := make(chan network.Order)

	go network.Listener(heartbeatCh)
	go network.Heart(worldviewCh, ip, id)
	go network.OrdersFromKB(orderCh, rmOrderCh)

	stateCh := make(chan hardware.ElevatorState)

	elevio.Init(ipStr+":15657", network.N)

	go hardware.HardwareManager(stateCh, orderCh, rmOrderCh)

	network.NetworkManager(id, worldviewCh, heartbeatCh, orderCh, rmOrderCh)

}
