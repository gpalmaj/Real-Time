package worldview

import (
	"heis/elevtype"
	"heis/order"
)

type ElevatorState struct {
	Floor     int
	Direction elevtype.Dirn
	Behaviour elevtype.Behaviour
	CabOrders []order.Order
}

type WorldView struct {
	HallOrders [][2]order.Order
	States     map[string]ElevatorState
}

func InitWorldView(id string, numFloors int) WorldView {
	worldView := WorldView{
		HallOrders: order.NewHallOrders(numFloors),
		States:     make(map[string]ElevatorState),
	}
	worldView.States[id] = ElevatorState{
		Floor:     config.FloorUnknown,
		Direction: elevtype.D_Stop,
		Behaviour: elevtype.B_Idle,
		CabOrders: order.NewCabOrders(numFloors),
	}
	return worldView
}
