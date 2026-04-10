package orders

import (
	"elevator/internal/types"
)

// isActiveHallCall reports whether a hall call is currently active.
func isActiveHallCall(state types.HallCallState) bool {
	return state == types.HallCallUnconfirmedAdd || state == types.HallCallConfirmedAdd
}

// isOwnedByself returns true if the hall call at the given floor/direction
// is active and assigned to this elevator.
func (om *OrderManager) isOwnedBySelf(floor int, dirIndex int) bool {
	call := om.hallCalls[floor][dirIndex]
	return isActiveHallCall(call.State) && call.AssignedTo == om.selfState.ID
}

// CreateClearRequestAction describes which requests should be cleared
// at the current floor for the current direction of travel.
// Only hall calls that are assigned to this elevator are cleared.
func (om *OrderManager) CreateClearRequestAction() types.ClearRequestAction {
	currentFloor := om.selfState.Floor
	currentDirection := om.selfState.Dir

	if currentFloor < 0 || currentFloor >= len(om.cabCalls) {
		return types.ClearRequestAction{}
	}

	clearAction := types.ClearRequestAction{
		ClearCab: om.cabCalls[currentFloor],
	}

	upOwned := om.isOwnedBySelf(currentFloor, types.HallUpIndex)
	downOwned := om.isOwnedBySelf(currentFloor, types.HallDownIndex)

	// When the door is open, clear all owned calls at this floor
	if om.selfState.Behaviour == types.BehaviourDoorOpen {
		clearAction.ClearHallUp = upOwned
		clearAction.ClearHallDown = downOwned
		return clearAction
	}

	switch currentDirection {
	case types.DirectionUp:
		if upOwned {
			clearAction.ClearHallUp = true
		} else if !hasRequestsAbove(currentFloor, om.cabCalls, om.hallCalls, om.selfState) && downOwned {
			clearAction.ClearHallDown = true
		}

	case types.DirectionDown:
		if downOwned {
			clearAction.ClearHallDown = true
		} else if !hasRequestsBelow(currentFloor, om.cabCalls, om.hallCalls, om.selfState) && upOwned {
			clearAction.ClearHallUp = true
		}

	case types.DirectionStop:
		if upOwned && downOwned {
			if hasRequestsAbove(currentFloor, om.cabCalls, om.hallCalls, om.selfState) {
				clearAction.ClearHallUp = true
			} else if hasRequestsBelow(currentFloor, om.cabCalls, om.hallCalls, om.selfState) {
				clearAction.ClearHallDown = true
			} else {
				clearAction.ClearHallUp = true
				clearAction.ClearHallDown = true
			}
		} else {
			if upOwned {
				clearAction.ClearHallUp = true
			}
			if downOwned {
				clearAction.ClearHallDown = true
			}
		}
	}

	return clearAction
}

// ApplyClearedRequests updates the request tables after the elevator has served a floor.
func (om *OrderManager) ApplyClearedRequests(clear types.ClearRequestAction) {
	currentFloor := om.selfState.Floor

	if currentFloor < 0 || currentFloor >= len(om.cabCalls) {
		return
	}

	if clear.ClearCab {
		om.cabCalls[currentFloor] = false
		_ = om.saveCabCalls()
	}

	changed := false
	if clear.ClearHallUp {
		om.selfHallCalls[currentFloor][types.HallUpIndex].State = types.HallCallUnconfirmedRemove
		om.selfHallCalls[currentFloor][types.HallUpIndex].Revision =
			om.hallCalls[currentFloor][types.HallUpIndex].Revision + 1
		om.selfHallCalls[currentFloor][types.HallUpIndex].AssignedTo = -1
		changed = true
	}

	if clear.ClearHallDown {
		om.selfHallCalls[currentFloor][types.HallDownIndex].State = types.HallCallUnconfirmedRemove
		om.selfHallCalls[currentFloor][types.HallDownIndex].Revision =
			om.hallCalls[currentFloor][types.HallDownIndex].Revision + 1
		om.selfHallCalls[currentFloor][types.HallDownIndex].AssignedTo = -1
		changed = true
	}

	if changed {
		om.recomputeHallCalls()
	}
}
