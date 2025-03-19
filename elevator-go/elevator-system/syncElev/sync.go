package sync

import (
	"Driver-go/elevator-system/elevio"
)

// Testing new implementation
type HallCallUpdate struct {
	ElevatorID int
	OrderID    int
	Floor      int
	Button     elevio.ButtonType
}

// Processes received hall call updates
func listenForHallCallUpdates(hallCallRx chan HallCallUpdate, updateChannel chan ElevatorState, hallCallTx chan HallCallUpdate) {
	for update := range hallCallRx {
		currentState, exists := getPeerStatus(update.ElevatorID)
		if !exists || update.OrderID > currentState.OrderID {
			// Update state with new hall call
			currentState.Requests[update.Floor][update.Button] = true
			currentState.OrderID = update.OrderID

			// Store updated state
			PeerStatus.Store(update.ElevatorID, currentState)

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

func broadcastHallCall(elevator Elevator, event elevio.ButtonEvent, hallCallTx chan HallCallUpdate) {
	msg := HallCallUpdate{
		ElevatorID: elevator.ID,
		OrderID:    elevator.OrderID,
		Floor:      event.Floor,
		Button:     event.Button,
	}
	hallCallTx <- msg
}

func handleReceivedHallCall(update HallCallUpdate, elevator *Elevator, hallCallTx chan HallCallUpdate) {
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
