package controller

import (
	"FinalProject_G92/config"
	"FinalProject_G92/controller/elevio"
	"fmt"
)

type FSMState int

const (
	Idle     FSMState = 0
	Moving   FSMState = 1
	DoorOpen FSMState = 2
	Stopped  FSMState = 3
)

const OrderTypes = 3

// ElevatorFSM holds the elevator's state and pending orders.
type ElevatorFSM struct {
	State      FSMState
	Floor      int
	Direction  elevio.MotorDirection
	Orders     [config.N][OrderTypes]bool // [floor][buttonType]
	Obstructed bool
}

// OnButtonPress stores an order and, if idle, starts moving or opens doors.
func (fsm *ElevatorFSM) OnButtonPress(floor int, btn elevio.ButtonType) {
	fsm.Orders[floor][btn] = true
	elevio.SetButtonLamp(btn, floor, true)

	switch fsm.State {
	case Idle:
		if floor == fsm.Floor {
			fsm.State = DoorOpen
			elevio.SetDoorOpenLamp(true)
			fsm.clearOrdersAtFloor()
		} else {
			fsm.chooseDirectionAndMove()

		}
	case DoorOpen, Moving, Stopped:
		// Order is stored, will be served when appropriate
	}
}

// OnFloorArrival updates the floor and stops if the elevator should serve this floor.
func (fsm *ElevatorFSM) OnFloorArrival(floor int) {
	fsm.Floor = floor
	elevio.SetFloorIndicator(floor)

	if fsm.shouldStop() {
		elevio.SetMotorDirection(elevio.MD_Stop)
		fsm.State = DoorOpen
		elevio.SetDoorOpenLamp(true)
		fsm.clearOrdersAtFloor()
	}
}

func (fsm *ElevatorFSM) OnObstruction(obstructed bool) {
	fsm.Obstructed = obstructed
}

func (fsm *ElevatorFSM) OnStopButton() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	fsm.State = Stopped
	fmt.Println("Elevator stopped")
}

// shouldStop returns true if the elevator should stop at its current floor:
// matching directional order, cab call, or no more orders ahead.
func (fsm *ElevatorFSM) shouldStop() bool {
	switch fsm.Direction {
	case elevio.MD_Up:
		return fsm.Orders[fsm.Floor][elevio.BT_HallUp] ||
			fsm.Orders[fsm.Floor][elevio.BT_Cab] ||
			!fsm.ordersAbove()
	case elevio.MD_Down:
		return fsm.Orders[fsm.Floor][elevio.BT_HallDown] ||
			fsm.Orders[fsm.Floor][elevio.BT_Cab] ||
			!fsm.ordersBelow()
	case elevio.MD_Stop:
		return true
	}
	return true
}

// clearOrdersAtFloor clears the cab call and the hall call matching travel direction.
// If reversing (no orders ahead), clears the opposite direction instead.
// Returns true if a hall call was cleared.
func (fsm *ElevatorFSM) clearOrdersAtFloor() bool {
	fsm.Orders[fsm.Floor][elevio.BT_Cab] = false
	elevio.SetButtonLamp(elevio.BT_Cab, fsm.Floor, false)

	// Clear the order matching travel direction
	if fsm.Direction == elevio.MD_Up && fsm.Orders[fsm.Floor][elevio.BT_HallUp] {
		fsm.Orders[fsm.Floor][elevio.BT_HallUp] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, fsm.Floor, false)
		return true
	}
	if fsm.Direction == elevio.MD_Down && fsm.Orders[fsm.Floor][elevio.BT_HallDown] {
		fsm.Orders[fsm.Floor][elevio.BT_HallDown] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, fsm.Floor, false)
		return true
	}

	// about to reverse
	if !fsm.ordersAbove() && fsm.Orders[fsm.Floor][elevio.BT_HallDown] {
		fsm.Orders[fsm.Floor][elevio.BT_HallDown] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, fsm.Floor, false)
		return true
	}
	if !fsm.ordersBelow() && fsm.Orders[fsm.Floor][elevio.BT_HallUp] {
		fsm.Orders[fsm.Floor][elevio.BT_HallUp] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, fsm.Floor, false)
		return true
	}

	return false
}

// chooseDirectionAndMove picks the next direction based on pending orders
// and starts the motor. Goes idle if no orders remain.
func (fsm *ElevatorFSM) chooseDirectionAndMove() {
	switch {
	case fsm.ordersAbove() && (fsm.Direction == elevio.MD_Up || fsm.Direction == elevio.MD_Stop):
		fsm.Direction = elevio.MD_Up
		elevio.SetMotorDirection(elevio.MD_Up)
		fsm.State = Moving
	case fsm.ordersBelow() && (fsm.Direction == elevio.MD_Down || fsm.Direction == elevio.MD_Stop):
		fsm.Direction = elevio.MD_Down
		elevio.SetMotorDirection(elevio.MD_Down)
		fsm.State = Moving
	case fsm.ordersAbove():
		fsm.Direction = elevio.MD_Up
		elevio.SetMotorDirection(elevio.MD_Up)
		fsm.State = Moving
	case fsm.ordersBelow():
		fsm.Direction = elevio.MD_Down
		elevio.SetMotorDirection(elevio.MD_Down)
		fsm.State = Moving
	default:
		fsm.Direction = elevio.MD_Stop
		elevio.SetMotorDirection(elevio.MD_Stop)
		fsm.State = Idle
	}
}

func (fsm *ElevatorFSM) ordersAbove() bool {
	for f := fsm.Floor + 1; f < config.N; f++ {
		for btn := range OrderTypes {
			if fsm.Orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func (fsm *ElevatorFSM) ordersBelow() bool {
	for f := 0; f < fsm.Floor; f++ {
		for btn := range OrderTypes {
			if fsm.Orders[f][btn] {
				return true
			}
		}
	}
	return false
}
