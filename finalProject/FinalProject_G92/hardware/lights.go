package hardware

import (
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/network"
)

func HallLights(lobby map[int]network.Node) {
	for {
		lights := network.UpdateLights(lobby) //should be able to call this everytime as long as WorldView has not removed unfinished calls
		for i := 0; i < network.N; i++ {
			if lights[i].Up {
				elevio.SetButtonLamp(elevio.BT_HallUp, i, true)
			} else {
				elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
			}
			if lights[i].Down {
				elevio.SetButtonLamp(elevio.BT_HallDown, i, true)
			} else {
				elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
			}

		}
	}
}

func CabLights() {

}
