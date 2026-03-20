package controller

import (
	"FinalProject_G92/config"
	"FinalProject_G92/controller/elevio"
	"FinalProject_G92/models"
	"fmt"
	"time"
)

// ElevatorController bridges the physical elevator hardware and the coordinator.
// It polls sensors, runs the FSM, emits new/removed orders, and receives assigned orders.
func ElevatorController(assignCh, orderCh, rmOrderCh chan models.Order, statusCh chan models.StatusMessage) {

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

	// Watchdog: if no floor is reached within 2x travel time, mark elevator as non-operational
	motorTimer := time.NewTimer(config.BetweenFloorsDuration * 2)

	for {
		select {
		case floor := <-floorCh:
			fmt.Printf("Floor %d\n", floor)

			motorTimer.Reset(config.BetweenFloorsDuration * 2)
			hadHallUp := fsm.Orders[floor][elevio.BT_HallUp]
			hadHallDown := fsm.Orders[floor][elevio.BT_HallDown]
			hadCab := fsm.Orders[floor][elevio.BT_Cab]
			prevState := fsm.State

			fsm.OnFloorArrival(floor)

			if hadHallUp && !fsm.Orders[floor][elevio.BT_HallUp] {
				rmOrderCh <- models.Order{Floor: floor, Dir: true}
			}
			if hadHallDown && !fsm.Orders[floor][elevio.BT_HallDown] {
				rmOrderCh <- models.Order{Floor: floor, Dir: false}
			}
			if hadCab && !fsm.Orders[floor][elevio.BT_Cab] {
				rmOrderCh <- models.Order{Floor: floor, Cab: true}
			}
			if prevState != DoorOpen && fsm.State == DoorOpen {
				go closeDoors(&fsm, rmOrderCh)
			}

			statusCh <- models.StatusMessage{Floor: floor, Direction: int(fsm.Direction), Operational: true}

		case btn := <-btnCh:
			fmt.Printf("Button: floor %d, type %d\n", btn.Floor, btn.Button)

			var newOrder models.Order
			newOrder.Floor = btn.Floor

			switch btn.Button {
			case elevio.BT_HallUp:
				newOrder.Dir = true
				orderCh <- newOrder
			case elevio.BT_HallDown:
				newOrder.Dir = false
				orderCh <- newOrder

			case elevio.BT_Cab:
				newOrder.Cab = true
				prevState := fsm.State
				fsm.OnButtonPress(btn.Floor, btn.Button)

				if fsm.Orders[newOrder.Floor][elevio.BT_Cab] {
					orderCh <- newOrder
				}

				if prevState == Idle && fsm.State == DoorOpen {
					go closeDoors(&fsm, rmOrderCh)
				}
			}

		case stop := <-stopCh:
			if stop {
				fsm.OnStopButton()
			}
			statusCh <- models.StatusMessage{Floor: fsm.Floor, Direction: int(fsm.Direction), Operational: !stop}

		case obstr := <-obstrCh:
			fsm.OnObstruction(obstr)

			statusCh <- models.StatusMessage{Floor: fsm.Floor, Direction: int(fsm.Direction), Operational: !obstr}

		case assignedOrder := <-assignCh:
			prevState := fsm.State
			if assignedOrder.Cab {
				fsm.OnButtonPress(assignedOrder.Floor, elevio.BT_Cab)
				if !fsm.Orders[assignedOrder.Floor][elevio.BT_Cab] {
					rmOrderCh <- models.Order{Floor: assignedOrder.Floor, Cab: true}
				}
			} else {
				btn := elevio.BT_HallUp
				if !assignedOrder.Dir {
					btn = elevio.BT_HallDown
				}
				fsm.OnButtonPress(assignedOrder.Floor, btn)
				if !fsm.Orders[assignedOrder.Floor][btn] {
					rmOrderCh <- assignedOrder
				}
			}
			if prevState == Idle && fsm.State == DoorOpen {
				go closeDoors(&fsm, rmOrderCh)
			}

		case <-motorTimer.C:
			if fsm.State == Moving {
				statusCh <- models.StatusMessage{Floor: fsm.Floor, Direction: int(fsm.Direction), Operational: false}
			}
			motorTimer.Reset(config.BetweenFloorsDuration * 2)

		}
	}
}

// closeDoors waits for the door open duration, handles obstruction, and implements
// direction-change announcement: if there's an opposite-direction hall call at this floor,
// it clears it and keeps the door open for another cycle before closing.
func closeDoors(fsm *ElevatorFSM, rmOrderCh chan models.Order) {
	for {
		time.Sleep(config.DoorOpenDuration)
		for fsm.Obstructed {
			time.Sleep(100 * time.Millisecond)
		}
		hadUp := fsm.Orders[fsm.Floor][elevio.BT_HallUp]
		hadDown := fsm.Orders[fsm.Floor][elevio.BT_HallDown]
		if (hadUp || hadDown) && fsm.clearOrdersAtFloor() {
			if hadUp && !fsm.Orders[fsm.Floor][elevio.BT_HallUp] {
				rmOrderCh <- models.Order{Floor: fsm.Floor, Dir: true}
			}
			if hadDown && !fsm.Orders[fsm.Floor][elevio.BT_HallDown] {
				rmOrderCh <- models.Order{Floor: fsm.Floor, Dir: false}
			}
			continue
		}
		break
	}
	elevio.SetDoorOpenLamp(false)
	fsm.chooseDirectionAndMove()
}
