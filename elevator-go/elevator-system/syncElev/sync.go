package syncElev

import (
	"Driver-go/elevator-system/communication"
	"Driver-go/elevator-system/elevatorStateMachine"
	"Driver-go/elevator-system/elevio"
)

/*
// Hall Call Channels
	hallCallRx := make(chan syncElev.HallCallUpdate)
	hallCallTx := make(chan syncElev.HallCallUpdate)

	go bcast.Transmitter(20200, hallCallTx)
	go bcast.Receiver(20200, hallCallRx)

	// Start Hall Call Update Listener
	go syncElev.ListenForHallCallUpdates(hallCallRx, updateChannel, hallCallTx)
	//go syncElev.BroadcastHallCall(<-updateChannel, <-drv_buttons, hallCallTx)
*/

// The code above here we tried implemnt during the lab 19.03

// Testing new implementation
type HallCallUpdate struct {
	ElevatorID int
	OrderID    int
	Floor      int
	Button     elevio.ButtonType
}

// Processes received hall call updates
func ListenForHallCallUpdates(hallCallRx chan HallCallUpdate, updateChannel chan communication.ElevatorState, hallCallTx chan HallCallUpdate) {

	for update := range hallCallRx {
		currentState, exists := communication.GetPeerStatus(update.ElevatorID)
		if !exists || update.OrderID > currentState.OrderID {
			// Update state with new hall call
			currentState.Requests[update.Floor][update.Button] = true
			currentState.OrderID = update.OrderID

			// Store updated state
			communication.PeerStatus.Store(update.ElevatorID, currentState)

			// Notify FSM of new request
			updateChannel <- currentState

			// Rebroadcast confirmation
			sendHallCallUpdate(update.ElevatorID, update.OrderID, update.Floor, update.Button, hallCallTx)
		}
	}
}

// Sends a hall call update
func sendHallCallUpdate(elevatorID int, orderID int, floor int, button elevio.ButtonType, hallCallTx chan HallCallUpdate) {
	update := HallCallUpdate{
		ElevatorID: elevatorID,
		OrderID:    orderID,
		Floor:      floor,
		Button:     button,
	}
	hallCallTx <- update
}

func BroadcastHallCall(elevator communication.ElevatorState, event elevio.ButtonEvent, hallCallTx chan HallCallUpdate) {
	msg := HallCallUpdate{
		ElevatorID: elevator.ID,
		OrderID:    elevator.OrderID,
		Floor:      event.Floor,
		Button:     event.Button,
	}
	hallCallTx <- msg
}

func handleReceivedHallCall(update HallCallUpdate, elevator *elevatorStateMachine.Elevator, hallCallTx chan HallCallUpdate) {
	if update.OrderID <= elevator.OrderID {
		return // Ignore duplicate or outdated updates
	}

	// Update state
	elevator.Requests[update.Floor][update.Button] = true
	elevator.OrderID = update.OrderID
	elevio.SetButtonLamp(update.Button, update.Floor, true)

	// Rebroadcast confirmation
	sendHallCallUpdate(update.ElevatorID, update.OrderID, update.Floor, update.Button, hallCallTx)
}

////// MAYBE WE CAN DO SOMETHING LIKE THIS

// we need to sync more than just the hallcab calls
// then we make an own module for distributing the orderes to the elevators

type SyncChannels struct {
}

// call function as goroutine in main, sending to network, reciving from network to sync
func SyncElevators(ch SyncChannels, id int) {
}
