package controller

import (
	"FinalProject_G92/config"
	"FinalProject_G92/controller/elevio"
	"time"
)

// ElevInit puts the elevator into a known-good state: stops the motor,
// turns off all lamps, and drives to the nearest floor if between floors.
func ElevInit(eState *ElevatorFSM) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	eState.Direction = elevio.MD_Stop
	eState.State = Idle
	elevio.SetDoorOpenLamp(false)

	for i := range config.N {
		elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
		elevio.SetButtonLamp(elevio.BT_Cab, i, false)
	}

	if elevio.GetFloor() == -1 {
		elevio.SetMotorDirection(elevio.MD_Down)
		for elevio.GetFloor() == -1 {
			time.Sleep(10 * time.Millisecond)
		}
		elevio.SetMotorDirection(elevio.MD_Stop)
	}

	eState.Floor = elevio.GetFloor()

}
