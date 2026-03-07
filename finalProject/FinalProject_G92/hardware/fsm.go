package hardware

import (
	"FinalProject_G92/config"
	"FinalProject_G92/hardware/elevio"
	"fmt"
	"time"
)

type FSMState int

const (
	Idle     FSMState = 0
	Moving   FSMState = 1
	DoorOpen FSMState = 2
	Stopped  FSMState = 3
)

const OrderTypes = 3

type ElevatorFSM struct {
	State     FSMState
	Floor     int
	Direction elevio.MotorDirection
	Orders    [config.N][OrderTypes]bool // floor x button type
}

func (fsm *ElevatorFSM) OnButtonPress(floor int, btn elevio.ButtonType) {
	fsm.Orders[floor][btn] = true
	elevio.SetButtonLamp(btn, floor, true)

	switch fsm.State {
	case Idle:
		fsm.chooseDirectionAndMove()
	case DoorOpen:
		if floor == fsm.Floor {
			// Re-open door (reset timer handled by caller)
			fsm.clearOrdersAtFloor()
		}
	case Moving, Stopped:
		// Order is stored, will be served when appropriate
	}
}

func (fsm *ElevatorFSM) OnFloorArrival(floor int) {
	fsm.Floor = floor
	elevio.SetFloorIndicator(floor)

	if fsm.shouldStop() {
		elevio.SetMotorDirection(elevio.MD_Stop)
		fsm.State = DoorOpen
		elevio.SetDoorOpenLamp(true)
		fsm.clearOrdersAtFloor()

		go func() {
			time.Sleep(config.DoorOpenDuration)
			elevio.SetDoorOpenLamp(false)
			fsm.chooseDirectionAndMove()
		}()
	}
}

func (fsm *ElevatorFSM) OnStopButton() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	fsm.State = Stopped
	fmt.Println("Elevator stopped")
}

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

func (fsm *ElevatorFSM) clearOrdersAtFloor() {
	for btn := range OrderTypes {
		fsm.Orders[fsm.Floor][btn] = false
		elevio.SetButtonLamp(elevio.ButtonType(btn), fsm.Floor, false)
	}
}

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
