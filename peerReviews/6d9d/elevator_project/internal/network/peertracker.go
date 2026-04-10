package network

import (
	"sort"
	"time"

	"elevator/internal/types"
)

const NoNewPeer = -1

// Struct sent on PeerUpdates channel every time a peer changes
type PeerUpdate struct {
	Peers []int // active peers
	New   int   // new peer that was added, or NoNewPeer if no new peer
	Lost  []int // peers that were lost
}

// Struct for tracking peers. Has a map containing the last time we saw each peer, and a timeout duration
type PeerTracker struct {
	Timeout  time.Duration
	lastSeen map[int]time.Time
}

// Creates new PeerTracker with the specified timeout duration, and an empty lastSeen map
func NewPeerTracker(timeoutDuration time.Duration) *PeerTracker {
	return &PeerTracker{
		Timeout:  timeoutDuration,
		lastSeen: make(map[int]time.Time),
	}
}

// Reads incoming NodeStates. For each NodeState, updates the last seen time for that peer.
// If it's a new peer, sends an update on the PeerUpdates channel
// Peridically checks for lost peers and sends updates on the PeerUpdates channel if any peers are lost
func (peerTracker *PeerTracker) TrackPeers(
	incomingStates <-chan types.NodeState,
	peerUpdates chan<- PeerUpdate,
	shutdownSignal <-chan struct{},
) {
	ticker := time.NewTicker(peerTracker.Timeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-shutdownSignal:
			return

		case incomingNodeState, channelOpen := <-incomingStates:
			if !channelOpen {
				return
			}

			peerID := incomingNodeState.ElevatorState.ID

			_, peerAlreadyKnown := peerTracker.lastSeen[peerID]
			peerTracker.lastSeen[peerID] = time.Now()

			if !peerAlreadyKnown {
				select {
				case peerUpdates <- PeerUpdate{
					Peers: peerTracker.sortedPeers(),
					New:   peerID,
					Lost:  nil,
				}:
				case <-shutdownSignal:
					return
				}
			}

		case <-ticker.C:
			lostPeerIDs := peerTracker.collectLost()
			if len(lostPeerIDs) > 0 {
				sort.Ints(lostPeerIDs)
				select {
				case peerUpdates <- PeerUpdate{
					Peers: peerTracker.sortedPeers(),
					New:   NoNewPeer,
					Lost:  lostPeerIDs,
				}:
				case <-shutdownSignal:
					return
				}
			}
		}
	}
}

// Checks the last seen times for all peers and returns the IDs of any peers that have not been seen within the timeout duration.
// Also removes lost peers from the lastSeen map.
func (peerTracker *PeerTracker) collectLost() []int {
	lostPeerIDs := make([]int, 0)
	currentTime := time.Now()

	for peerID, lastSeenAt := range peerTracker.lastSeen {
		if currentTime.Sub(lastSeenAt) > peerTracker.Timeout {
			lostPeerIDs = append(lostPeerIDs, peerID)
			delete(peerTracker.lastSeen, peerID)
		}
	}

	return lostPeerIDs
}

// Sorts peers by ID, makes it cleaner
func (peerTracker *PeerTracker) sortedPeers() []int {
	activePeerIDs := make([]int, 0, len(peerTracker.lastSeen))
	for peerID := range peerTracker.lastSeen {
		activePeerIDs = append(activePeerIDs, peerID)
	}
	sort.Ints(activePeerIDs)
	return activePeerIDs
}
