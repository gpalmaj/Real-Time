package distributor

import (
	"heis/config"
	"heis/costfunction"
	"heis/driver/elevio"
	"heis/elevtype"
	"heis/order"
	"heis/worldview"
)

func computeAndSendAssignment(id string, worldView worldview.WorldView, peersAlive []string, assignmentCh chan [config.N_FLOORS][config.N_BUTTONS]bool) {

	hallReqs := extractConfirmedHallRequests(worldView)
	states := buildHRAStates(worldView, peersAlive)

	if len(states) == 0 {
		return
	}

	assigned, ok := costfunction.Compute(id, hallReqs, states)
	if !ok {
		return
	}

	reqs := mergeAssignedWithCabOrders(id, assigned, worldView)
	select {
	case <-assignmentCh:
	default:
	}
	assignmentCh <- reqs
}

func extractConfirmedHallRequests(worldView worldview.WorldView) [][2]bool {
	hallReqs := make([][2]bool, config.N_FLOORS)
	for f := range worldView.HallOrders {
		hallReqs[f][int(elevtype.B_HallUp)] = worldView.HallOrders[f][int(elevtype.B_HallUp)].Status == order.OrderStatusConfirmed
		hallReqs[f][int(elevtype.B_HallDown)] = worldView.HallOrders[f][int(elevtype.B_HallDown)].Status == order.OrderStatusConfirmed
	}
	return hallReqs
}

func buildHRAStates(worldView worldview.WorldView, peersAlive []string) map[string]costfunction.HRAElevState {
	states := make(map[string]costfunction.HRAElevState)

	confirmedHallOrders := make([][2]bool, config.N_FLOORS)
	for f := range worldView.HallOrders {
		confirmedHallOrders[f][int(elevtype.B_HallUp)] = worldView.HallOrders[f][elevtype.B_HallUp].Status == order.OrderStatusConfirmed
		confirmedHallOrders[f][int(elevtype.B_HallDown)] = worldView.HallOrders[f][elevtype.B_HallDown].Status == order.OrderStatusConfirmed
	}

	for id, s := range worldView.States {
		if s.Floor < 0 {
			continue
		}
		if !order.ContainsID(peersAlive, id) {
			continue
		}

		cabReqs := make([]bool, config.N_FLOORS)
		for f, o := range s.CabOrders {
			cabReqs[f] = o.Status == order.OrderStatusConfirmed
		}

		states[id] = costfunction.HRAElevState{
			Behavior:    behaviourToString(s.Behaviour),
			Floor:       s.Floor,
			Direction:   dirnToString(s.Direction),
			CabRequests: cabReqs,
		}
	}
	return states
}

func behaviourToString(b elevtype.Behaviour) string {
	switch b {
	case elevtype.B_Moving:
		return "moving"
	case elevtype.B_DoorOpen:
		return "doorOpen"
	default:
		return "idle"
	}
}

func dirnToString(d elevtype.Dirn) string {
	switch d {
	case elevtype.D_Up:
		return "up"
	case elevtype.D_Down:
		return "down"
	default:
		return "stop"
	}
}

func mergeAssignedWithCabOrders(id string, assigned [][2]bool, worldView worldview.WorldView) [config.N_FLOORS][config.N_BUTTONS]bool {
	var reqs [config.N_FLOORS][config.N_BUTTONS]bool
	state, exists := worldView.States[id]
	for f := 0; f < config.N_FLOORS; f++ {
		reqs[f][int(elevio.BT_HallUp)] = assigned[f][int(elevtype.B_HallUp)]
		reqs[f][int(elevio.BT_HallDown)] = assigned[f][int(elevtype.B_HallDown)]
		if exists {
			reqs[f][int(elevio.BT_Cab)] = state.CabOrders[f].Status == order.OrderStatusConfirmed
		}
	}
	return reqs
}
