package main

import (
	"FinalProject_G92/config"
	"FinalProject_G92/debug"
	"FinalProject_G92/hardware"
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/models"
	"FinalProject_G92/network"
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	ipStr, err := network.LocalIP()
	if err != nil {
		fmt.Println("Error finding IP: ", err)
	}

	ip := net.ParseIP(ipStr)
	fmt.Println(ip)

	id := int(ip.To4()[3]) // derive from ip as standartd
	port := config.ElevatorServerPort
	if len(os.Args) > 1 { // from command line to test on single machine
		id, _ = strconv.Atoi(os.Args[1])
		port += id
	}

	portStr := strconv.Itoa(port)

	elevio.Init(ipStr+":"+portStr, config.N)

	// channels
	heartbeatCh := make(chan models.Heartbeat)
	worldviewCh := make(chan models.Worldview)
	orderCh := make(chan models.Order)
	rmOrderCh := make(chan models.Order)
	lightsCh := make(chan [config.N]models.HallCall)
	statusCh := make(chan models.StatusMessage)
	assignCh := make(chan models.Order, 8)

	// launch goroutines
	go network.HeartbeatListener(heartbeatCh)
	go network.HeartbeatSender(worldviewCh, ip, id)
	go debug.OrdersFromKB(orderCh, rmOrderCh)
	go hardware.HallLights(lightsCh)
	go hardware.HardwareManager(assignCh, orderCh, rmOrderCh, statusCh)

	network.NetworkManager(id, worldviewCh, heartbeatCh, assignCh, orderCh, rmOrderCh, lightsCh, statusCh)
}
