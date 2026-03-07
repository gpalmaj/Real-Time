package hardware

import (
	"FinalProject_G92/config"
	"FinalProject_G92/hardware/elevio"
)

func ElevInit(eState *ElevatorFSM) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	eState.Direction = elevio.MD_Stop

	eState.Floor = elevio.GetFloor()

	eState.State = Idle

	elevio.SetDoorOpenLamp(false)

	for i := 0; i < config.N; i++ {
		elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
		elevio.SetButtonLamp(elevio.BT_Cab, i, false)
	}
}
