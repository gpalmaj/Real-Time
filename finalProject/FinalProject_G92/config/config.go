package config

import "time"

const (
	// N is the number of floors
	N = 4
	// Port is the UDP broadcast port
	Port = 3000
	// DisconnectTimeout is how long before a node is considered disconnected
	DisconnectTimeout = 3 * time.Second
	// HeartbeatInterval is how often heartbeats are sent
	HeartbeatInterval = 1 * time.Second
	//DoorOpenDuratino is how much time the door stays open
	DoorOpenDuration = 3 * time.Second
)
