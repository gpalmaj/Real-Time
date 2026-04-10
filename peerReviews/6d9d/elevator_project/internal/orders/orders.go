package orders

import (
	"fmt"
	"os"

	"elevator/internal/config"
	"elevator/internal/types"
)

// OrderManager maintains the local view of all requests and elevator states.
// It decides what this elevator should do based on the shared system state.
type OrderManager struct {
	selfState types.ElevatorState

	// Local hall-call contribution from this node only.
	selfHallCalls [][]types.HallCall

	// Merged/shared hall-call view used for scheduling and lamps.
	hallCalls [][]types.HallCall

	cabCalls []bool

	peerStates    map[int]types.ElevatorState
	peerHallCalls map[int][][]types.HallCall

	storageDir string
}

// NewOrderManager creates an order manager with empty request tables.
func NewOrderManager(numFloors int, selfID int) *OrderManager {
	selfHallCalls := make([][]types.HallCall, numFloors)
	hallCalls := make([][]types.HallCall, numFloors)

	for floor := 0; floor < numFloors; floor++ {
		selfHallCalls[floor] = make([]types.HallCall, config.NumberOfHallDirections)
		hallCalls[floor] = make([]types.HallCall, config.NumberOfHallDirections)

		for dir := 0; dir < config.NumberOfHallDirections; dir++ {
			selfHallCalls[floor][dir].AssignedTo = -1
			hallCalls[floor][dir].AssignedTo = -1
		}
	}

	orderManager := &OrderManager{
		selfState: types.ElevatorState{
			ID:        selfID,
			Floor:     -1,
			Dir:       types.DirectionStop,
			Behaviour: types.BehaviourIdle,
			Available: true,
		},
		selfHallCalls: selfHallCalls,
		hallCalls:     hallCalls,
		cabCalls:      make([]bool, numFloors),
		peerStates:    make(map[int]types.ElevatorState),
		peerHallCalls: make(map[int][][]types.HallCall),
		storageDir:    "data",
	}

	if err := os.MkdirAll(orderManager.storageDir, 0o755); err != nil {
		fmt.Println("warning: could not create storage directory:", err)
	}

	if err := orderManager.loadCabCalls(); err != nil {
		fmt.Println("warning: could not load cab calls:", err)
	}

	return orderManager
}

// RegisterButtonPress updates the local request state when a button is pressed.
func (om *OrderManager) RegisterButtonPress(event types.ButtonEvent) {
	if event.Floor < 0 || event.Floor >= len(om.cabCalls) {
		return
	}

	if event.Button < types.ButtonHallUp || event.Button > types.ButtonCab {
		return
	}

	switch event.Button {
	case types.ButtonCab:
		if !om.cabCalls[event.Floor] {
			om.cabCalls[event.Floor] = true
			_ = om.saveCabCalls()
		}

	case types.ButtonHallUp:
		if event.Floor == len(om.cabCalls)-1 {
			return
		}

		current := om.hallCalls[event.Floor][types.HallUpIndex].State
		if current != types.HallCallUnconfirmedAdd && current != types.HallCallConfirmedAdd {
			bestID := findBestElevator(
				event.Floor,
				types.HallUpIndex,
				om.selfState,
				om.peerStates,
				om.hallCalls,
			)

			om.selfHallCalls[event.Floor][types.HallUpIndex].State = types.HallCallUnconfirmedAdd
			om.selfHallCalls[event.Floor][types.HallUpIndex].Revision =
				om.hallCalls[event.Floor][types.HallUpIndex].Revision + 1
			om.selfHallCalls[event.Floor][types.HallUpIndex].AssignedTo = bestID
			om.recomputeHallCalls()
		}

	case types.ButtonHallDown:
		if event.Floor == 0 {
			return
		}

		current := om.hallCalls[event.Floor][types.HallDownIndex].State
		if current != types.HallCallUnconfirmedAdd && current != types.HallCallConfirmedAdd {
			bestID := findBestElevator(
				event.Floor,
				types.HallDownIndex,
				om.selfState,
				om.peerStates,
				om.hallCalls,
			)

			om.selfHallCalls[event.Floor][types.HallDownIndex].State = types.HallCallUnconfirmedAdd
			om.selfHallCalls[event.Floor][types.HallDownIndex].Revision =
				om.hallCalls[event.Floor][types.HallDownIndex].Revision + 1
			om.selfHallCalls[event.Floor][types.HallDownIndex].AssignedTo = bestID
			om.recomputeHallCalls()
		}
	}
}

