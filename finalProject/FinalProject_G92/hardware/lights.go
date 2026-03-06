package hardware

import (
	"FinalProject_G92/config"
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/types"
)

func HallLights(lightsCh <-chan [config.N]types.HallCall) {
	for lights := range lightsCh {
		for i := 0; i < config.N; i++ {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, lights[i].Up)
			elevio.SetButtonLamp(elevio.BT_HallDown, i, lights[i].Down)
		}
	}
}
