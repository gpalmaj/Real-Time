package elevatorcontrol

import (
	"elevator/internal/elevio"
	"elevator/internal/types"
)

func toMotorDirection(dir types.Direction) elevio.MotorDirection {
	switch dir {
	case types.DirectionUp:
		return elevio.MD_Up
	case types.DirectionDown:
		return elevio.MD_Down
	default:
		return elevio.MD_Stop
	}
}

func toButtonType(button elevio.ButtonType) types.ButtonType {
	switch button {
	case elevio.BT_HallUp:
		return types.ButtonHallUp
	case elevio.BT_HallDown:
		return types.ButtonHallDown
	default:
		return types.ButtonCab
	}
}

func toButtonEvent(event elevio.ButtonEvent) types.ButtonEvent {
	return types.ButtonEvent{
		Floor:  event.Floor,
		Button: toButtonType(event.Button),
	}
}
