package main

import (
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
	floor     int                       `json:"floor"`
	dirn      MotorDirection            `json:"dirn"`
	requests  [N_FLOORS][N_BUTTONS]bool `json:"requests"`
	Dirn      MotorDirection            `json:"dirn"`
	Requests  [N_FLOORS][N_BUTTONS]bool `json:"requests"`
	active    bool
	state ElevState

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
)

// PeerStatus as sync.Map instead??
// sateUpdates as a chan with the stateUpdates

// initNetwork initializes the UDP connection
func initNetwork(elevatorID int, updateChannel chan ElevatorState) {
	addr, _ := net.ResolveUDPAddr("udp", ":"+UDP_PORT)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP server", err)
		return
	}
	udpConn = conn

	go listenForUpdates(updateChannel) // Listening for messages
	go retransmitState(elevatorID)     // Periodic retransmission of state updates
	go sendHeartbeat(elevatorID)
	go detectFailures()
	go processStateUpdates(updateChannel)
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
	addr, _ := net.ResolveUDPAddr("udp", BROADCAST_ADDR)
	ticker := time.NewTicker(RETRANSMIT_RATE)
	defer ticker.Stop()

	for range ticker.C {
		if state, exists := getPeerStatus(elevatorID); exists {
			sendStateUpdate(state, addr)
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
		fmt.Println("Error sending UDP packet:", err) // error
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
