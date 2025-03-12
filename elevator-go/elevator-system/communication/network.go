package communication

import (
	"Driver-go/elevator-system/elevatorStateMachine"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Constants for the network module
const (
	UDP_PORT        = "15657"
	BROADCAST_ADDR  = "255.255.255.255:15657"
	RETRANSMIT_RATE = 500 * time.Millisecond
	TIMEOUT_LIMIT   = 2 * time.Second
	HEARTBEAT_RATE  = 500 * time.Millisecond
)

// ElevatorState struct
type ElevatorState struct {
	floor     int                                                                 `json:"floor"`
	dirn      elevatorStateMachine.Dirn                                           `json:"dirn"`
	requests  [elevatorStateMachine.N_FLOORS][elevatorStateMachine.N_BUTTONS]bool `json:"requests"`
	active    bool
	behaviour elevatorStateMachine.ElevatorBehaviour

	// The ones above are the same as in the Elevator struct minus a few, the ones below is needed for the state
	// Probably a cleaner way to implement this, since we already have a similar struct

	ID          int       `json:"id"`
	IsMaster    bool      `json:"is_master"`
	LastUpdated time.Time `json:"-"`
	Heartbeat   time.Time `json:"-"`
}

var (
	PeerStatus   sync.Map // PeerStaus tracks last update time for each elevator
	stateUpdates = make(chan ElevatorState)
	udpConn      *net.UDPConn // UDP socket
	activePeers  []string
)

// PeerStatus as sync.Map instead??
// sateUpdates as a chan with the stateUpdates

// initNetwork initializes the UDP connection
func InitNetwork(elevatorID int, updateChannel chan ElevatorState) {
	// Use a fixed port for peer discovery (same for all elevators)
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	// Each elevator broadcasts its presence and listens for other peers
	go peers.Transmitter(15647, fmt.Sprintf("elevator-%d", elevatorID), peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// Unique port for each elevator broadcasting
	statePort := 20100 + elevatorID // Example: 20101, 20102, ...

	stateTx := make(chan ElevatorState)
	stateRx := make(chan ElevatorState)

	go bcast.Transmitter(statePort, stateTx) // Transmit state on unique port
	go bcast.Receiver(statePort, stateRx)    // Receive state on unique port

	// Track peer updates
	go func() {
		for {
			select {
			case peerUpdate := <-peerUpdateCh:
				activePeers = peerUpdate.Peers
				fmt.Println("Updated peer list:", activePeers)
			case receivedState := <-stateRx:
				// Process state updates from other elevators
				updateChannel <- receivedState
			}
		}
	}()

	// Start periodic state transmission
	go func() {
		for {
			time.Sleep(1 * time.Second)
			if state, exists := getPeerStatus(elevatorID); exists {
				stateTx <- state
			}
		}
	}()
}

// listenForUpdates recives UDP packets and updates PeerStatus
func listenForUpdates(updateChannel chan ElevatorState) {
	buffer := make([]byte, 1024)
	for {
		n, _, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reciving UDP packet", err)
			continue
		}

		var receivedState ElevatorState

		// check if state recived is valid
		err = json.Unmarshal(buffer[:n], &receivedState)
		if err != nil {
			fmt.Println("Error decoding json:", err)
			continue
		}

		receivedState.LastUpdated = time.Now()
		stateUpdates <- receivedState
	}
}

func processStateUpdates(updateChannel chan ElevatorState) {
	for state := range stateUpdates {
		PeerStatus.Store(state.ID, state)
		updateChannel <- state
	}
}

// retransmitState periodically sends the local elevator's state
func retransmitState(elevatorID int) {
	ticker := time.NewTicker(RETRANSMIT_RATE)
	defer ticker.Stop()

	for range ticker.C {
		if state, exists := getPeerStatus(elevatorID); exists {
			sendStateUpdate(state, activePeers) // Send only to known peers
		}
	}
}

// sendStateUpdate serializes and broadcasts the state
func sendStateUpdate(elevator ElevatorState, peersList []string) {
	if udpConn == nil {
		fmt.Println("udpConn is nil, can't send update")
		return
	}

	data, err := json.Marshal(elevator)
	if err != nil {
		fmt.Println("Error with JSON format:", err)
		return
	}

	// Send to each discovered peer
	for _, peer := range peersList {
		addr, _ := net.ResolveUDPAddr("udp", peer+":"+UDP_PORT)
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
			continue
		}
		defer conn.Close()

		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Error sending UDP packet to", peer, ":", err)
		}
	}
}

func sendHeartbeat(elevatorID int) {
	ticker := time.NewTicker(HEARTBEAT_RATE)
	defer ticker.Stop()

	for range ticker.C {
		if state, exist := getPeerStatus(elevatorID); exist {
			state.Heartbeat = time.Now()
			PeerStatus.Store(elevatorID, state)
		}
	}
}

// detectFailures identifies unresponsive elevators
func detectFailures() {
	for {
		time.Sleep(TIMEOUT_LIMIT)
		now := time.Now()
		PeerStatus.Range(func(key, value interface{}) bool {
			id := key.(int)
			state := value.(ElevatorState)
			if now.Sub(state.Heartbeat) > TIMEOUT_LIMIT {
				fmt.Printf("Elevator %d is unresponsive\n", id)
				PeerStatus.Delete(id)
			}
			return true
		})
	}
}

func getPeerStatus(id int) (ElevatorState, bool) {
	val, ok := PeerStatus.Load(id)
	if ok {
		return val.(ElevatorState), true
	}
	return ElevatorState{}, false
}
