package orders

import (
	"elevator/internal/types"
)

// findBestElevator returns the ID of the elevator with the lowest cost
// for the requested floor. An elevator is excluded from consideration if it
// is already assigned to the opposite hall-call direction at the same floor,
// ensuring that both calls can be served by different elevators.
func findBestElevator(
	floor int,
	hallDirectionIndex int,
	selfState types.ElevatorState,
	peerStates map[int]types.ElevatorState,
	hallCalls [][]types.HallCall,
) int {

	bestID := -1
	bestCost := 0

	if selfState.Available && !isAssignedToOppositeDirection(selfState.ID, floor, hallDirectionIndex, hallCalls) {
		bestID = selfState.ID
		bestCost = elevatorCost(selfState, floor, hallDirectionIndex)
	}

	for id, peer := range peerStates {
		if !peer.Available {
			continue
		}

		if isAssignedToOppositeDirection(id, floor, hallDirectionIndex, hallCalls) {
			continue
		}

		cost := elevatorCost(peer, floor, hallDirectionIndex)

		if bestID == -1 || cost < bestCost || (cost == bestCost && id < bestID) {
			bestCost = cost
			bestID = id
		}
	}

	// Fallback: if all candidates were excluded (e.g. only one elevator exists),
	// retry without the opposite-direction constraint.
	if bestID == -1 {
		if selfState.Available {
			bestID = selfState.ID
			bestCost = elevatorCost(selfState, floor, hallDirectionIndex)
		}

		for id, peer := range peerStates {
			if !peer.Available {
				continue
			}
			cost := elevatorCost(peer, floor, hallDirectionIndex)
			if bestID == -1 || cost < bestCost || (cost == bestCost && id < bestID) {
				bestCost = cost
				bestID = id
			}
		}
	}

	if bestID == -1 {
		return selfState.ID
	}

	return bestID
}

// isAssignedToOppositeDirection returns true if the given elevator ID is
// already assigned to the opposite hall-call direction at the same floor.
func isAssignedToOppositeDirection(
	elevatorID int,
	floor int,
	hallDirectionIndex int,
	hallCalls [][]types.HallCall,
) bool {
	if floor < 0 || floor >= len(hallCalls) {
		return false
	}

	oppositeIndex := types.HallUpIndex
	if hallDirectionIndex == types.HallUpIndex {
		oppositeIndex = types.HallDownIndex
	}

	oppositeCall := hallCalls[floor][oppositeIndex]
	active := oppositeCall.State == types.HallCallUnconfirmedAdd ||
		oppositeCall.State == types.HallCallConfirmedAdd

	return active && oppositeCall.AssignedTo == elevatorID
}

// elevatorCost estimates the assignment cost for a hall call
// based on floor distance and current movement direction.
func elevatorCost(
	e types.ElevatorState,
	requestFloor int,
	hallDirectionIndex int,
) int {

	if e.Floor == -1 {
		return 1000
	}

	distance := abs(e.Floor - requestFloor)
	cost := distance

	requestDirection := types.DirectionUp
	if hallDirectionIndex == types.HallDownIndex {
		requestDirection = types.DirectionDown
	}

	if e.Dir == types.DirectionStop {
		return cost
	}

	// moving toward request
	if (e.Dir == types.DirectionUp && requestFloor >= e.Floor) ||
		(e.Dir == types.DirectionDown && requestFloor <= e.Floor) {

		// correct direction
		if e.Dir == requestDirection {
			return cost
		}

		// wrong direction but passing
		return cost + 2
	}

	// moving away
	return cost + 4
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
