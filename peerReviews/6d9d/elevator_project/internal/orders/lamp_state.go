package orders

import "elevator/internal/types"

// ButtonLampState returns the desired lamp state derived from current order state.
func (om *OrderManager) ButtonLampState() types.LightState {
	ls := types.LightState{
		Cab:  make([]bool, len(om.cabCalls)),
		Hall: make([][]bool, len(om.cabCalls)),
	}

	for floor := 0; floor < len(om.cabCalls); floor++ {
		ls.Cab[floor] = om.cabCalls[floor]
		ls.Hall[floor] = make([]bool, 2)

		upState := om.hallCalls[floor][types.HallUpIndex].State
		downState := om.hallCalls[floor][types.HallDownIndex].State

		ls.Hall[floor][types.HallUpIndex] =
			upState == types.HallCallUnconfirmedAdd || upState == types.HallCallConfirmedAdd

		ls.Hall[floor][types.HallDownIndex] =
			downState == types.HallCallUnconfirmedAdd || downState == types.HallCallConfirmedAdd
	}

	return ls
}
