package main

import (
	"FinalProject_G92/network"
	"net"
	"os"
	"strconv"
)

func main() {
	ip := net.ParseIP("127.0.0.1")

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

	network.NetworkManager(id, worldviewCh, heartbeatCh, orderCh, rmOrderCh)

}
