package distributor

import (
	"heis/order"
	"heis/worldview"
)

func mergeHallOrders(id string, local [][2]order.Order, received [][2]order.Order, peersAlive []string) [][2]order.Order {
	for f := range local {
		for b := range local[f] {
			local[f][b] = order.MergeOrder(id, local[f][b], received[f][b], peersAlive)
		}
	}
	return local
}

func mergeStates(
	localID string,
	localStates map[string]worldview.ElevatorState,
	senderID string,
	receivedStates map[string]worldview.ElevatorState,
	peersAlive []string,
) map[string]worldview.ElevatorState {

	// Update the physical state of the elevator that sent this message
	if recvState, ok := receivedStates[senderID]; ok {
		if existing, exists := localStates[senderID]; exists {
			existing.Floor = recvState.Floor
			existing.Direction = recvState.Direction
			existing.Behaviour = recvState.Behaviour
			localStates[senderID] = existing
		} else {
			localStates[senderID] = recvState
		}
	}

	for recvStateID, recvState := range receivedStates {
		if recvStateID == localID {
			continue
		}
		localState, exists := localStates[recvStateID]
		if !exists {
			localStates[recvStateID] = recvState
			continue
		}
		for f := range recvState.CabOrders {
			if recvStateID == senderID{
				localState.CabOrders[f] = order.MergeOrder(
					localID,
					localState.CabOrders[f],
					recvState.CabOrders[f],
					peersAlive,
				)
			}
		}
		localStates[recvStateID] = localState
	}

	return localStates
}

func copyWorldView(worldView worldview.WorldView) worldview.WorldView {
	copiedWorldView := worldview.WorldView{
		HallOrders: make([][2]order.Order, len(worldView.HallOrders)),
		States:     make(map[string]worldview.ElevatorState),
	}
	for f := range worldView.HallOrders {
		for b := range worldView.HallOrders[f] {
			copiedWorldView.HallOrders[f][b] = order.Order{
				Status:  worldView.HallOrders[f][b].Status,
				Barrier: order.CopyBarrier(worldView.HallOrders[f][b].Barrier),
			}
		}
	}
	for id, state := range worldView.States {
		newCabOrders := make([]order.Order, len(state.CabOrders))
		for f, o := range state.CabOrders {
			newCabOrders[f] = order.Order{
				Status:  o.Status,
				Barrier: order.CopyBarrier(o.Barrier),
			}
		}
		copiedState := worldview.ElevatorState{
			Floor:     state.Floor,
			Direction: state.Direction,
			Behaviour: state.Behaviour,
			CabOrders: newCabOrders,
		}
		copiedWorldView.States[id] = copiedState
	}
	return copiedWorldView
}

func copyHallOrders(orders [][2]order.Order) [][2]order.Order {
	copied := make([][2]order.Order, len(orders))
	for f := range orders {
		for b := range orders[f] {
			copied[f][b] = order.Order{
				Status:  orders[f][b].Status,
				Barrier: order.CopyBarrier(orders[f][b].Barrier),
			}
		}
	}
	return copied
}

func hallOrdersStatusChanged(before, after [][2]order.Order) bool {
    for f := range before {
        for b := range before[f] {
            if before[f][b].Status != after[f][b].Status {
                return true
            }
        }
    }
    return false
}