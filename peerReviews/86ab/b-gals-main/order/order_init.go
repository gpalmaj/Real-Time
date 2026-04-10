package order

func NewOrder() Order {
	return Order{
		Status:  OrderStatusUnknown,
		Barrier: make([]string, 0),
	}
}

func NewHallOrders(numFloors int) [][2]Order { 
	orders := make([][2]Order, numFloors)

	for floor := range orders {
		orders[floor][0] = NewOrder() //HallUp = 0
		orders[floor][1] = NewOrder() //HallDown = 1
	}
	return orders
}

func NewCabOrders(numFloors int) []Order {
	orders := make([]Order, numFloors)
	for floor := range orders {
		orders[floor] = NewOrder()
	}
	return orders
}
