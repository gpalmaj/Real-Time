package network

import (
	"FinalProject_G92/config"
	"FinalProject_G92/types"
	"fmt"
	"time"
)

func MergeWorldview(local *types.Worldview, remote types.Worldview) {
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

func UpdateCabCallLog(wv *types.Worldview, lobby map[int]types.Node) {
	for key := range lobby {
		wv.CabCallLog[key] = lobby[key].Worldview.CabCalls
	}
}

func ComputeHallLights(lobby map[int]types.Node) [config.N]types.HallCall {
	var lights [config.N]types.HallCall
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

func DetectDisconnections(lobby map[int]types.Node, timeout time.Duration) {
	for id, node := range lobby {
		if node.Alive && time.Since(node.Lastseen) > timeout {
			node.Alive = false
			lobby[id] = node
			fmt.Printf("Node %d disconnected\n", id)
		}
	}
}
