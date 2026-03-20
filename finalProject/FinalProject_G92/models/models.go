package models

import (
	"FinalProject_G92/config"
	"net"
	"time"
)

// HallCall represents the up/down call state for a single floor.
// Sequence numbers enable conflict-free merging: highest sequence wins.
type HallCall struct {
	Up      bool
	Down    bool
	UpSeq   int
	DownSeq int
}

// Order represents a request to add or remove a call.
// If Cab is true, it is a cab call. Otherwise Dir indicates up (true) or down (false).
type Order struct {
	Cab   bool
	Dir   bool
	Floor int
}

// StatusMessage is a snapshot of an elevator's current state, sent from controller to coordinator.
type StatusMessage struct {
	Floor       int
	Direction   int
	Operational bool
}

// Worldview is a node's complete view of the system: shared hall calls,
// local cab calls, a backup log of all nodes' cab calls for crash recovery, and elevator status.
type Worldview struct {
	HallCalls  [config.N]HallCall
	CabCalls   [config.N]bool
	CabCallLog map[int][config.N]bool
	Status     StatusMessage
}

// Heartbeat is the packet broadcast over UDP, containing the sender's identity and full worldview.
type Heartbeat struct {
	ID        int
	IP        net.IP
	Worldview Worldview
}

// Node represents a peer elevator in the lobby, tracking its liveness and last known worldview.
type Node struct {
	Alive     bool
	Lastseen  time.Time
	Worldview Worldview
}
