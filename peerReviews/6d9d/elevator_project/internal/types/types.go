package types

// Direction represents the movement direction of an elevator.
type Direction int

const (
	DirectionDown Direction = -1
	DirectionStop Direction = 0
	DirectionUp   Direction = 1
)

// Behaviour represents the current operating mode of an elevator.
type Behaviour int

const (
	BehaviourIdle Behaviour = iota
	BehaviourMoving
	BehaviourDoorOpen
)

// ButtonType identifies which elevator button generated a request.
type ButtonType int

const (
	ButtonHallUp   ButtonType = 0
	ButtonHallDown ButtonType = 1
	ButtonCab      ButtonType = 2
)

// Hall call indices are used to address the shared hall-call table.
const (
	HallUpIndex   = 0
	HallDownIndex = 1
)

// ButtonEvent represents a button press detected by the elevator system.
type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

// FloorSensorEvent represents an update from the floor sensor.
// Floor should be a valid floor index.
type FloorSensorEvent struct {
	Floor int
}

// ObstructionEvent represents a change in the obstruction signal.
type ObstructionEvent struct {
	Obstructed bool
}

// DoorTimerEvent represents expiration of the door-open timer.
type DoorTimerEvent struct{}

// ElevatorState describes the information about an elevator
// that other nodes need in order to compute scheduling decisions.
type ElevatorState struct {
	ID             int
	Floor          int
	Dir            Direction
	Behaviour      Behaviour
	SequenceNumber uint32
	Available      bool
}

// HallCallState describes the distributed lifecycle of a hall request.
type HallCallState int

const (
	HallCallAbsent HallCallState = iota
	HallCallUnconfirmedAdd
	HallCallConfirmedAdd
	HallCallUnconfirmedRemove
)

// HallCall represents a shared hall request in the distributed system.
// Revision increases whenever the state changes so peers can resolve conflicts.
type HallCall struct {
	State      HallCallState
	Revision   uint32
	AssignedTo int
}

// NodeState is the complete state one elevator node shares with the network.
type NodeState struct {
	ElevatorState ElevatorState
	HallCalls     [][]HallCall
}

// Action represents what the elevator controller should do next.
type Action struct {
	Direction Direction
	OpenDoor  bool
}

// ClearRequestAction describes which requests should be cleared
// when the elevator serves a floor.
type ClearRequestAction struct {
	ClearCab      bool
	ClearHallUp   bool
	ClearHallDown bool
}

// LightState describes the desired state of all button lamps.
type LightState struct {
	Cab  []bool
	Hall [][]bool
}
