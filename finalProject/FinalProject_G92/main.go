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

	elevio.Init(ipStr+":15657", config.N)

	var id int
	if len(os.Args) > 1 {
		id, _ = strconv.Atoi(os.Args[1])
	}

	// channels
	heartbeatCh := make(chan models.Heartbeat)
	worldviewCh := make(chan models.Worldview)
	orderCh := make(chan models.Order)
	rmOrderCh := make(chan models.Order)
	lightsCh := make(chan [config.N]models.HallCall)
	statusCh := make(chan models.StatusMessage)

	// launch goroutines
	go network.HeartbeatListener(heartbeatCh)
	go network.HeartbeatSender(worldviewCh, ip, id)
	go debug.OrdersFromKB(orderCh, rmOrderCh)
	go hardware.HallLights(lightsCh)
	go hardware.HardwareManager(orderCh, rmOrderCh, statusCh)

	network.NetworkManager(id, worldviewCh, heartbeatCh, orderCh, rmOrderCh, lightsCh, statusCh)
}
