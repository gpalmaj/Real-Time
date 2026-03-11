package network

import (
	"FinalProject_G92/config"
	"FinalProject_G92/debug"
	"FinalProject_G92/models"
	"time"
)

func NetworkManager(myId int, worldviewCh chan models.Worldview, heartbeatCh chan models.Heartbeat, assignCh, newOrder, removeOrder chan models.Order, lightsCh chan<- [config.N]models.HallCall, statusCh chan models.StatusMessage) {

	var wv models.Worldview
	lobby := make(map[int]models.Node)
	wv.CabCallLog = make(map[int][config.N]bool)

	booted := false

	disconnectTicker := time.NewTicker(1 * time.Second)
	defer disconnectTicker.Stop()

	for {
		select {
		case hb := <-heartbeatCh:
			node := lobby[hb.ID]
			node.Worldview.HallCalls = hb.Worldview.HallCalls
			node.Worldview.Status = hb.Worldview.Status
			if !booted {
				if myCabCalls, ok := hb.Worldview.CabCallLog[myId]; ok {
					wv.CabCalls = myCabCalls
					booted = true
					for f, active := range myCabCalls {
						if active {
							assignCh <- models.Order{Floor: f, Cab: true}
						}
					}
				}
			}

			node.Worldview.CabCalls = hb.Worldview.CabCalls
			node.Lastseen = time.Now()
			node.Alive = true
			lobby[hb.ID] = node

			MergeWorldview(&wv, hb.Worldview)
			UpdateCabCallLog(&wv, lobby)
			consensusCalls := ComputeHallLights(lobby)
			worldviewCh <- wv
			select {
			case lightsCh <- consensusCalls:
			default:
			}

			//assigning orders
			for _, order := range Assign(myId, consensusCalls, lobby) {
				assignCh <- order
			}

			debug.PrintLobby(lobby)

		case no := <-newOrder:
			if no.Cab {
				wv.CabCalls[no.Floor] = true
			} else if no.Dir {
				wv.HallCalls[no.Floor].Up = true
				wv.HallCalls[no.Floor].UpSeq++
			} else {
				wv.HallCalls[no.Floor].Down = true
				wv.HallCalls[no.Floor].DownSeq++
			}

		case ro := <-removeOrder:
			if ro.Cab {
				wv.CabCalls[ro.Floor] = false
			} else if ro.Dir {
				wv.HallCalls[ro.Floor].Up = false
				wv.HallCalls[ro.Floor].UpSeq++
			} else {
				wv.HallCalls[ro.Floor].Down = false
				wv.HallCalls[ro.Floor].DownSeq++
			}

		case sm := <-statusCh:
			wv.Status = sm
		case <-disconnectTicker.C:
			DetectDisconnections(lobby, config.DisconnectTimeout)
		}
	}
}
