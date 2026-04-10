package distributor

import (
	"fmt"
	"heis/config"
	"heis/driver/elevio"
	"heis/network/bcast"
	"heis/network/peers"
	"heis/order"
	"heis/saveIfCrash"
	"heis/worldview"
	"time"
)

const (
	broadcastInterval  = 50 * time.Millisecond
	lampUpdateInterval = 200 * time.Millisecond
)

type networkMessage struct {
	SenderID  string
	WorldView worldview.WorldView
}

func Distributor(
	id string,
	stateCh <-chan worldview.ElevatorState,
	finReqCh <-chan elevio.ButtonEvent,
	assignmentCh chan [config.N_FLOORS][config.N_BUTTONS]bool,
	peerUpdateCh <-chan peers.PeerUpdate,
) {
	localWorldView := worldview.InitWorldView(id, config.N_FLOORS)
	savedCabOrders := saveIfCrash.LoadCabOrders(id, config.N_FLOORS)
	for f := range savedCabOrders {
		if savedCabOrders[f].Status >= order.OrderStatusUnconfirmed {
			savedCabOrders[f].Status = order.OrderStatusConfirmed
			savedCabOrders[f].Barrier = []string{}
		}
	}
	restoredState := localWorldView.States[id]
	restoredState.CabOrders = savedCabOrders
	localWorldView.States[id] = restoredState

	peersAlive := []string{id}
	lastSeen := make(map[string]bool)
	lastSeen[id] = true

	messageTx := make(chan networkMessage, 1)
	messageRx := make(chan networkMessage)
	go bcast.Transmitter(config.PortDistributor, messageTx)
	go bcast.Receiver(config.PortDistributor, messageRx)

	driverButtonCh := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(driverButtonCh)

	broadcastTicker := time.NewTicker(broadcastInterval)
	defer broadcastTicker.Stop()
	lampTicker := time.NewTicker(lampUpdateInterval)
	defer lampTicker.Stop()

	for {
		select {
		case <-broadcastTicker.C:
			messageTx <- networkMessage{
				SenderID:  id,
				WorldView: copyWorldView(localWorldView),
			}

		case message := <-messageRx:
			if message.SenderID == id {
				continue
			}
			lastSeen[message.SenderID] = true

			hallBefore := copyHallOrders(localWorldView.HallOrders)

			localWorldView.HallOrders = mergeHallOrders(
				id,
				localWorldView.HallOrders,
				message.WorldView.HallOrders,
				peersAlive,
			)
			localWorldView.States = mergeStates(
				id,
				localWorldView.States,
				message.SenderID,
				message.WorldView.States,
				peersAlive,
			)

			if hallOrdersStatusChanged(hallBefore, localWorldView.HallOrders) {
				for _, peer := range peersAlive {
					if peer != id {
						lastSeen[peer] = false
					}
				}
			}

			if allPeersSeen(peersAlive, lastSeen) {
				computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)
			}

		case buttonEvent := <-driverButtonCh:
			localWorldView = handleButtonPress(id, localWorldView, buttonEvent, peersAlive)
			if allPeersSeen(peersAlive, lastSeen) {
				computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)
			}

		case completedOrder := <-finReqCh:
			localWorldView = handleFinishedOrder(id, localWorldView, completedOrder, peersAlive)
			localWorldView.HallOrders = mergeHallOrders(
				id,
				localWorldView.HallOrders,
				copyHallOrders(localWorldView.HallOrders),
				peersAlive,
			)
			if allPeersSeen(peersAlive, lastSeen) {
				computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)
			}

		case state := <-stateCh:
			elevatorState := localWorldView.States[id]
			elevatorState.Floor = state.Floor
			elevatorState.Direction = state.Direction
			elevatorState.Behaviour = state.Behaviour
			localWorldView.States[id] = elevatorState

		case peerUpdate := <-peerUpdateCh:
			peersAlive = peerUpdate.Peers
			if !order.ContainsID(peersAlive, id) {
				peersAlive = append(peersAlive, id)
			}
			// Reset lastSeen — only keep peers that are still alive
			newLastSeen := make(map[string]bool)
			for _, peer := range peersAlive {
				if lastSeen[peer] {
					newLastSeen[peer] = true
				}
			}
			lastSeen = newLastSeen
			lastSeen[id] = true

			fmt.Println("Peers alive:", peersAlive)
			if allPeersSeen(peersAlive, lastSeen) {
				computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)
			}

		case <-lampTicker.C:
			updateButtonLamps(id, localWorldView)
		}
	}
}

func allPeersSeen(peersAlive []string, lastSeen map[string]bool) bool {
	for _, peer := range peersAlive {
		if !lastSeen[peer] {
			return false
		}
	}
	return true
}
