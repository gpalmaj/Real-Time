package order

type OrderStatus int

const (
	OrderStatusUnknown       OrderStatus = -1 
	OrderStatusNone 		 OrderStatus = 0  
	OrderStatusUnconfirmed   OrderStatus = 1  
	OrderStatusConfirmed     OrderStatus = 2 
	OrderStatusFinished      OrderStatus = 3  
)

type Order struct {
	Status  OrderStatus 
	Barrier []string   

}
