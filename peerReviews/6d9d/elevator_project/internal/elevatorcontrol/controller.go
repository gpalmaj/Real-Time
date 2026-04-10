package elevatorcontrol

import (
	"elevator/internal/config"
	"elevator/internal/elevio"
	"elevator/internal/types"
)

func Run(
	selfID int,
	actionCh <-chan types.Action,
	buttonEventCh chan<- types.ButtonEvent,
	stateCh chan<- types.ElevatorState,
	servedFloorCh chan<- int,
) {
	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstruction := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstruction)

	localFSM := newElevatorFSM(
		selfID,
		elevio.GetFloor(),
		stateCh,
		servedFloorCh,
	)

	localFSM.publishStateIfChanged()

	for {
		select {
		case btn := <-drvButtons:
			// Additionally safty for the first floor and top floor
			if btn.Floor == 0 && btn.Button == elevio.BT_HallDown {
				continue
			}
			if btn.Floor == config.NumberOfFloors-1 && btn.Button == elevio.BT_HallUp {
				continue
			}
			buttonEventCh <- toButtonEvent(btn)

		case floor := <-drvFloors:
			localFSM.handleFloorArrival(floor)

		case obstructed := <-drvObstruction:
			localFSM.handleObstructionChange(obstructed)

		case action := <-actionCh:
			localFSM.executeAction(action)

		case <-localFSM.doorTimerCh:
			localFSM.handleDoorTimeout()

		case <-localFSM.movementTimerCh:
			localFSM.handleMovementTimeout()
		}
	}
}
