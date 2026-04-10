package controller

import (
	"FinalProject_G92/config"
	"FinalProject_G92/controller/elevio"
	"FinalProject_G92/models"
)

// HallLights receives consensus hall call state from the coordinator
// and updates the physical hall button lamps accordingly.
func HallLights(lightsCh <-chan [config.N]models.HallCall) {
	for lights := range lightsCh {
		for i := range config.N {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, lights[i].Up)
			elevio.SetButtonLamp(elevio.BT_HallDown, i, lights[i].Down)
		}
	}
}