// CopyHallCalls returns a deep copy of the hall-call table.
func (om *OrderManager) CopyHallCalls() [][]types.HallCall {
	hallCallsCopy := make([][]types.HallCall, len(om.hallCalls))

	for floor := 0; floor < len(om.hallCalls); floor++ {
		hallCallsCopy[floor] = make([]types.HallCall, len(om.hallCalls[floor]))
		copy(hallCallsCopy[floor], om.hallCalls[floor])
	}

	return hallCallsCopy
}

// CopySelfHallCalls returns a deep copy of this node's local hall-call contribution.
func (om *OrderManager) CopySelfHallCalls() [][]types.HallCall {
	out := make([][]types.HallCall, len(om.selfHallCalls))
	for floor := 0; floor < len(om.selfHallCalls); floor++ {
		out[floor] = make([]types.HallCall, len(om.selfHallCalls[floor]))
		copy(out[floor], om.selfHallCalls[floor])
	}
	return out
}

// CopySelfState returns the current local elevator state.
func (om *OrderManager) CopySelfState() types.ElevatorState {
	return om.selfState
}

// CreateNodeState returns the complete local node state for broadcasting.
func (om *OrderManager) CreateNodeState() types.NodeState {
	return types.NodeState{
		ElevatorState: om.CopySelfState(),
		HallCalls:     om.CopySelfHallCalls(),
	}
}

// UpdateSelfState stores the latest local elevator state and assigns a new sequence number.
func (om *OrderManager) UpdateSelfState(newState types.ElevatorState) {
	selfID := om.selfState.ID
	becameUnavailable := om.selfState.Available && !newState.Available

	newState.SequenceNumber = om.selfState.SequenceNumber + 1
	om.selfState = newState

	if becameUnavailable {
		om.reassignHallCallsFromUnavailableElevator(selfID)
	}
}

// HandleButtonEvent records a new local button event.
func (om *OrderManager) HandleButtonEvent(event types.ButtonEvent) {
	om.RegisterButtonPress(event)
}

// CopyCabCalls returns a copy of the local cab-call table.
func (om *OrderManager) CopyCabCalls() []bool {
	cabCallsCopy := make([]bool, len(om.cabCalls))
	copy(cabCallsCopy, om.cabCalls)
	return cabCallsCopy
}

// PrintSystemState prints the current known system state for debugging.
func (om *OrderManager) PrintSystemState() {

	fmt.Println("---------- SYSTEM STATE ----------")

	fmt.Println("Self Elevator:")
	fmt.Printf("  ID: %d\n", om.selfState.ID)
	fmt.Printf("  Floor: %d\n", om.selfState.Floor)
	fmt.Printf("  Direction: %v\n", om.selfState.Dir)
	fmt.Printf("  Behaviour: %v\n", om.selfState.Behaviour)
	fmt.Printf("  Available: %v\n", om.selfState.Available)

	fmt.Println("\nPeer Elevators:")

	if len(om.peerStates) == 0 {
		fmt.Println("  (none)")
	}

	for id, state := range om.peerStates {
		fmt.Printf("Elevator %d -> Floor: %d Direction: %v Behaviour: %v Available: %v\n",
			id, state.Floor, state.Dir, state.Behaviour, state.Available)
	}

	fmt.Println("\nCab Calls:")

	for floor := 0; floor < len(om.cabCalls); floor++ {
		if om.cabCalls[floor] {
			fmt.Printf("  Floor %d: ACTIVE\n", floor)
		}
	}

	fmt.Println("\nHall Calls:")

	for floor := 0; floor < len(om.hallCalls); floor++ {

		up := om.hallCalls[floor][types.HallUpIndex]
		down := om.hallCalls[floor][types.HallDownIndex]

		fmt.Printf(
			"  Floor %d -> Up: %v (rev %d, assigned %d)  Down: %v (rev %d, assigned %d)\n",
			floor,
			up.State,
			up.Revision,
			up.AssignedTo,
			down.State,
			down.Revision,
			down.AssignedTo,
		)
	}

	fmt.Println("----------------------------------")
}
