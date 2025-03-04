package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Constants for the network module
const (
	UDP_PORT        = "30000"
	BROADCAST_ADDR  = "255.255.255.255:30000"
	RETRANSMIT_RATE = 500 * time.Millisecond
	TIMEOUT_LIMIT   = 2 * time.Second
	HEARTBEAT_RATE  = 500 * time.Millisecond
)

// ElevatorState struct
type ElevatorState struct {
	Floor     int                       `json:"floor"`
	Dirn      Dirn                      `json:"dirn"`
	Requests  [N_FLOORS][N_BUTTONS]bool `json:"requests"`
	active    bool
	behaviour ElevatorBehaviour

	// The ones above are the same as in the Elevator struct minus a few, the ones below is needed for the state
	// Probably a cleaner way to implement this, since we already have a similar struct

	ID          int       `json:"id"`
	IsMaster    bool      `json:"is_master"`
	LastUpdated time.Time `json:"-"`
	Heartbeat   time.Time `json:"-"`
}

var (
	PeerStatus = make(map[int]ElevatorState) // PeerStaus tracks last update time for each elevator
	udpConn    *net.UDPConn                  // UDP socket
)

// PeerStatus as sync.Map instead??
// sateUpdates as a chan with the stateUpdates

// initNetwork initializes the UDP connection
func initNetwork(elevatorID int, updateChannel chan ElevatorState) {
	// Use a unique port for each elevator based on its ID
	localPort := fmt.Sprintf("30%03d", elevatorID) // e.g., 30001, 30002, 30003

	addr, err := net.ResolveUDPAddr("udp", ":"+localPort)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}
	udpConn = conn

	fmt.Println("Elevator", elevatorID, "listening on UDP port", localPort)

	go listenForUpdates(updateChannel)            // Listening for messages
	go retransmitState(elevatorID, updateChannel) // Periodic retransmission of state updates
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

		// Updated PeerStatus map
		receivedState.LastUpdated = time.Now()
		PeerStatus[receivedState.ID] = receivedState

		// Send update to FSM
		updateChannel <- receivedState
	}
}

// retransmitState periodically sends the local elevator's state
func retransmitState(elevatorID int, updateChannel chan ElevatorState) {
	ticker := time.NewTicker(RETRANSMIT_RATE)
	defer ticker.Stop()

	for range ticker.C {
		for peerID, state := range PeerStatus {
			if peerID != elevatorID { // Don't send to itself
				peerPort := fmt.Sprintf("30%03d", peerID) // Get peer's listening port (e.g., 30002)
				addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+peerPort)
				if err != nil {
					fmt.Println("Error resolving UDP address:", err)
					continue
				}

				conn, err := net.DialUDP("udp", nil, addr) // Correctly set remote address
				if err != nil {
					fmt.Println("Error creating UDP connection:", err)
					continue
				}
				defer conn.Close()

				sendStateUpdate(state, BROADCAST_ADDR) // Now properly sends data
			}
		}
	}
}

// sendStateUpdate serializes and broadcasts the state
func sendStateUpdate(elevator ElevatorState, addr *net.UDPAddr) {
	conn, err := net.Dial("udp", BROADCAST_ADDR)
	if err != nil {
		fmt.Println("Error connecting to broadcast:", err)
		return
	}
	defer conn.Close()

	data, err := json.Marshal(elevator)
	if err != nil {
		fmt.Println("Error with JSON format:", err)
		return
	}

	_, err = udpConn.WriteToUDP(data, addr)
	if err != nil {
		fmt.Println("Error sending UDP packet:", err)
	}
}

// DetectFailures identifies unresponsive elevators
func (em *ElevatorManager) DetectFailures() {
	for id, elevator := range em.Elevators {
		if time.Since(elevator.lastSeen) > 3*time.Second {
			fmt.Printf("Elevator %d unresponsive!\n", id)
			elevator.active = false

			// Redistribute hall calls
			for f := 0; f < N_FLOORS; f++ {
				if elevator.requests[f][B_HallUp] {
					em.AssignHallCall(f, "up")
				}
				if elevator.requests[f][B_HallDown] {
					em.AssignHallCall(f, "down")
				}
			}

			// If master is down, elect again
			if id == em.MasterID {
				em.ElectMaster()
			}
		}
	}
}

// shouldn't really to the redistribute and electing in the network-module as done above
