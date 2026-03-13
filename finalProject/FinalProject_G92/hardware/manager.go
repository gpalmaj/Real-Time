package hardware

import (
	"FinalProject_G92/config"
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/models"
	"fmt"
	"time"
)

func HardwareManager(assignCh, orderCh, rmOrderCh chan models.Order, statusCh chan models.StatusMessage) {

	var fsm ElevatorFSM
	fsm.ClearedOrders = make(chan models.Order, config.N*OrderTypes)
	ElevInit(&fsm)

	floorCh := make(chan int)
	btnCh := make(chan elevio.ButtonEvent)
	stopCh := make(chan bool)
	obstrCh := make(chan bool)

	go elevio.PollFloorSensor(floorCh)
	go elevio.PollButtons(btnCh)
	go elevio.PollStopButton(stopCh)
	go elevio.PollObstructionSwitch(obstrCh)

	motorTimer := time.NewTimer(config.BetweenFloorsDuration * 2)

	for {
		select {
		case floor := <-floorCh:
			fmt.Printf("Floor %d\n", floor)
			motorTimer.Reset(config.BetweenFloorsDuration * 2)
			fsm.OnFloorArrival(floor)
			statusCh <- models.StatusMessage{Floor: floor, Direction: int(fsm.Direction), Operational: true}

		case cleared := <-fsm.ClearedOrders:
			rmOrderCh <- cleared

		case btn := <-btnCh:
			fmt.Printf("Button: floor %d, type %d\n", btn.Floor, btn.Button)

			var no models.Order
			no.Floor = btn.Floor

			switch btn.Button {
			case elevio.BT_HallUp:
				no.Dir = true
				orderCh <- no
			case elevio.BT_HallDown:
				no.Dir = false
				orderCh <- no
			case elevio.BT_Cab:
				no.Cab = true
				fsm.OnButtonPress(btn.Floor, btn.Button)
				orderCh <- no
			}

		case stop := <-stopCh:
			if stop {
				fsm.OnStopButton()
			}
			statusCh <- models.StatusMessage{Floor: fsm.Floor, Direction: int(fsm.Direction), Operational: !stop}

		case obstr := <-obstrCh:
			fsm.OnObstruction(obstr)
			statusCh <- models.StatusMessage{Floor: fsm.Floor, Direction: int(fsm.Direction), Operational: !obstr}

		case ao := <-assignCh:
			if ao.Cab {
				fsm.OnButtonPress(ao.Floor, elevio.BT_Cab)
			} else {
				btn := elevio.BT_HallUp
				if !ao.Dir {
					btn = elevio.BT_HallDown
				}
				fsm.OnButtonPress(ao.Floor, btn)
			}

		case <-motorTimer.C:
			if fsm.State == Moving {
				statusCh <- models.StatusMessage{Floor: fsm.Floor, Direction: int(fsm.Direction), Operational: false}
			}
			motorTimer.Reset(config.BetweenFloorsDuration * 2)

		}
	}
}
