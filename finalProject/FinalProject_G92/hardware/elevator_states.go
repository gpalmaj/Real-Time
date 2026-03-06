package hardware

import (
	"FinalProject_G92/config"
	"FinalProject_G92/hardware/elevio"
)

type ElevatorState struct {
	CurrentDirection elevio.MotorDirection
	CurrentFloor     int
	Busy             bool
	Stopped          bool
	DoorsOpen        bool
	LightsOn         bool
}

func ElevInit(eState *ElevatorState) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	eState.CurrentDirection = elevio.MD_Stop

	eState.CurrentFloor = elevio.GetFloor()

	eState.Busy = false

	eState.Stopped = elevio.GetStop()

	elevio.SetDoorOpenLamp(false)
	eState.DoorsOpen = false

	for i := 0; i < config.N; i++ {
		elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
		elevio.SetButtonLamp(elevio.BT_Cab, i, false)
	}
	eState.LightsOn = false
}
