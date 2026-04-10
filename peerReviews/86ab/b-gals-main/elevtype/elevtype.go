package elevtype

type Dirn int

const (
	D_Down Dirn = -1
	D_Stop Dirn = 0
	D_Up   Dirn = 1
)

type Behaviour int

const (
	B_Idle     	Behaviour = 0
	B_Moving	Behaviour = 1
	B_DoorOpen	Behaviour = 2
)

type Button int

const (
	B_HallUp   	Button = 0
	B_HallDown 	Button = 1
	B_Cab		Button = 2
)

