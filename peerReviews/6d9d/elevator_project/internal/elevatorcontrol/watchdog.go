package elevatorcontrol

import (
	"time"

	"elevator/internal/config"
	"elevator/internal/elevio"
	"elevator/internal/types"
)

func (fsm *elevatorFSM) restartMovementTimer() {
	if fsm.movementTimer == nil {
		fsm.movementTimer = time.NewTimer(config.MovementWatchdogTimeoutDuration)
		fsm.movementTimerCh = fsm.movementTimer.C
		return
	}

	// Stop the timer and drain timeout signal before reuse
	if !fsm.movementTimer.Stop() {
		select {
		case <-fsm.movementTimer.C:
		default:
		}
	}

	fsm.movementTimer.Reset(config.MovementWatchdogTimeoutDuration)
	fsm.movementTimerCh = fsm.movementTimer.C
}

func (fsm *elevatorFSM) stopMovementTimer() {
	if fsm.movementTimer == nil {
		fsm.movementTimerCh = nil
		return
	}

	// Stop the timer and drain timeout signal before reuse
	if !fsm.movementTimer.Stop() {
		select {
		case <-fsm.movementTimer.C:
		default:
		}
	}

	fsm.movementTimerCh = nil
}

func (fsm *elevatorFSM) handleMovementTimeout() {
	if fsm.state.Behaviour != types.BehaviourMoving {
		return
	}

	elevio.SetMotorDirection(elevio.MD_Stop)
	fsm.stopMovementTimer()

	// Set elevator unavailabile, update states(no stop state therefore idle)
	fsm.movementTimedOut = true
	fsm.state.Behaviour = types.BehaviourIdle
	fsm.state.Dir = types.DirectionStop
	fsm.updateAvailability()

	fsm.publishStateIfChanged()
}
