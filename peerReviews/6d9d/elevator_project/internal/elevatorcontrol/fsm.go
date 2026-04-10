package elevatorcontrol

import (
	"time"

	"elevator/internal/config"
	"elevator/internal/elevio"
	"elevator/internal/types"
)

type elevatorFSM struct {
	state types.ElevatorState

	lastPublishedState types.ElevatorState
	hasPublishedState  bool

	obstructed       bool
	movementTimedOut bool

	doorTimer       *time.Timer
	doorTimerCh     <-chan time.Time
	movementTimer   *time.Timer
	movementTimerCh <-chan time.Time

	servedFloorCh chan<- int
	stateCh       chan<- types.ElevatorState
}

func newElevatorFSM(
	selfID int,
	initialFloor int,
	stateCh chan<- types.ElevatorState,
	servedFloorCh chan<- int,
) *elevatorFSM {
	fsm := &elevatorFSM{
		state: types.ElevatorState{
			ID:        selfID,
			Floor:     initialFloor,
			Dir:       types.DirectionStop,
			Behaviour: types.BehaviourIdle,
			Available: true,
		},
		stateCh:       stateCh,
		servedFloorCh: servedFloorCh,
	}

	elevio.SetDoorOpenLamp(false)

	if initialFloor >= 0 {
		elevio.SetFloorIndicator(initialFloor)
		elevio.SetMotorDirection(elevio.MD_Stop)
	} else {
		// Start between floors: orderlogic handles direction.
		fsm.state.Dir = types.DirectionStop
		fsm.state.Behaviour = types.BehaviourIdle
		elevio.SetMotorDirection(elevio.MD_Stop)
	}

	return fsm
}

func (fsm *elevatorFSM) publishStateIfChanged() {
	if fsm.hasPublishedState &&
		fsm.state.Floor == fsm.lastPublishedState.Floor &&
		fsm.state.Dir == fsm.lastPublishedState.Dir &&
		fsm.state.Behaviour == fsm.lastPublishedState.Behaviour &&
		fsm.state.Available == fsm.lastPublishedState.Available {
		return
	}

	select {
	case fsm.stateCh <- fsm.state:
	default:
	}
	fsm.lastPublishedState = fsm.state
	fsm.hasPublishedState = true
}

func (fsm *elevatorFSM) executeAction(action types.Action) {
	// While the door is open, ignore all movement/stop actions.
	// Only allow another OpenDoor action to extend the open time.
	if fsm.state.Behaviour == types.BehaviourDoorOpen {
		if action.OpenDoor {
			fsm.enterDoorOpen()
		}
		return
	}

	// Ignore open door requests between floors(guard)
	if action.OpenDoor {
		if fsm.state.Floor >= 0 {
			fsm.enterDoorOpen()
		}
		return
	}

	switch action.Direction {
	case types.DirectionUp, types.DirectionDown:
		// To avoid spamming
		if fsm.state.Behaviour == types.BehaviourMoving && fsm.state.Dir == action.Direction {
			return
		}
		fsm.startMoving(action.Direction)

	case types.DirectionStop:
		// Dont stop between floors
		if fsm.state.Floor == -1 {
			return
		}
		// Dont go to idle if already in idle
		if fsm.state.Behaviour == types.BehaviourIdle && fsm.state.Dir == types.DirectionStop {
			return
		}
		fsm.enterIdle()
	}
}

func (fsm *elevatorFSM) startMoving(dir types.Direction) {
	fsm.stopDoorTimer()
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(toMotorDirection(dir))

	fsm.state.Dir = dir
	fsm.state.Behaviour = types.BehaviourMoving
	fsm.updateAvailability()
	fsm.state.Floor = -1

	fsm.restartMovementTimer()
	fsm.publishStateIfChanged()
}

func (fsm *elevatorFSM) enterDoorOpen() {
	wasDoorOpen := fsm.state.Behaviour == types.BehaviourDoorOpen

	fsm.stopMovementTimer()
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)

	fsm.state.Behaviour = types.BehaviourDoorOpen
	fsm.updateAvailability()

	fsm.restartDoorTimer()
	fsm.publishStateIfChanged()

	// Report served floor only the first time door open at the current floor
	if !wasDoorOpen && fsm.state.Floor >= 0 {
		select {
		case fsm.servedFloorCh <- fsm.state.Floor:
		default:
			// non-blocking to avoid crashing issues when main is not ready
		}
	}
}

func (fsm *elevatorFSM) enterIdle() {
	fsm.stopDoorTimer()
	fsm.stopMovementTimer()

	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(false)

	fsm.state.Behaviour = types.BehaviourIdle
	fsm.state.Dir = types.DirectionStop
	fsm.updateAvailability()

	fsm.publishStateIfChanged()
}

func (fsm *elevatorFSM) closeDoor() {
	elevio.SetDoorOpenLamp(false)
	fsm.stopDoorTimer()
}

func (fsm *elevatorFSM) handleFloorArrival(floor int) {
	fsm.state.Floor = floor
	elevio.SetFloorIndicator(floor)

	// Check if the elevator is back from a timeout, proof of return/healty
	if fsm.movementTimedOut {
		fsm.movementTimedOut = false
		fsm.updateAvailability()
	}

	// While moving, restart watchdog timer every floor arrival
	if fsm.state.Behaviour == types.BehaviourMoving {
		fsm.restartMovementTimer()
	}

	fsm.publishStateIfChanged()
}

func (fsm *elevatorFSM) handleObstructionChange(obstructed bool) {
	fsm.obstructed = obstructed
	fsm.updateAvailability()

	// Full door open period after the obstruction is removed.
	if fsm.state.Behaviour == types.BehaviourDoorOpen && !obstructed {
		fsm.restartDoorTimer()
	}

	fsm.publishStateIfChanged()
}

func (fsm *elevatorFSM) updateAvailability() {
	fsm.state.Available = !fsm.obstructed && !fsm.movementTimedOut
}

func (fsm *elevatorFSM) handleDoorTimeout() {
	// Close door when open and no obstruction is present.
	if fsm.state.Behaviour != types.BehaviourDoorOpen {
		return
	}
	if fsm.obstructed {
		return
	}

	fsm.closeDoor()

	// Go to idle, but keep dir as "memory" for order managment
	fsm.state.Behaviour = types.BehaviourIdle
	fsm.updateAvailability()

	fsm.publishStateIfChanged()
}

func (fsm *elevatorFSM) restartDoorTimer() {
	if fsm.doorTimer == nil {
		fsm.doorTimer = time.NewTimer(config.DoorOpenDuration)
		fsm.doorTimerCh = fsm.doorTimer.C
		return
	}

	// Stop the timer and drain timeout signal before reuse
	if !fsm.doorTimer.Stop() {
		select {
		case <-fsm.doorTimer.C:
		default:
		}
	}

	fsm.doorTimer.Reset(config.DoorOpenDuration)
	fsm.doorTimerCh = fsm.doorTimer.C
}

func (fsm *elevatorFSM) stopDoorTimer() {
	if fsm.doorTimer == nil {
		fsm.doorTimerCh = nil
		return
	}

	// Stop the timer and drain timeout signal before reuse
	if !fsm.doorTimer.Stop() {
		select {
		case <-fsm.doorTimer.C:
		default:
		}
	}

	fsm.doorTimerCh = nil
}
