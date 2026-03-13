package hardware

import (
	"FinalProject_G92/config"
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/models"
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
	State         FSMState
	Floor         int
	Direction     elevio.MotorDirection
	Orders        [config.N][OrderTypes]bool
	Obstructed    bool
	ClearedOrders chan models.Order
}

// --- Event handlers ---

func (fsm *ElevatorFSM) OnButtonPress(floor int, btn elevio.ButtonType) {
	fsm.Orders[floor][btn] = true
	elevio.SetButtonLamp(btn, floor, true)

	switch fsm.State {
	case Idle:
		if floor == fsm.Floor {
			fsm.openDoorAndServe()
		} else {
			fsm.chooseDirectionAndMove()
		}
	case DoorOpen, Moving, Stopped:
		// stored, serveFloor loop or future stop will handle it
	}
}

func (fsm *ElevatorFSM) OnFloorArrival(floor int) {
	fsm.Floor = floor
	elevio.SetFloorIndicator(floor)

	if fsm.shouldStop() {
		elevio.SetMotorDirection(elevio.MD_Stop)
		fsm.openDoorAndServe()
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

// --- Door serving ---

func (fsm *ElevatorFSM) openDoorAndServe() {
	fsm.State = DoorOpen
	elevio.SetDoorOpenLamp(true)
	go fsm.serveFloor()
}

func (fsm *ElevatorFSM) serveFloor() {
	for {
		fsm.clearOneHallCall()
		time.Sleep(config.DoorOpenDuration)
		for fsm.Obstructed {
			time.Sleep(100 * time.Millisecond)
		}
		if fsm.hasOrdersAtFloor() {
			continue
		}
		break
	}
	elevio.SetDoorOpenLamp(false)
	fsm.chooseDirectionAndMove()
}

// --- Order clearing ---

func (fsm *ElevatorFSM) clearOneHallCall() {
	fsm.clearOrder(elevio.BT_Cab)

	if fsm.preferUpDirection() {
		if fsm.Orders[fsm.Floor][elevio.BT_HallUp] {
			fsm.clearOrder(elevio.BT_HallUp)
		} else {
			fsm.clearOrder(elevio.BT_HallDown)
		}
	} else {
		if fsm.Orders[fsm.Floor][elevio.BT_HallDown] {
			fsm.clearOrder(elevio.BT_HallDown)
		} else {
			fsm.clearOrder(elevio.BT_HallUp)
		}
	}
}

func (fsm *ElevatorFSM) clearOrder(btn elevio.ButtonType) {
	if !fsm.Orders[fsm.Floor][btn] {
		return
	}
	fsm.Orders[fsm.Floor][btn] = false
	elevio.SetButtonLamp(btn, fsm.Floor, false)

	order := models.Order{Floor: fsm.Floor}
	switch btn {
	case elevio.BT_Cab:
		order.Cab = true
	case elevio.BT_HallUp:
		order.Dir = true
	}
	fsm.ClearedOrders <- order
}

// --- Direction logic ---

func (fsm *ElevatorFSM) preferUpDirection() bool {
	switch {
	case fsm.ordersAbove():
		return true
	case fsm.ordersBelow():
		return false
	default:
		return fsm.Direction != elevio.MD_Down
	}
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

// --- Order queries ---

func (fsm *ElevatorFSM) hasOrdersAtFloor() bool {
	return fsm.Orders[fsm.Floor][elevio.BT_HallUp] ||
		fsm.Orders[fsm.Floor][elevio.BT_HallDown] ||
		fsm.Orders[fsm.Floor][elevio.BT_Cab]
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
