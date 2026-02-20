package hardware

import (
	"FinalProject_G92/hardware/elevio"
	"fmt"
)

func ExecuteOrder(btnCh <-chan elevio.ButtonEvent) {
	for {
		evt := <-btnCh
		fmt.Printf("Button pressed on floor %d, type %v\n", evt.Floor, evt.Button)

		obstructionPressed := elevio.GetObstruction()
		stopPressed := elevio.GetStop()

		state := ElevatorState{}
		btn := elevio.ButtonEvent{}

		switch evt.Button {

		case elevio.BT_HallUp: //these need extra code for assigning different calls

		case elevio.BT_HallDown: //see above

		case elevio.BT_Cab:
			if btn.Floor > state.CurrentFloor {
				state.CurrentDirection = elevio.MD_Up
				elevio.SetMotorDirection(elevio.MD_Up)
			}
			if btn.Floor < state.CurrentFloor {
				state.CurrentDirection = elevio.MD_Down
				elevio.SetMotorDirection(elevio.MD_Down)
			}
			if btn.Floor == state.CurrentFloor {
				state.CurrentDirection = elevio.MD_Stop
			}

		}

		if obstructionPressed {

		}

		if stopPressed {
			lastDirection := state.CurrentDirection
			if !state.Stopped {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				//continue last direction, orders not forgotten
				elevio.SetMotorDirection(lastDirection)
			}
		}

	}
}
