
package fsm
 
import (
	"time"
	"heis/config"
	"heis/elevtype"
)
 
type Dirn = elevtype.Dirn
type Button = elevtype.Button
 
const (
	D_Down = elevtype.D_Down
	D_Stop = elevtype.D_Stop
	D_Up   = elevtype.D_Up
	B_HallUp   = elevtype.B_HallUp
	B_HallDown = elevtype.B_HallDown
	B_Cab      = elevtype.B_Cab
)
 
type ElevatorState int
const(
	ES_Idle ElevatorState = iota
	ES_DoorOpen
	ES_Moving
)
 
type ElevatorConfig struct{
	DoorOpenDuration time.Duration
}
 
type Elevator struct{
	floor		int
	dirn		Dirn
	requests 	[config.N_FLOORS][config.N_BUTTONS]bool
	state 		ElevatorState
	elevConfig	ElevatorConfig
}
 
type DirnBehaviourPair struct{
	dirn	Dirn
	state	ElevatorState
}
 
type ElevatorStateMsg struct {
	Floor     int
	Dirn      Dirn
	State     ElevatorState
}
 
func elevatorInit() Elevator{
	return Elevator{
		floor: config.FloorUnknown,
		dirn: D_Stop,
		requests: [config.N_FLOORS][config.N_BUTTONS]bool{},
		state: ES_Idle,
		elevConfig: ElevatorConfig{DoorOpenDuration: config.DoorOpenDuration},
	}
}