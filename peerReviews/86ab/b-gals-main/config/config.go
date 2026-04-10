package config

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Network
const (
	PortPeers       = 19001
	PortDistributor = 19002
)

// Hardware
const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

// Paths
const HRAPath = "costfunction/cost_fns/hall_request_assigner/"

// Timing
const (
	DoorOpenDuration             = 3 * time.Second
	MotorLossTimeout             = 3 * time.Second
	ObstrTripsBeforeFloorInvalid = 2
)

// Per-elevator settings read from .con file
type ElevatorConfig struct {
	Id     string
	PortIo int
}

func ReadConfigFile(path string) ElevatorConfig {
	conf := ElevatorConfig{
		Id:     "elev1",
		PortIo: 15657,
	}
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		if len(words) < 2 {
			log.Fatalf("Invalid config line: %s", line)
		}
		switch words[0] {
		case "--id":
			conf.Id = words[1]
		case "--portio":
			port, err := strconv.Atoi(words[1])
			if err != nil {
				log.Fatalf("Invalid port: %s", words[1])
			}
			conf.PortIo = port
		default:
			log.Fatalf("Unknown config key: %s", words[0])
		}
	}
	return conf
}

const FloorUnknown = -1