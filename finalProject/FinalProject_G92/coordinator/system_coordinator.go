package coordinator

import (
	"FinalProject_G92/config"
	"FinalProject_G92/debug"
	"FinalProject_G92/models"
	"time"
)

// SystemCoordinator is the central coordination loop. It merges peer worldviews,
// recovers cab calls on boot, computes hall light consensus, and assigns orders
// to the local elevator via the cost function. Runs on the main goroutine.
func SystemCoordinator(myId int, worldviewCh chan models.Worldview, heartbeatCh chan models.Heartbeat, assignCh, newOrder, removeOrder chan models.Order, lightsCh chan<- [config.N]models.HallCall, statusCh chan models.StatusMessage) {

	var wv models.Worldview
	lobby := make(map[int]models.Node)
	wv.CabCallLog = make(map[int][config.N]bool)

	// booted tracks whether cab calls have been recovered from a peer after a restart
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

		case no := <-newOrder: // add call to local worldview, seq bump propagates via next heartbeat
			if no.Cab {
				wv.CabCalls[no.Floor] = true
			} else if no.Dir {
				wv.HallCalls[no.Floor].Up = true
				wv.HallCalls[no.Floor].UpSeq++
			} else {
				wv.HallCalls[no.Floor].Down = true
				wv.HallCalls[no.Floor].DownSeq++
			}

		case removedOrder := <-removeOrder: // clear call from local worldview, seq bump propagates removal
			if removedOrder.Cab {
				wv.CabCalls[removedOrder.Floor] = false
			} else if removedOrder.Dir {
				wv.HallCalls[removedOrder.Floor].Up = false
				wv.HallCalls[removedOrder.Floor].UpSeq++
			} else {
				wv.HallCalls[removedOrder.Floor].Down = false
				wv.HallCalls[removedOrder.Floor].DownSeq++
			}

		case statusMessage := <-statusCh:
			wv.Status = statusMessage
		case <-disconnectTicker.C:
			DetectDisconnections(lobby, config.DisconnectTimeout)
		}
	}
}
