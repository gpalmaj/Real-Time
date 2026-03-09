package network

import (
	"FinalProject_G92/config"
	"FinalProject_G92/models"
	"fmt"
	"time"
)

func MergeWorldview(local *models.Worldview, remote models.Worldview) {
	for i := range config.N {
		if local.HallCalls[i].UpSeq < remote.HallCalls[i].UpSeq {
			local.HallCalls[i].Up = remote.HallCalls[i].Up
			local.HallCalls[i].UpSeq = remote.HallCalls[i].UpSeq
		}
		if local.HallCalls[i].DownSeq < remote.HallCalls[i].DownSeq {
			local.HallCalls[i].Down = remote.HallCalls[i].Down
			local.HallCalls[i].DownSeq = remote.HallCalls[i].DownSeq
		}
	}
}

func UpdateCabCallLog(wv *models.Worldview, lobby map[int]models.Node) {
	newLog := make(map[int][config.N]bool, len(lobby))
	for key := range lobby {
		newLog[key] = lobby[key].Worldview.CabCalls
	}

	wv.CabCallLog = newLog
}

func ComputeHallLights(lobby map[int]models.Node) [config.N]models.HallCall {
	var lights [config.N]models.HallCall
	for i := range config.N {
		allUp := true
		for _, elevator := range lobby {
			if !elevator.Worldview.HallCalls[i].Up {
				allUp = false
				break
			}
		}
		lights[i].Up = allUp

		allDown := true
		for _, elevator := range lobby {
			if !elevator.Worldview.HallCalls[i].Down {
				allDown = false
				break
			}
		}
		lights[i].Down = allDown
	}
	return lights
}

func DetectDisconnections(lobby map[int]models.Node, timeout time.Duration) {
	for id, node := range lobby {
		if node.Alive && time.Since(node.Lastseen) > timeout {
			node.Alive = false
			lobby[id] = node
			fmt.Printf("Node %d disconnected\n", id)
		}
	}
}
