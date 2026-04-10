package network

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"elevator/internal/config"
	"elevator/internal/types"
)

const (
	readDeadline = 200 * time.Millisecond
	rxBufferSize = 8192
)

type Network struct {
	// Interface to the rest of the system:
	Tx          chan types.NodeState // system -> network
	Rx          chan types.NodeState // network -> system
	PeerUpdates chan PeerUpdate

	// internal:
	peerTrackerInput chan types.NodeState
	shutdownSignal   chan struct{}

	senderID           int
	trackSelf          bool // if true, track the sender itself for testing purposes
	senderConnection   *net.UDPConn
	receiverConnection *net.UDPConn
}

// Init begins UDP broadcast sender, and receiver, and a peer tracker
// Initializes the network struct, used for communication within the module and the channels for communication with the rest of the system
// Exposes the channels netw.Tx, netw.Rx and netw.PeerUpdates
func Init(broadcastPort int, senderID int) (*Network, error) {
	// Receiver: listens on braodcast port
	receiverAddress := &net.UDPAddr{IP: net.IPv4zero, Port: broadcastPort}
	receiverConn, err := net.ListenUDP("udp4", receiverAddress)
	if err != nil {
		return nil, fmt.Errorf("ListenUDP: %w", err)
	}

	// Sender: set up broadcast at 255.255.255.255:broadcastPort
	broadcastAddress := &net.UDPAddr{IP: net.IPv4bcast, Port: broadcastPort}
	senderConn, err := net.DialUDP("udp4", nil, broadcastAddress)
	if err != nil {
		_ = receiverConn.Close()
		return nil, fmt.Errorf("DialUDP: %w", err)
	}

	networkStruct := &Network{
		Tx:          make(chan types.NodeState, 64),
		Rx:          make(chan types.NodeState, 64),
		PeerUpdates: make(chan PeerUpdate, 32),

		peerTrackerInput: make(chan types.NodeState, 64),
		shutdownSignal:   make(chan struct{}),

		senderID:  senderID,
		trackSelf: false,

		senderConnection:   senderConn,
		receiverConnection: receiverConn,
	}

	go networkStruct.senderLoop()
	go networkStruct.receiverLoop()

	// Starts peer tracker based on received snapshots
	peerTracker := NewPeerTracker(config.PeerTimeout)
	go peerTracker.TrackPeers(networkStruct.peerTrackerInput, networkStruct.PeerUpdates, networkStruct.shutdownSignal)

	return networkStruct, nil
}

// Sender loop takes in a node state from the Tx channel, packs it into JSON and sends it over UDP broadcast
func (networkStruct *Network) senderLoop() {
	defer func() { _ = networkStruct.senderConnection.Close() }()

	for {
		select {
		case <-networkStruct.shutdownSignal:
			return

		case nodeState, channelOpen := <-networkStruct.Tx:
			if !channelOpen {
				return
			}

			nodeState.ElevatorState.ID = networkStruct.senderID // Ensure the correct sender ID is set
			JSONbytes, err := json.Marshal(nodeState)
			if err != nil {
				fmt.Println("JSON.Marshal error:", err)
				continue
			}
			_, err = networkStruct.senderConnection.Write(JSONbytes)
			if err != nil {
				fmt.Println("UDP write error", err)
				continue
			}
		}
	}
}

// Receiver loop creates a buffer and reads incoming UDP packets into it. Unpacks it from JSON into a NodeState
// Sends the NodeState to both peerTracker and the Rx channel for the rest of the system
func (networkStruct *Network) receiverLoop() {
	defer func() { _ = networkStruct.receiverConnection.Close() }()

	buffer := make([]byte, rxBufferSize)

	for {
		_ = networkStruct.receiverConnection.SetReadDeadline(time.Now().Add(readDeadline))

		bytesRead, address, err := networkStruct.receiverConnection.ReadFromUDP(buffer)
		if err != nil {
			if networkError, ok := err.(net.Error); ok && networkError.Timeout() {
				continue
			}

			select {
			case <-networkStruct.shutdownSignal:
				return
			default:
			}

			fmt.Println("UDP read error:", err)
			continue
		}

		var incomingNodeState types.NodeState
		err = json.Unmarshal(buffer[:bytesRead], &incomingNodeState)
		if err != nil {
			fmt.Printf("JSON.Unmarshal error from %v: %v\n", address, err)
			continue
		}

		// If trackSelf is false, ignore messages from self
		if !networkStruct.trackSelf && incomingNodeState.ElevatorState.ID == networkStruct.senderID {
			continue
		}

		select {
		case networkStruct.peerTrackerInput <- incomingNodeState:
		case <-networkStruct.shutdownSignal:
			return
		}

		select {
		case networkStruct.Rx <- incomingNodeState:
		case <-networkStruct.shutdownSignal:
			return
		}
	}
}

// Signals all goroutines to stop and closes the UDP connections
func (networkStruct *Network) Close() {
	select {
	case <-networkStruct.shutdownSignal:
		return
	default:
		close(networkStruct.shutdownSignal)
	}

	_ = networkStruct.receiverConnection.Close()
	_ = networkStruct.senderConnection.Close()
}
