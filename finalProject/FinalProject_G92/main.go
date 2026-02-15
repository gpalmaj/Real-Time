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

	go network.Listener(heartbeatCh, ip)
	go network.Heart(worldviewCh, ip)

	//worldview
	go network.WorldviewManager(worldviewCh, heartbeatCh)

	// Print heartbeats as they arrive
	for hb := range heartbeatCh {
		network.PrintWorldView(hb.Worldview)
	}
}
