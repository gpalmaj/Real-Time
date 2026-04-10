package distributor

import (
	"heis/config"
	"heis/driver/elevio"
	"heis/elevtype"
	"heis/order"
	"heis/worldview"
)

func updateButtonLamps(id string, worldView worldview.WorldView) {
	for f := 0; f < config.N_FLOORS; f++ {
		elevio.SetButtonLamp(elevio.BT_HallUp, f,
			worldView.HallOrders[f][int(elevtype.B_HallUp)].Status >= order.OrderStatusConfirmed)
		elevio.SetButtonLamp(elevio.BT_HallDown, f,
			worldView.HallOrders[f][int(elevtype.B_HallDown)].Status >= order.OrderStatusConfirmed)
	}
	if state, ok := worldView.States[id]; ok {
		for f := 0; f < config.N_FLOORS; f++ {
			elevio.SetButtonLamp(elevio.BT_Cab, f,
				state.CabOrders[f].Status >= order.OrderStatusConfirmed)
		}
	}
}
