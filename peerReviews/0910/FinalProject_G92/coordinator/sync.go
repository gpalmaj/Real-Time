package coordinator

import (
	"FinalProject_G92/config"
	"FinalProject_G92/models"
	"fmt"
	"time"
)

// MergeWorldview adopts the remote's hall call state for any floor where
// the remote has a higher sequence number. This is how nodes converge.
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

// UpdateCabCallLog snapshots every node's cab calls into the local worldview.
// This log is broadcast in heartbeats so a rebooting node can recover its cab calls.
func UpdateCabCallLog(wv *models.Worldview, lobby map[int]models.Node) {
	newLog := make(map[int][config.N]bool, len(lobby))
	for key := range lobby {
		newLog[key] = lobby[key].Worldview.CabCalls
	}

	wv.CabCallLog = newLog
}

// ComputeHallLights determines which hall lights should be on.
// A light is on only if all alive nodes agree the call exists (consensus).
func ComputeHallLights(lobby map[int]models.Node) [config.N]models.HallCall {
	var lights [config.N]models.HallCall
	for i := range config.N {
		allUp := true
		for _, elevator := range lobby {
			if !elevator.Alive {
				continue
			}
			if !elevator.Worldview.HallCalls[i].Up {
				allUp = false
				break
			}
		}
		lights[i].Up = allUp

		allDown := true
		for _, elevator := range lobby {
			if !elevator.Alive {
				continue
			}
			if !elevator.Worldview.HallCalls[i].Down {
				allDown = false
				break
			}
		}
		lights[i].Down = allDown
	}
	return lights
}

// DetectDisconnections marks nodes as dead if no heartbeat has been received within timeout.
func DetectDisconnections(lobby map[int]models.Node, timeout time.Duration) {
	for id, node := range lobby {
		if node.Alive && time.Since(node.Lastseen) > timeout {
			node.Alive = false
			lobby[id] = node
			fmt.Printf("Node %d disconnected\n", id)
		}
	}
}
