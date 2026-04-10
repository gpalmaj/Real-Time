package elevatorcontrol

import (
	"elevator/internal/elevio"
	"elevator/internal/types"
)

func ApplyButtonLamps(ls types.LightState) {
	for floor := 0; floor < len(ls.Cab); floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, ls.Cab[floor])

		if floor < len(ls.Cab)-1 {
			elevio.SetButtonLamp(elevio.BT_HallUp, floor, ls.Hall[floor][types.HallUpIndex])
		}
		if floor > 0 {
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, ls.Hall[floor][types.HallDownIndex])
		}
	}
}
