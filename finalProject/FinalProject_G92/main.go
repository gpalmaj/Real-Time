package main

import (
	"FinalProject_G92/network"
	"net"
)

func main() {
	ip := net.ParseIP("127.0.0.1")

	//handles incomming worldviews
	heartbeatCh := make(chan network.Heartbeat)

	//handles outgoing worldviews
	worldviewCh := make(chan [network.N]network.Call)

	//handles new orders
	orderCh := make(chan network.Order)
	rmOrderCh := make(chan network.Order)

	go network.Listener(heartbeatCh, ip)
	go network.Heart(worldviewCh, ip)
	go network.OrdersFromKB(orderCh, rmOrderCh)

	network.WorldviewManager(worldviewCh, heartbeatCh, orderCh, rmOrderCh)

}
