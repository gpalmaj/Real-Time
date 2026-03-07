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
	//DoorOpenDuration is how much time the door stays open
	DoorOpenDuration = 3 * time.Second
	//BetweenFloorsDuration is how much time the elevator takes to go up or down a floor
	BetweenFloorsDuration = 2 * time.Second
)
