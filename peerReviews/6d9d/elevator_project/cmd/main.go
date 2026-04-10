package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"elevator/internal/config"
	"elevator/internal/elevatorcontrol"
	"elevator/internal/elevio"
	"elevator/internal/faulttolerance"
	"elevator/internal/network"
	"elevator/internal/orders"
	"elevator/internal/types"
)

func main() {
	elevatorID := flag.Int("id", -1, "unique elevator ID")
	flag.Parse()

	if *elevatorID < 0 {
		log.Fatal("missing required --id <number>")
	}

	faulttolerance.RunProcessPair(*elevatorID)
	elevio.Init("localhost:15657", config.NumberOfFloors)

	for f := 0; f < config.NumberOfFloors; f++ {
		elevio.SetButtonLamp(elevio.BT_Cab, f, false)

		if f < config.NumberOfFloors-1 {
			elevio.SetButtonLamp(elevio.BT_HallUp, f, false)
		}
		if f > 0 {
			elevio.SetButtonLamp(elevio.BT_HallDown, f, false)
		}
	}

	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)

	actionCh := make(chan types.Action, 16)
	buttonEventCh := make(chan types.ButtonEvent, 16)
	stateCh := make(chan types.ElevatorState, 16)
	servedFloorCh := make(chan int, 16)

	orderManager := orders.NewOrderManager(config.NumberOfFloors, *elevatorID)

	netw, err := network.Init(29999, *elevatorID)
	if err != nil {
		log.Fatal(err)
	}
	defer netw.Close()

	go elevatorcontrol.Run(
		*elevatorID,
		actionCh,
		buttonEventCh,
		stateCh,
		servedFloorCh,
	)

	fmt.Printf("Elevator node %d started\n", *elevatorID)
	fmt.Printf("Floors: %d\n", config.NumberOfFloors)

	applyLamps := func() {
		elevatorcontrol.ApplyButtonLamps(orderManager.ButtonLampState())
	}

	sendAction := func() {
		action := orderManager.ComputeNextAction()
		select {
		case actionCh <- action:
		default:
		}
	}

	refreshOutputs := func() {
		applyLamps()
		sendAction()
	}

	applyLamps()

	heartbeatTicker := time.NewTicker(100 * time.Millisecond)
	defer heartbeatTicker.Stop()

	for {
		select {
		case btn := <-buttonEventCh:
			fmt.Printf("Button event: %+v\n", btn)
			orderManager.HandleButtonEvent(btn)
			refreshOutputs()

		case state := <-stateCh:
			orderManager.UpdateSelfState(state)
			sendAction()

		case servedFloor := <-servedFloorCh:
			fmt.Printf("Served floor: %d\n", servedFloor)

			clearAction := orderManager.CreateClearRequestAction()
			orderManager.ApplyClearedRequests(clearAction)
			refreshOutputs()

		case nodeState := <-netw.Rx:
			if orderManager.HandleNodeStateUpdate(nodeState) {
				fmt.Printf("Received state from peer %d\n", nodeState.ElevatorState.ID)
				orderManager.PrintSystemState()
				refreshOutputs()
			}

		case peerUpdate := <-netw.PeerUpdates:
			if peerUpdate.New != network.NoNewPeer {
				fmt.Printf("New peer: %d\n", peerUpdate.New)
			}

			for _, lostPeerID := range peerUpdate.Lost {
				fmt.Printf("Lost peer: %d\n", lostPeerID)
				orderManager.RemovePeer(lostPeerID)
			}

			refreshOutputs()

		case <-heartbeatTicker.C:
			select {
			case netw.Tx <- orderManager.CreateNodeState():
			default:
			}
		}
	}
}
