package orders

import "elevator/internal/types"

// ComputeNextAction decides what this elevator should do next
// based on local cab calls and the merged hall-call view.
func (om *OrderManager) ComputeNextAction() types.Action {
	currentFloor := om.selfState.Floor
	currentDirection := om.selfState.Dir

	if currentFloor < 0 {
		// Choose direction based on available requests
		if currentDirection == types.DirectionUp || currentDirection == types.DirectionDown {
			return types.Action{Direction: currentDirection}
		}

		// Restart between floors: choose direction based on available requests
		for floor := 0; floor < len(om.cabCalls); floor++ {
			if om.cabCalls[floor] {
				if floor >= len(om.cabCalls)/2 {
					return types.Action{Direction: types.DirectionUp}
				}
				return types.Action{Direction: types.DirectionDown}
			}
		}

		// Check for assigned hall requests above/below
		if hasRequestsAbove(0, om.cabCalls, om.hallCalls, om.selfState) {
			return types.Action{Direction: types.DirectionUp}
		}
		if hasRequestsBelow(len(om.cabCalls)-1, om.cabCalls, om.hallCalls, om.selfState) {
			return types.Action{Direction: types.DirectionDown}
		}

		return types.Action{Direction: types.DirectionDown}
	}

	if currentFloor >= len(om.cabCalls) {
		return types.Action{Direction: types.DirectionStop}
	}

	switch currentDirection {
	case types.DirectionUp, types.DirectionDown, types.DirectionStop:
	default:
		return types.Action{Direction: types.DirectionStop}
	}

	// Handling when the elevator is at the current floor with the door open
	if om.selfState.Behaviour == types.BehaviourDoorOpen {
		// Check if the current floor needs to be served
		if shouldServeCurrentFloor(currentFloor, currentDirection, om.cabCalls, om.hallCalls, om.selfState) {
			return types.Action{
				Direction: currentDirection, // Keep the current direction
				OpenDoor:  true,
			}
		}

		// Pick the next direction after serving this floor
		nextDirection := chooseDirection(
			currentFloor,
			currentDirection,
			om.cabCalls,
			om.hallCalls,
			om.selfState,
		)

		return types.Action{
			Direction: nextDirection,
			OpenDoor:  false,
		}
	}

	// When serving the current floor, the elevator should stop and open doors if needed
	if shouldServeCurrentFloor(
		currentFloor,
		currentDirection,
		om.cabCalls,
		om.hallCalls,
		om.selfState,
	) {
		return types.Action{
			Direction: types.DirectionStop,
			OpenDoor:  true,
		}
	}

	// Move to the next valid request direction
	nextDirection := chooseDirection(
		currentFloor,
		currentDirection,
		om.cabCalls,
		om.hallCalls,
		om.selfState,
	)

	return types.Action{
		Direction: nextDirection,
		OpenDoor:  false,
	}
}
