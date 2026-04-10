package fsm
 
import (
	"heis/config"
	"heis/driver/elevio"
)
 
func ChooseDirection(e Elevator) DirnBehaviourPair {
	f := e.floor
	if f < 0 || f >= config.N_FLOORS {
		return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
	}
	here := hasRequestsAt(e, f)
	above := requestsAbove(e, f)
	below := requestsBelow(e, f)
 
	switch e.dirn {
	case D_Up:
		switch {
		case above:
			return DirnBehaviourPair{dirn: D_Up, state: ES_Moving}
		case here:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_DoorOpen}
		case below:
			return DirnBehaviourPair{dirn: D_Down, state: ES_Moving}
		default:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
		}
	case D_Down:
		switch {
		case below:
			return DirnBehaviourPair{dirn: D_Down, state: ES_Moving}
		case here:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_DoorOpen}
		case above:
			return DirnBehaviourPair{dirn: D_Up, state: ES_Moving}
		default:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
		}
	default:
		switch {
		case here:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_DoorOpen}
		case above:
			return DirnBehaviourPair{dirn: D_Up, state: ES_Moving}
		case below:
			return DirnBehaviourPair{dirn: D_Down, state: ES_Moving}
		default:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
		}
	}
}
 
func ShouldStop(e Elevator) bool {
	f := e.floor
	if f < 0 || f >= config.N_FLOORS {
		return false
	}
	switch e.dirn {
	case D_Down:
		if e.requests[f][B_Cab] || e.requests[f][B_HallDown] {
			return true
		}
		return e.requests[f][B_HallUp] && !requestsBelow(e, f)
	case D_Up:
		if e.requests[f][B_Cab] || e.requests[f][B_HallUp] {
			return true
		}
		return e.requests[f][B_HallDown] && !requestsAbove(e, f)
	case D_Stop:
		return hasRequestsAt(e, f)
	default:
		return false
	}
}
 
func collectClearedEvents(e *Elevator, floor int) []elevio.ButtonEvent {
	if floor < 0 || floor >= config.N_FLOORS {
		return nil
	}
	var events []elevio.ButtonEvent
 
	if e.requests[floor][B_Cab] {
		events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab})
		e.requests[floor][B_Cab] = false
	}
 
	switch e.dirn {
	case D_Up:
		if e.requests[floor][B_HallUp] {
			events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp})
			e.requests[floor][B_HallUp] = false
		}
		if !requestsAbove(*e, floor) && e.requests[floor][B_HallDown] {
			events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown})
			e.requests[floor][B_HallDown] = false
		}
	case D_Down:
		if e.requests[floor][B_HallDown] {
			events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown})
			e.requests[floor][B_HallDown] = false
		}
		if !requestsBelow(*e, floor) && e.requests[floor][B_HallUp] {
			events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp})
			e.requests[floor][B_HallUp] = false
		}
	case D_Stop:
		for b := Button(0); b < config.N_BUTTONS; b++ {
			if e.requests[floor][b] {
				events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(b)})
				e.requests[floor][b] = false
			}
		}
	}
	return events
}
 
func flushPendingFinReqs(pending []elevio.ButtonEvent, finReqCh chan<- elevio.ButtonEvent) {
	for _, ev := range pending {
		finReqCh <- ev
	}
}
 
func hasRequestsAt(e Elevator, floor int) bool {
	for b := 0; b < config.N_BUTTONS; b++ {
		if e.requests[floor][b] {
			return true
		}
	}
	return false
}
 
func requestsAbove(e Elevator, floor int) bool {
	for f := floor + 1; f < config.N_FLOORS; f++ {
		if hasRequestsAt(e, f) {
			return true
		}
	}
	return false
}
 
func requestsBelow(e Elevator, floor int) bool {
	for f := 0; f < floor; f++ {
		if hasRequestsAt(e, f) {
			return true
		}
	}
	return false
}