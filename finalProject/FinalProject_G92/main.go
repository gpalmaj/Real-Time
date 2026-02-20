package main

import (
	//"FinalProject_G92/hardware"
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

	//ID is usied for testing with many instences on the same machine.
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

	elevio.Init(ipStr+":15657", network.N)

	//lobby := make(map[int]network.Node)

	btnCh := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(btnCh)

	flrCh := make(chan int)
	go elevio.PollFloorSensor(flrCh)

	obstrCh := make(chan bool)
	go elevio.PollObstructionSwitch(obstrCh)

	stopCh := make(chan bool)
	go elevio.PollStopButton(stopCh)

	//go hardware.HallLights(lobby)

	//hardware.ExecuteOrder(btnCh)HallLights

	network.NetworkManager(id, worldviewCh, heartbeatCh, orderCh, rmOrderCh)

}
