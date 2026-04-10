package orders

import (
	"elevator/internal/types"
)

// reassignHallCallsFromUnavailableElevator reassigns active hall calls
// that were previously assigned to an unavailable elevator.
func (om *OrderManager) reassignHallCallsFromUnavailableElevator(elevatorID int) {
	changed := false

	for floor := 0; floor < len(om.hallCalls); floor++ {
		for direction := 0; direction < len(om.hallCalls[floor]); direction++ {
			call := om.hallCalls[floor][direction]

			active := isActiveHallCall(call.State)

			if !active || call.AssignedTo != elevatorID {
				continue
			}

			newAssignedTo := findBestElevator(
				floor,
				direction,
				om.selfState,
				om.peerStates,
				om.hallCalls,
			)

			om.selfHallCalls[floor][direction].State = types.HallCallUnconfirmedAdd
			om.selfHallCalls[floor][direction].Revision = call.Revision + 1
			om.selfHallCalls[floor][direction].AssignedTo = newAssignedTo
			changed = true
		}
	}

	if changed {
		om.recomputeHallCalls()
	}
}
