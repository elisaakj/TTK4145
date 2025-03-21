package communication

import (
	//"Driver-go/elevator-system/elevatorStateMachine"
	"Driver-go/elevator-system/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"fmt"
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
	Floor    int                   `json:"floor"`
	Dirn     elevio.MotorDirection `json:"dirn"`
	Requests [][]bool              `json:"requests"`
	Active   bool                  `json:"active"`
	//Behaviour elevatorStateMachine.ElevatorBehaviour `json:"behavoiur"`
	// The ones above are the same as in the Elevator struct minus a few, the ones below is needed for the state
	// Probably a cleaner way to implement this, since we already have a similar struct

	ID          int       `json:"id"`
	IsMaster    bool      `json:"is_master"`
	LastUpdated time.Time `json:"-"`
	Heartbeat   time.Time `json:"-"`
	OrderID     int
}

var (
	PeerStatus  sync.Map // PeerStaus tracks last update time for each elevator
	activePeers []string
)

// initNetwork initializes the UDP connection
func InitNetwork(elevatorID int, updateChannel chan ElevatorState) {
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(15647, fmt.Sprintf("elevator-%d", elevatorID), peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	statePort := 20100 + elevatorID
	stateTx := make(chan ElevatorState)
	stateRx := make(chan ElevatorState)

	go bcast.Transmitter(statePort, stateTx)
	go bcast.Receiver(statePort, stateRx)

	// //  Hall Call Channels
	// hallCallTx := make(chan HallCallUpdate)
	// hallCallRx := make(chan HallCallUpdate)

	// //  Start Hall Call Transmitter/Receiver
	// go bcast.Transmitter(20200, hallCallTx)
	// go bcast.Receiver(20200, hallCallRx)

	// //  Start Hall Call Update Listener
	// go listenForHallCallUpdates(hallCallRx, updateChannel)

	go func() {
		for {
			select {
			case peerUpdate := <-peerUpdateCh:
				activePeers = peerUpdate.Peers
				fmt.Printf("Peers: %q\n:", peerUpdate.Peers)
				fmt.Printf("New: %q\n:", peerUpdate.New)
				fmt.Printf("Lost: %q\n:", peerUpdate.Lost)
				time.Sleep(1 * time.Second)
				// Kan bruke print over her til testing, men må ha stopTimer
			case receivedState := <-stateRx:
				updateChannel <- receivedState
			}
		}
	}()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			if state, exists := GetPeerStatus(elevatorID); exists {
				stateTx <- state
			}
		}
	}()
}

// listenForUpdates recives UDP packets and updates peerStatus

func GetPeerStatus(id int) (ElevatorState, bool) {
	val, ok := PeerStatus.Load(id)
	if ok {
		return val.(ElevatorState), true
	}
	return ElevatorState{}, false
}
