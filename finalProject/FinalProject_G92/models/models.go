package models

import (
	"FinalProject_G92/config"
	"net"
	"time"
)

type HallCall struct {
	Up      bool
	Down    bool
	UpSeq   int
	DownSeq int
}

type Order struct {
	Cab   bool
	Dir   bool
	Floor int
}

type StatusMessage struct {
	Floor       int
	Direction   int
	Operational bool
}

type Worldview struct {
	HallCalls  [config.N]HallCall
	CabCalls   [config.N]bool
	CabCallLog map[int][config.N]bool
	Status     StatusMessage
}

type Heartbeat struct {
	ID        int
	IP        net.IP
	Worldview Worldview
}

type Node struct {
	Alive     bool
	Lastseen  time.Time
	Worldview Worldview
}
