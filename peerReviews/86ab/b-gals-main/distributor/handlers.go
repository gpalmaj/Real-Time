package distributor

import (
	"heis/driver/elevio"
	"heis/order"
	"heis/saveIfCrash"
	"heis/worldview"
)

func handleButtonPress(id string, worldView worldview.WorldView, buttonEvent elevio.ButtonEvent, peersAlive []string) worldview.WorldView {
	alone := isSoleElevator(peersAlive)

	switch buttonEvent.Button {
	case elevio.BT_Cab:
		state := worldView.States[id]
		if state.CabOrders[buttonEvent.Floor].Status == order.OrderStatusNone ||
			state.CabOrders[buttonEvent.Floor].Status == order.OrderStatusUnknown {
			state.CabOrders[buttonEvent.Floor].Status = order.OrderStatusConfirmed
			state.CabOrders[buttonEvent.Floor].Barrier = []string{}
			worldView.States[id] = state
			saveIfCrash.SaveCabOrders(id, state.CabOrders)
		}
	default:
		b := int(buttonEvent.Button)
		if worldView.HallOrders[buttonEvent.Floor][b].Status == order.OrderStatusNone ||
			worldView.HallOrders[buttonEvent.Floor][b].Status == order.OrderStatusUnknown {
			worldView.HallOrders[buttonEvent.Floor][b].Status = order.OrderStatusUnconfirmed
			worldView.HallOrders[buttonEvent.Floor][b].Barrier = []string{id}
			if alone {
				worldView.HallOrders[buttonEvent.Floor][b].Status = order.OrderStatusConfirmed
				worldView.HallOrders[buttonEvent.Floor][b].Barrier = []string{}
			}
		}
	}
	return worldView
}

func handleFinishedOrder(id string, worldView worldview.WorldView, buttonEvent elevio.ButtonEvent, peersAlive []string) worldview.WorldView {
	alone := isSoleElevator(peersAlive)

	switch buttonEvent.Button {
	case elevio.BT_Cab:
		state := worldView.States[id]
		state.CabOrders[buttonEvent.Floor].Status = order.OrderStatusNone
		state.CabOrders[buttonEvent.Floor].Barrier = []string{}
		worldView.States[id] = state
		saveIfCrash.SaveCabOrders(id, state.CabOrders)
	default:
		b := int(buttonEvent.Button)
		worldView.HallOrders[buttonEvent.Floor][b].Status = order.OrderStatusFinished
		worldView.HallOrders[buttonEvent.Floor][b].Barrier = []string{id}
		if alone {
			worldView.HallOrders[buttonEvent.Floor][b].Status = order.OrderStatusNone
			worldView.HallOrders[buttonEvent.Floor][b].Barrier = []string{}
		}
	}
	return worldView
}

func isSoleElevator(peersAlive []string) bool {
	const minimumPeersForBarrier = 2
	return len(peersAlive) < minimumPeersForBarrier
}
