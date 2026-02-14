package main

import (
	"FinalProject_G92/network"
	"net"
)

func main() {
	ip := net.ParseIP("127.0.0.1")

	worldviewCh := make(chan [network.N]network.Call)
	heartbeatCh := make(chan network.Heartbeat)

	go network.Heart(worldviewCh, ip)
	go network.Listener(heartbeatCh, ip)

	// Send an initial worldview
	var wv [network.N]network.Call
	wv[0] = network.Call{Up: true}
	worldviewCh <- wv

	// Print heartbeats as they arrive
	for hb := range heartbeatCh {
		network.PrintWorldView(hb.Worldview)
	}
}
