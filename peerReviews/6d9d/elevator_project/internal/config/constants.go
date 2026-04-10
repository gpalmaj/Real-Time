package config

import (
	"time"
)

// Building configuration
const (
	NumberOfFloors         = 4
	NumberOfHallDirections = 2
)

// Network configuration
const (
	PeerTimeout = 3 * time.Second
)

// Elevator configuration
const (
	DoorOpenDuration = 3 * time.Second
)

// Watchdog configuration
const (
	MovementWatchdogTimeoutDuration = 5 * time.Second
)
