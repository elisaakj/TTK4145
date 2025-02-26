package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Constants for the network module
const (
	//UDP_PORT        = "30000"
	//BROADCAST_ADDR  = "255.255.255.255:30000"
	RETRANSMIT_RATE = 500 * time.Millisecond
	TIMEOUT_LIMIT   = 2 * time.Second
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
}

// PeerStaus tracks last update time for each elevator
var PeerStatus = make(map[int]ElevatorState)

// UDP socket
var udpConn *net.UDPConn

// initNetwork initializes the UDP connection
func initNetwork(elevatorID int, updateChannel chan ElevatorState) {
    // Use a unique port for each elevator based on its ID
    localPort := fmt.Sprintf("30%03d", elevatorID)  // e.g., 30001, 30002, 30003

    addr, err := net.ResolveUDPAddr("udp", ":"+localPort)
    if err != nil {
        fmt.Println("Error resolving UDP address:", err)
        return
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        fmt.Println("Error starting UDP server on port", localPort, ":", err)
        return
    }
    udpConn = conn

    fmt.Println("Elevator", elevatorID, "listening on UDP port", localPort)

    // Listening for messages
    go listenForUpdates(updateChannel)

    // Periodic retransmission of state updates
    go retransmitState(elevatorID, updateChannel)
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

	conn, err := net.DialUDP("udp", nil, nil) // Open UDP connection for sending
    if err != nil {
        fmt.Println("Error creating UDP connection:", err)
        return
    }
    defer conn.Close()

    for range ticker.C {
        for peerID, state := range PeerStatus {
            if peerID != elevatorID { // Don't send to itself
                peerPort := fmt.Sprintf("30%03d", peerID)  // Get peer's listening port (e.g., 30002)
                addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+peerPort)
                if err != nil {
                    fmt.Println("Error resolving UDP address:", err)
                    continue
                }
                
                sendStateUpdate(state, conn, addr)
            }
        }
    }
}

// sendStateUpdate serializes and broadcasts the state
func sendStateUpdate(elevator ElevatorState, conn *net.UDPConn, addr *net.UDPAddr) {
    data, err := json.Marshal(elevator)
    if err != nil {
        fmt.Println("Error with JSON format:", err)
        return
    }

    _, err = conn.WriteToUDP(data, addr)
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
