package hardware

import (
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/network"
	"fmt"
)

func HardwareManager(stateCh chan ElevatorState, orderCh, rmOrderCh chan network.Order) {

	var state ElevatorState
	ElevInit(&state)

	floorCh := make(chan int)
	btnCh := make(chan elevio.ButtonEvent)
	stopCh := make(chan bool)
	obstrCh := make(chan bool)

	go elevio.PollFloorSensor(floorCh)
	go elevio.PollButtons(btnCh)
	go elevio.PollStopButton(stopCh)
	go elevio.PollObstructionSwitch(obstrCh)

	for {
		select {
		case floor := <-floorCh:
			elevio.SetFloorIndicator(floor)
			//logic for if stopping
			fmt.Printf("on floor %d", floor)
			state.CurrentFloor = floor
			stateCh <- state
		case btn := <-btnCh:
			//Process button call
			var no network.Order
			no.Floor = btn.Floor
			switch btn.Button {

			case elevio.BT_HallDown:
				no.Cab = false
				no.Dir = false

				fmt.Printf("Call to %d: DOWN", no.Floor)

			case elevio.BT_HallUp:
				no.Cab = false
				no.Dir = true
				fmt.Printf("Call to %d: UP", no.Floor)

			case elevio.BT_Cab:
				no.Cab = true
				no.Dir = false
				fmt.Printf("Cab call to %d", no.Floor)
			}

			orderCh <- no

		case stop := <-stopCh:
			state.Stopped = stop

			//stop handling
			stateCh <- state

		case obstr := <-obstrCh:
			//handle obstruction
			if obstr {

				fmt.Println("Obstruction!")
			} else {
				fmt.Println("Cleared!")
			}
			stateCh <- state

		}

	}

}

func Mover(targetFloor, currentFloor int) {
	if targetFloor > currentFloor {
		elevio.SetMotorDirection(elevio.MD_Up)
	} else if targetFloor < currentFloor {
		elevio.SetMotorDirection(elevio.MD_Down)
	} else {
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}
