package network

import (
	"FinalProject_G92/config"
)

func Cost(
	floor, direction, targetFloor, targetDir int,
	cabCalls [config.N]bool,
) int {

	//						SETUP

	//	construct a local map and include cabcalls
	var localOrders [config.N][3]bool
	for f := range config.N {
		localOrders[f][2] = cabCalls[f]
	}

	// injecting target call
	localOrders[targetFloor][targetDir] = true

	//	return value
	cost := 0

	//	values for simulation
	simFloor := floor
	simDir := direction

	for {
		// if the elevator is at the floor and sould stop, cost is zero
		if simFloor == targetFloor && shouldStop(simFloor, simDir, localOrders) {
			return cost
		}
		// checks if should stop at the floor
		if shouldStop(simFloor, simDir, localOrders) {
			//adds door cost
			cost += int(config.DoorOpenDuration)
			clearAtFloor(&localOrders, simFloor)
			simDir = 0
		}

		simDir = chooseDirection(simFloor, simDir, localOrders)

		//Simulation arrived
		if simDir == 0 {
			return cost
		}

		//Simulation travels up or down
		if simDir == 1 {
			simFloor++
		} else {
			simFloor--
		}

		//adds time between floors cost
		cost += int(config.BetweenFloorsDuration)
		//INFO time is in nanoseconds, so very high number there. Investigate if issues.
	}

}

// stops if there is a matching call there
func shouldStop(floor, dir int, orders [config.N][3]bool) bool {
	switch dir {
	case 1:
		return orders[floor][0] || orders[floor][2] || !ordersAbove(floor, orders)
	case -1:
		return orders[floor][1] || orders[floor][2] || !ordersBelow(floor, orders)
	case 0:
		return true
	}
	return true
}

// see where it has to go next
func chooseDirection(floor, dir int, orders [config.N][3]bool) int {
	switch {
	case ordersAbove(floor, orders) && (dir == 1 || dir == 0):
		return 1
	case ordersBelow(floor, orders) && (dir == -1 || dir == 0):
		return -1
	case ordersAbove(floor, orders):
		return 1
	case ordersBelow(floor, orders):
		return -1
	default:
		return 0
	}
}

func clearAtFloor(orders *[config.N][3]bool, floor int) {
	for btn := range 3 {
		orders[floor][btn] = false
	}
}

func ordersAbove(floor int, orders [config.N][3]bool) bool {
	for f := floor + 1; f < config.N; f++ {
		for btn := range 3 {
			if orders[f][btn] {
				return true
			}
		}
	}

	return false
}

func ordersBelow(floor int, orders [config.N][3]bool) bool {
	for f := range floor {
		for btn := range 3 {
			if orders[f][btn] {
				return true
			}
		}
	}

	return false
}
