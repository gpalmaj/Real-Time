package fsm
 
import (
	"heis/config"
	"heis/driver/elevio"
	"time"
)
 
func RunFSM(
	io ElevatorIO,
	assignmentCh <-chan [config.N_FLOORS][config.N_BUTTONS]bool,
	finReqCh chan<- elevio.ButtonEvent,
	stateCh chan<- ElevatorStateMsg,
	floorCh <-chan int,
	obstrCh <-chan bool,
) {
	e := elevatorInit()
	doorTimer := NewTimer()
	motorTimer := NewTimer()
 
	obstrActive := false
	obstrCounter := 0
	obstrStoredFloor := config.FloorUnknown
 
	var pendingFinReqs []elevio.ButtonEvent
 
	e.floor = findFloor(io, floorCh)
	io.SetFloorIndicator(e.floor)
	e.dirn = D_Stop
	e.state = ES_Idle
	stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}
 
	for {
		select {
 
		case assignments := <-assignmentCh:
			e.requests = assignments
			switch e.state {
			case ES_Idle:
				pair := ChooseDirection(e)
				e.dirn, e.state = pair.dirn, pair.state
				switch e.state {
				case ES_DoorOpen:
					io.SetDoorOpenLamp(true)
					doorTimer.Start(e.elevConfig.DoorOpenDuration)
					pendingFinReqs = collectClearedEvents(&e, e.floor)
					setAllLights(io, e)
				case ES_Moving:
					io.SetMotorDirection(e.dirn)
					motorTimer.Start(config.MotorLossTimeout)
				case ES_Idle:
				}
			case ES_Moving:
				pair := ChooseDirection(e)
				if pair.state == ES_Idle {
					io.SetMotorDirection(D_Stop)
					motorTimer.Stop()
					e.dirn = D_Stop
					e.state = ES_Idle
				}
			case ES_DoorOpen:
				setAllLights(io, e)
			}
			stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}
 
		case newFloor := <-floorCh:
			motorTimer.Stop()
			e.floor = newFloor
			io.SetFloorIndicator(newFloor)
 
			if e.state != ES_Moving {
				stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}
				continue
			}
 
			if ShouldStop(e) {
				io.SetMotorDirection(D_Stop)
				io.SetDoorOpenLamp(true)
				pendingFinReqs = collectClearedEvents(&e, e.floor)
				doorTimer.Start(e.elevConfig.DoorOpenDuration)
				setAllLights(io, e)
				e.state = ES_DoorOpen
			} else {
				motorTimer.Start(config.MotorLossTimeout)
			}
			stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}
 
		case <-doorTimer.C():
			if e.state != ES_DoorOpen {
				continue
			}
 
			if obstrActive {
				doorTimer.Start(e.elevConfig.DoorOpenDuration)
				if obstrCounter == config.ObstrTripsBeforeFloorInvalid {
					obstrStoredFloor = e.floor
					e.floor = config.FloorUnknown
					stateCh <- ElevatorStateMsg{Floor: config.FloorUnknown, Dirn: e.dirn, State: e.state}
				}
				obstrCounter++
				continue
			}
 
			if e.floor == config.FloorUnknown && obstrStoredFloor != config.FloorUnknown {
				e.floor = obstrStoredFloor
				obstrStoredFloor = config.FloorUnknown
			}
 
			flushPendingFinReqs(pendingFinReqs, finReqCh)
			pendingFinReqs = nil
 
			pair := ChooseDirection(e)
			e.dirn, e.state = pair.dirn, pair.state
 
			switch e.state {
			case ES_DoorOpen:
				io.SetDoorOpenLamp(true)
				doorTimer.Start(e.elevConfig.DoorOpenDuration)
				pendingFinReqs = collectClearedEvents(&e, e.floor)
				setAllLights(io, e)
			case ES_Moving:
				io.SetDoorOpenLamp(false)
				io.SetMotorDirection(e.dirn)
				motorTimer.Start(config.MotorLossTimeout)
			case ES_Idle:
				io.SetDoorOpenLamp(false)
				io.SetMotorDirection(D_Stop)
			}
			stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}
 
		case <-motorTimer.C():
			e.floor = config.FloorUnknown
			e.dirn = D_Up
			e.state = ES_Moving
			stateCh <- ElevatorStateMsg{Floor: config.FloorUnknown, Dirn: e.dirn, State: e.state}
			io.SetMotorDirection(D_Up)
 
		case obstr := <-obstrCh:
			obstrActive = obstr
			if obstr {
				if e.state == ES_DoorOpen {
					doorTimer.Start(e.elevConfig.DoorOpenDuration)
				}
			} else {
				obstrCounter = 0
				if e.state == ES_DoorOpen {
					doorTimer.Start(e.elevConfig.DoorOpenDuration)
				}
				stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}
			}
		}
	}
}
 
func setAllLights(io ElevatorIO, e Elevator) {
	for f := 0; f < config.N_FLOORS; f++ {
		io.SetButtonLamp(f, B_Cab, e.requests[f][B_Cab])
	}
}
 
func findFloor(io ElevatorIO, floorCh <-chan int) int {
	if f := elevio.GetFloor(); f != config.FloorUnknown {
		return f
	}
	io.SetMotorDirection(D_Up)
	reversed := false
	t := time.NewTimer(config.MotorLossTimeout)
	defer t.Stop()
	for {
		select {
		case floor := <-floorCh:
			io.SetMotorDirection(D_Stop)
			return floor
		case <-t.C:
			if !reversed {
				io.SetMotorDirection(D_Down)
				reversed = true
			}
		}
	}
}