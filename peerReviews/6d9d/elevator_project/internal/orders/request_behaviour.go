package orders

import (
	"elevator/internal/types"
)

// chooseDirection decides which direction this elevator should move next
// based on its current floor, direction, and pending local requests.
func chooseDirection(
	currentFloor int,
	currentDirection types.Direction,
	cabCalls []bool,
	hallCalls [][]types.HallCall,
	selfState types.ElevatorState,
) types.Direction {

	if currentDirection == types.DirectionUp {
		if hasRequestsAbove(currentFloor, cabCalls, hallCalls, selfState) {
			return types.DirectionUp
		}
		if hasRequestsBelow(currentFloor, cabCalls, hallCalls, selfState) {
			return types.DirectionDown
		}
		return types.DirectionStop
	}

	if currentDirection == types.DirectionDown {
		if hasRequestsBelow(currentFloor, cabCalls, hallCalls, selfState) {
			return types.DirectionDown
		}
		if hasRequestsAbove(currentFloor, cabCalls, hallCalls, selfState) {
			return types.DirectionUp
		}
		return types.DirectionStop
	}

	if hasRequestsAbove(currentFloor, cabCalls, hallCalls, selfState) {
		return types.DirectionUp
	}

	if hasRequestsBelow(currentFloor, cabCalls, hallCalls, selfState) {
		return types.DirectionDown
	}

	return types.DirectionStop
}

// shouldServeCurrentFloor decides whether the elevator should serve at the current floor.
func shouldServeCurrentFloor(
	currentFloor int,
	currentDirection types.Direction,
	cabCalls []bool,
	hallCalls [][]types.HallCall,
	selfState types.ElevatorState,
) bool {
	if cabCalls[currentFloor] {
		return true
	}

	hasUpHere := isHallCallForThisElevator(
		currentFloor,
		types.HallUpIndex,
		hallCalls,
		selfState,
	)

	hasDownHere := isHallCallForThisElevator(
		currentFloor,
		types.HallDownIndex,
		hallCalls,
		selfState,
	)

	// If the elevator is idle or door-open at this floor,
	// serve any owned hall call here immediately.
	if selfState.Behaviour == types.BehaviourIdle || selfState.Behaviour == types.BehaviourDoorOpen {
		return hasUpHere || hasDownHere
	}

	switch currentDirection {
	case types.DirectionUp:
		if hasUpHere {
			return true
		}
		return !hasRequestsAbove(currentFloor, cabCalls, hallCalls, selfState) && hasDownHere

	case types.DirectionDown:
		if hasDownHere {
			return true
		}
		return !hasRequestsBelow(currentFloor, cabCalls, hallCalls, selfState) && hasUpHere

	case types.DirectionStop:
		return hasUpHere || hasDownHere

	default:
		return false
	}
}

// hasRequestsAbove returns true if there are any pending requests
// above the current floor for this elevator.
func hasRequestsAbove(
	currentFloor int,
	cabCalls []bool,
	hallCalls [][]types.HallCall,
	selfState types.ElevatorState,
) bool {
	for floor := currentFloor + 1; floor < len(cabCalls); floor++ {
		if cabCalls[floor] {
			return true
		}

		if isHallCallForThisElevator(
			floor,
			types.HallUpIndex,
			hallCalls,
			selfState,
		) {
			return true
		}

		if isHallCallForThisElevator(
			floor,
			types.HallDownIndex,
			hallCalls,
			selfState,
		) {
			return true
		}
	}

	return false
}

// hasRequestsBelow returns true if there are any pending requests
// below the current floor for this elevator.
func hasRequestsBelow(
	currentFloor int,
	cabCalls []bool,
	hallCalls [][]types.HallCall,
	selfState types.ElevatorState,
) bool {
	for floor := 0; floor < currentFloor; floor++ {
		if cabCalls[floor] {
			return true
		}

		if isHallCallForThisElevator(
			floor,
			types.HallUpIndex,
			hallCalls,
			selfState,
		) {
			return true
		}

		if isHallCallForThisElevator(
			floor,
			types.HallDownIndex,
			hallCalls,
			selfState,
		) {
			return true
		}
	}

	return false
}

// isHallCallForThisElevator returns true when an active hall call
// is assigned to this elevator.
func isHallCallForThisElevator(
	floor int,
	hallDirectionIndex int,
	hallCalls [][]types.HallCall,
	selfState types.ElevatorState,
) bool {
	state := hallCalls[floor][hallDirectionIndex].State

	switch state {
	case types.HallCallUnconfirmedAdd, types.HallCallConfirmedAdd:
		return hallCalls[floor][hallDirectionIndex].AssignedTo == selfState.ID
	default:
		return false
	}
}
