package orders

import (
	"elevator/internal/types"
)

// MergePeerState stores a peer elevator state if it is newer than the current one.
func (om *OrderManager) MergePeerState(peerState types.ElevatorState) {
	if peerState.ID == om.selfState.ID {
		return
	}

	currentState, peerExists := om.peerStates[peerState.ID]
	if peerExists && peerState.SequenceNumber <= currentState.SequenceNumber {
		return
	}

	becameUnavailable := peerExists &&
		currentState.Available &&
		!peerState.Available

	om.peerStates[peerState.ID] = peerState

	if becameUnavailable {
		om.reassignHallCallsFromUnavailableElevator(peerState.ID)
	}
}

// MergeHallCalls merges hall-call state received from another node.
func (om *OrderManager) MergeHallCalls(peerID int, peerHallCalls [][]types.HallCall) {
	if peerID == om.selfState.ID {
		return
	}

	if len(peerHallCalls) != len(om.hallCalls) {
		return
	}

	for floor := 0; floor < len(om.hallCalls); floor++ {
		if len(peerHallCalls[floor]) != len(om.hallCalls[floor]) {
			return
		}
	}

	om.peerHallCalls[peerID] = cloneHallCalls(peerHallCalls)
	om.recomputeHallCalls()
}

// MergeNodeState merges a complete node state received from another elevator.
func (om *OrderManager) MergeNodeState(nodeState types.NodeState) {
	peerID := nodeState.ElevatorState.ID

	om.MergePeerState(nodeState.ElevatorState)
	om.MergeHallCalls(peerID, nodeState.HallCalls)
}

// HandleNodeStateUpdate merges a full node state received from the network.
func (om *OrderManager) HandleNodeStateUpdate(nodeState types.NodeState) bool {
	peerID := nodeState.ElevatorState.ID
	changed := false

	currentState, peerExists := om.peerStates[peerID]
	if !peerExists || nodeState.ElevatorState.SequenceNumber > currentState.SequenceNumber {
		om.MergePeerState(nodeState.ElevatorState)
		changed = true
	}

	before := om.CopyHallCalls()
	om.MergeHallCalls(peerID, nodeState.HallCalls)

	if !hallCallsEqual(before, om.hallCalls) {
		changed = true
	}

	return changed
}

// RemovePeer removes a lost peer from local state and reassigns any hall calls that were assigned to it.
func (om *OrderManager) RemovePeer(peerID int) {
	delete(om.peerStates, peerID)
	delete(om.peerHallCalls, peerID)

	om.reassignHallCallsFromUnavailableElevator(peerID)
	om.recomputeHallCalls()
}

// recomputeHallCalls rebuilds the merged hall-call view
// from this node's local contribution and the latest peer contributions.
func (om *OrderManager) recomputeHallCalls() {
	for floor := 0; floor < len(om.hallCalls); floor++ {
		for direction := 0; direction < len(om.hallCalls[floor]); direction++ {

			newestRevision := om.selfHallCalls[floor][direction].Revision
			newestState := om.selfHallCalls[floor][direction].State
			newestAssignedTo := om.selfHallCalls[floor][direction].AssignedTo
			winnerNodeID := om.selfState.ID

			for peerID, peerHallCalls := range om.peerHallCalls {
				if _, alive := om.peerStates[peerID]; !alive {
					continue
				}

				peer := peerHallCalls[floor][direction]

				if peer.Revision > newestRevision ||
					(peer.Revision == newestRevision && peerID < winnerNodeID) {
					newestRevision = peer.Revision
					newestState = peer.State
					newestAssignedTo = peer.AssignedTo
					winnerNodeID = peerID
				}
			}

			om.hallCalls[floor][direction].Revision = newestRevision
			om.hallCalls[floor][direction].AssignedTo = newestAssignedTo

			switch newestState {
			case types.HallCallUnconfirmedAdd, types.HallCallConfirmedAdd:
				om.hallCalls[floor][direction].State = types.HallCallConfirmedAdd
			default:
				om.hallCalls[floor][direction].State = types.HallCallAbsent
				om.hallCalls[floor][direction].AssignedTo = -1
			}
		}
	}
}

// hallCallsEqual checks whether two hall-call tables are identical.
func hallCallsEqual(a, b [][]types.HallCall) bool {
	if len(a) != len(b) {
		return false
	}
	for f := 0; f < len(a); f++ {
		if len(a[f]) != len(b[f]) {
			return false
		}
		for d := 0; d < len(a[f]); d++ {
			if a[f][d] != b[f][d] {
				return false
			}
		}
	}
	return true
}

// cloneHallCalls creates a deep copy of a hall-call table.
func cloneHallCalls(hallCalls [][]types.HallCall) [][]types.HallCall {
	hallCallsCopy := make([][]types.HallCall, len(hallCalls))

	for floor := 0; floor < len(hallCalls); floor++ {
		hallCallsCopy[floor] = make([]types.HallCall, len(hallCalls[floor]))
		copy(hallCallsCopy[floor], hallCalls[floor])
	}

	return hallCallsCopy
}
