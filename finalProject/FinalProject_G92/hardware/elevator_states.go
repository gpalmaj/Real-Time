package hardware

import (
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/network"
)

type ElevatorState struct {
	CurrentDirection elevio.MotorDirection
	CurrentFloor     int
	Busy             bool //do we need a state where the elevator will not take any orders?
	Stopped          bool
	DoorsOpen        bool
	LightsOn         bool //maybe not needed, depends on light function
}

func ElevInit(eState *ElevatorState) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	eState.CurrentDirection = elevio.MD_Stop

	eState.CurrentFloor = elevio.GetFloor()

	eState.Busy = false //will the program continue orders when restarted?

	eState.Stopped = elevio.GetStop()

	elevio.SetDoorOpenLamp(false)
	eState.DoorsOpen = false

	for i := 0; i < network.N; i++ {
		elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
		elevio.SetButtonLamp(elevio.BT_Cab, i, false)
	}
	eState.LightsOn = false
}
