package hardware

import (
	"FinalProject_G92/config"
	"FinalProject_G92/hardware/elevio"
	"FinalProject_G92/models"
)

func HallLights(lightsCh <-chan [config.N]models.HallCall) {
	for lights := range lightsCh {
		for i := range config.N {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, lights[i].Up)
			elevio.SetButtonLamp(elevio.BT_HallDown, i, lights[i].Down)
		}
	}
}
