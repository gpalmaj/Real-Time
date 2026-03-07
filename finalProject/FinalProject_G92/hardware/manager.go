package hardware

import (
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/models"
	"fmt"
)

func HardwareManager(orderCh, rmOrderCh chan models.Order, statusCh chan models.StatusMessage) {

	var fsm ElevatorFSM
	ElevInit(&fsm)

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
			fmt.Printf("Floor %d\n", floor)

			// Check if we had orders at this floor before the FSM clears them
			hadHallUp := fsm.Orders[floor][elevio.BT_HallUp]
			hadHallDown := fsm.Orders[floor][elevio.BT_HallDown]
			hadCab := fsm.Orders[floor][elevio.BT_Cab]

			fsm.OnFloorArrival(floor)

			// Send remove orders for any orders the FSM cleared at this floor
			if hadHallUp && !fsm.Orders[floor][elevio.BT_HallUp] {
				rmOrderCh <- models.Order{Floor: floor, Dir: true}
			}
			if hadHallDown && !fsm.Orders[floor][elevio.BT_HallDown] {
				rmOrderCh <- models.Order{Floor: floor, Dir: false}
			}
			if hadCab && !fsm.Orders[floor][elevio.BT_Cab] {
				rmOrderCh <- models.Order{Floor: floor, Cab: true}
			}

			statusCh <- models.StatusMessage{Floor: floor, Direction: int(fsm.Direction), Operational: true}

		case btn := <-btnCh:
			fmt.Printf("Button: floor %d, type %d\n", btn.Floor, btn.Button)

			var no models.Order
			no.Floor = btn.Floor

			switch btn.Button {
			case elevio.BT_HallUp:
				no.Dir = true
			case elevio.BT_HallDown:
				no.Dir = false
			case elevio.BT_Cab:
				no.Cab = true
			}

			orderCh <- no
			fsm.OnButtonPress(btn.Floor, btn.Button)

		case stop := <-stopCh:
			if stop {
				fsm.OnStopButton()
			}
			statusCh <- models.StatusMessage{Floor: fsm.Floor, Direction: int(fsm.Direction), Operational: !stop}

		case obstr := <-obstrCh:
			if obstr {
				fmt.Println("Obstruction!")
			} else {
				fmt.Println("Obstruction cleared")
			}
			statusCh <- models.StatusMessage{Floor: fsm.Floor, Direction: int(fsm.Direction), Operational: !obstr}
		}
	}
}
