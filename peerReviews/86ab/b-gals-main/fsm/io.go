package fsm

import (
	"heis/config"
	"heis/elevtype"
)


type ElevatorIO interface {
	SetMotorDirection(dir elevtype.Dirn)
	SetDoorOpenLamp(on bool)
	SetFloorIndicator(floor int)
	SetButtonLamp(floor int, buttonEvent elevtype.Button, on bool)
}


type IO struct {
	drv Driver
}

func NewIO(drv Driver) *IO {
	return &IO{drv: drv}
}


func (io *IO) SetMotorDirection(dir elevtype.Dirn) {
	if io.drv == nil {
		return
	}
	io.drv.SetMotorDirection(int(dir))
}

func (io *IO) SetDoorOpenLamp(on bool) {
	if io.drv == nil {
		return
	}
	io.drv.SetDoorOpenLamp(on)
}

func (io *IO) SetFloorIndicator(floor int) {
	if io.drv == nil {
		return
	}
	if floor < 0 || floor >= config.N_FLOORS {
		return
	}
	io.drv.SetFloorIndicator(floor)
}

func (io *IO) SetButtonLamp(floor int, buttonEvent elevtype.Button, on bool) {
	if io.drv == nil {
		return
	}
	if floor < 0 || floor >= config.N_FLOORS {
		return
	}
	io.drv.SetButtonLamp(int(buttonEvent), floor, on)
}







