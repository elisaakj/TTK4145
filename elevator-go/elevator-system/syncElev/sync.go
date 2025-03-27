package syncElev

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevatorManager"
	"Driver-go/elevator-system/elevatorStateMachine"
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/orderid"
	"Network-go/network/peers"
	"fmt"
	"time"
)

func syncElevatorInit(id string) config.SyncElevator {
	requests := make([][]config.OrderInfo, config.NUM_FLOORS)
	for floor := range requests {
		requests[floor] = make([]config.OrderInfo, config.NUM_BUTTONS)
	}
	return config.SyncElevator{Requests: requests, ID: id, Floor: 0, Behave: config.Behaviour(config.IDLE)}
}

func broadcast(elevators []*config.SyncElevator, chTx chan<- []config.SyncElevator) {
	tempElev := make([]config.SyncElevator, 0)
	for _, elev := range elevators {
		tempElev = append(tempElev, *elev)
	}
	chTx <- tempElev
	time.Sleep(100 * time.Millisecond)
}

// // call function as goroutine in main, sending to network, reciving from network to sync
// func SyncElevators(id string, chNewLocalOrder chan elevio.ButtonEvent, chNewLocalState chan elevatorStateMachine.Elevator, chMsgFromUDP chan []config.SyncElevator,
// 	chMsgToUDP chan []config.SyncElevator, chOrderToLocal chan elevio.ButtonEvent, chPeerUpdate chan peers.PeerUpdate, chClearLocalHallOrders chan bool) {

// 	// Load persisted OrderID state from file
// 	err := orderid.Load()
// 	if err != nil {
// 		fmt.Println("[OrderID] Warning: Failed to load persistent OrderID store:", err)
// 	}
// 	orderid.DebugPrint()

// 	var localElevatorIndex int

// 	elevators := make([]*config.SyncElevator, 0)
// 	localElevator := new(config.SyncElevator)
// 	*localElevator = syncElevatorInit(id)
// 	elevators = append(elevators, localElevator)
// 	localElevatorIndex = getIndexByID(elevators, id)

// 	connectionTimer := time.NewTimer(time.Duration(3) * time.Second)
// 	select {
// 	case newElevators := <-chMsgFromUDP:
// 		for _, elev := range newElevators {
// 			if elev.ID == elevators[localElevatorIndex].ID {
// 				for floor := range elev.Requests {
// 					if elev.Requests[floor][elevio.BUTTON_CAB].State == config.Confirmed || elev.Requests[floor][elevio.BUTTON_CAB].State == config.Order {
// 						chNewLocalOrder <- elevio.ButtonEvent{
// 							Floor:  floor,
// 							Button: elevio.ButtonType(elevio.BUTTON_CAB)}
// 					}
// 				}
// 			}
// 		}
// 		break
// 	case <-connectionTimer.C:
// 		break
// 	}

// 	for {
// 		select {
// 		case newOrder := <-chNewLocalOrder:
// 			if newOrder.Button == elevio.BUTTON_CAB {
// 				currentID := orderid.IncrementAndGet(newOrder.Floor, int(newOrder.Button))
// 				elevators[getIndexByID(elevators, id)].Requests[newOrder.Floor][newOrder.Button] = config.OrderInfo{
// 					State:   config.Confirmed,
// 					OrderID: currentID,
// 				}
// 				chOrderToLocal <- newOrder
// 				broadcast(elevators, chMsgToUDP)
// 				break
// 			}

// 			fmt.Printf("Assigning hall call: Floor %d, Button %v\n", newOrder.Floor, newOrder.Button)
// 			assignedIdx := elevatorManager.AssignOrders(elevators, newOrder)

// 			if assignedIdx != -1 {
// 				currentID := orderid.IncrementAndGet(newOrder.Floor, int(newOrder.Button))
// 				elevators[assignedIdx].Requests[newOrder.Floor][newOrder.Button] = config.OrderInfo{
// 					State:   config.Order,
// 					OrderID: currentID,
// 				}
// 				fmt.Printf("[ASSIGN] Elevator %s gets (%d,%d) OrderID: %d\n", elevators[assignedIdx].ID, newOrder.Floor, newOrder.Button, currentID)
// 				if elevators[assignedIdx].ID == id {
// 					chOrderToLocal <- newOrder
// 				}
// 				broadcast(elevators, chMsgToUDP)
// 			} else {
// 				fmt.Println("Warning: No elevator available to assign order")
// 			}
// 			setHallLights(elevators)

// 		case newState := <-chNewLocalState:
// 			if newState.Floor != elevators[localElevatorIndex].Floor ||
// 				newState.State == config.IDLE ||
// 				newState.State == config.DOOR_OPEN {
// 				elevators[localElevatorIndex].Behave = config.Behaviour(int(newState.State))
// 				elevators[localElevatorIndex].Floor = newState.Floor
// 				elevators[localElevatorIndex].Dirn = config.Direction(int(newState.Dirn))
// 			}
// 			for floor := range elevators[localElevatorIndex].Requests {
// 				for button := range elevators[localElevatorIndex].Requests[floor] {
// 					if !newState.Requests[floor][button] &&
// 						elevators[localElevatorIndex].Requests[floor][button].State == config.Confirmed {
// 						elevators[localElevatorIndex].Requests[floor][button].State = config.Complete
// 					}
// 					if elevators[localElevatorIndex].Behave != config.Behaviour(config.UNAVAILABLE) &&
// 						newState.Requests[floor][button] &&
// 						elevators[localElevatorIndex].Requests[floor][button].State != config.Confirmed {
// 						elevators[localElevatorIndex].Requests[floor][button].State = config.Confirmed
// 					}
// 				}
// 			}
// 			setHallLights(elevators)
// 			broadcast(elevators, chMsgToUDP)
// 			removeCompletedOrders(elevators)

// 		case newElevators := <-chMsgFromUDP:

// 			updateElevators(elevators, newElevators, localElevatorIndex)
// 			elevatorManager.ReassignOrders(elevators, chNewLocalOrder, id)
// 			for _, newElev := range newElevators {
// 				elevExist := false
// 				for _, elev := range elevators {
// 					if elev.ID == newElev.ID {
// 						elevExist = true
// 						for f := range elev.Requests {
// 							for b := range elev.Requests[f] {
// 								if newElev.Requests[f][b].OrderID > elev.Requests[f][b].OrderID {
// 									elev.Requests[f][b] = newElev.Requests[f][b]
// 									orderid.UpdateIfGreater(f, b, newElev.Requests[f][b].OrderID)
// 									fmt.Printf("[SYNC] %s updated (%d,%d) to OrderID %d\n", elev.ID, f, b, newElev.Requests[f][b].OrderID)
// 								}
// 							}
// 						}
// 						elev.Floor = newElev.Floor
// 						elev.Dir = newElev.Dir
// 						elev.Behave = newElev.Behave
// 					}
// 				}
// 				if !elevExist {
// 					addNewElevator(&elevators, newElev)
// 				}
// 			}
// 			extractNewOrder := comfirmNewOrder(elevators[localElevatorIndex])
// 			setHallLights(elevators)
// 			removeCompletedOrders(elevators)
// 			if extractNewOrder != nil {
// 				tempOrder := elevio.ButtonEvent{
// 					Button: elevio.ButtonType(extractNewOrder.Button),
// 					Floor:  extractNewOrder.Floor}
// 				chOrderToLocal <- tempOrder
// 				broadcast(elevators, chMsgToUDP)
// 			}

// 		case peer := <-chPeerUpdate:
// 			if len(peer.Lost) != 0 {
// 				for _, stringLostID := range peer.Lost {
// 					for _, elev := range elevators {
// 						if stringLostID == elev.ID {
// 							elev.Behave = config.Behaviour(config.UNAVAILABLE)
// 						}
// 						elevatorManager.ReassignOrders(elevators, chNewLocalOrder, id)
// 						for floor := range elev.Requests {
// 							for button := 0; button < len(elev.Requests[floor])-1; button++ {
// 								elev.Requests[floor][button].State = config.None
// 							}
// 						}
// 					}
// 				}
// 			}
// 			setHallLights(elevators)
// 			broadcast(elevators, chMsgToUDP)
// 		}
// 	}
// }

func removeCompletedOrders(elevators []*config.SyncElevator) {
	for _, elev := range elevators {
		for floor := range elev.Requests {
			for button := range elev.Requests[floor] {
				if elev.Requests[floor][button].State == config.Complete {
					elev.Requests[floor][button].State = config.None
				}
			}
		}
	}
}

// Updates local elevator array from received elevator array from network
func updateElevators(elevators []*config.SyncElevator, newElevators []config.SyncElevator, localElevatorIndex int) {
	if elevators[localElevatorIndex].ID != newElevators[localElevatorIndex].ID {
		for _, elev := range elevators {
			if elev.ID == newElevators[localElevatorIndex].ID {
				for floor := range elev.Requests {
					for button := range elev.Requests[floor] {
						if !(elev.Requests[floor][button].State == config.Confirmed && newElevators[localElevatorIndex].Requests[floor][button].State == config.Order) {
							elev.Requests[floor][button] = newElevators[localElevatorIndex].Requests[floor][button]
						}
						elev.Floor = newElevators[localElevatorIndex].Floor
						elev.Dirn = newElevators[localElevatorIndex].Dirn
						elev.Behave = newElevators[localElevatorIndex].Behave
					}
				}
			}
		}
		for _, newElev := range newElevators {
			if newElev.ID == elevators[localElevatorIndex].ID {
				for floor := range newElev.Requests {
					for button := range newElev.Requests[floor] {
						if elevators[localElevatorIndex].Behave != config.Behaviour(config.UNAVAILABLE) {
							if newElev.Requests[floor][button].State == config.Order {
								(*elevators[localElevatorIndex]).Requests[floor][button].State = config.Order
							}
						}
					}
				}
			}
		}
	}
}

// Adds newElevator to local elevator array
func addNewElevator(elevators *[]*config.SyncElevator, newElevator config.SyncElevator) {
	tempElev := new(config.SyncElevator)
	*tempElev = syncElevatorInit(newElevator.ID)
	(*tempElev).Behave = newElevator.Behave
	(*tempElev).Dirn = newElevator.Dirn
	(*tempElev).Floor = newElevator.Floor
	for floor := range tempElev.Requests {
		for button := range tempElev.Requests[floor] {
			tempElev.Requests[floor][button] = newElevator.Requests[floor][button]
		}
	}
	*elevators = append(*elevators, tempElev)
}

func comfirmNewOrder(elev *config.SyncElevator) *elevio.ButtonEvent {
	for floor := range elev.Requests {
		for button := 0; button < len(elev.Requests[floor]); button++ {
			if elev.Requests[floor][button].State == config.Order {
				elev.Requests[floor][button].State = config.Confirmed
				tempOrder := new(elevio.ButtonEvent)
				*tempOrder = elevio.ButtonEvent{
					Floor:  floor,
					Button: elevio.ButtonType(button)}
				return tempOrder
			}
		}
	}
	return nil
}

func setHallLights(elevators []*config.SyncElevator) {
	for button := 0; button < config.NUM_BUTTONS-1; button++ {
		for floor := 0; floor < config.NUM_FLOORS; floor++ {
			isLightOn := false
			for _, elev := range elevators {
				if elev.Requests[floor][button].State == config.Confirmed || elev.Requests[floor][button].State == config.Order {
					isLightOn = true
				}
			}
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, isLightOn)
		}
	}
}

func getIndexByID(elevators []*config.SyncElevator, id string) int {
	for i, elev := range elevators {
		if elev.ID == id {
			return i
		}
	}
	return -1
}

// call function as goroutine in main, sending to network, reciving from network to sync
func SyncElevators(id string, chNewLocalOrder chan elevio.ButtonEvent, chNewLocalState chan elevatorStateMachine.Elevator, chMsgFromUDP chan []config.SyncElevator,
	chMsgToUDP chan []config.SyncElevator, chOrderToLocal chan elevio.ButtonEvent, chPeerUpdate chan peers.PeerUpdate, chClearLocalHallOrders chan bool) {

	// Load persisted OrderID state from file
	err := orderid.Load(id)
	if err != nil {
		fmt.Println("[OrderID] Warning: Failed to load persistent OrderID store:", err)
	}
	orderid.DebugPrint()

	var localElevatorIndex int

	elevators := make([]*config.SyncElevator, 0)
	localElevator := new(config.SyncElevator)
	*localElevator = syncElevatorInit(id)
	elevators = append(elevators, localElevator)
	localElevatorIndex = getIndexByID(elevators, id)

	connectionTimer := time.NewTimer(time.Duration(3) * time.Second)
	select {
	case newElevators := <-chMsgFromUDP:
		for _, elev := range newElevators {
			if elev.ID == elevators[localElevatorIndex].ID {
				for floor := range elev.Requests {
					if elev.Requests[floor][elevio.BUTTON_CAB].State == config.Confirmed || elev.Requests[floor][elevio.BUTTON_CAB].State == config.Order {
						chNewLocalOrder <- elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.ButtonType(elevio.BUTTON_CAB)}
					}
				}
			}
		}
		break
	case <-connectionTimer.C:
		break
	}

	for {
		select {

		//case <-time.After(100 * time.Millisecond):
		//	broadcast(elevators, chMsgToUDP)

		case newOrder := <-chNewLocalOrder:
			if newOrder.Button == elevio.BUTTON_CAB {
				currentID := orderid.IncrementAndGet(newOrder.Floor, int(newOrder.Button), id)
				elevators[getIndexByID(elevators, id)].Requests[newOrder.Floor][newOrder.Button] = config.OrderInfo{
					State:   config.Confirmed,
					OrderID: currentID,
				}
				chOrderToLocal <- newOrder
				broadcast(elevators, chMsgToUDP)
				break
			}

			fmt.Printf("Assigning hall call: Floor %d, Button %v\n", newOrder.Floor, newOrder.Button)

			assignedIdx := elevatorManager.AssignOrders(elevators, newOrder, "")

			if assignedIdx != -1 {
				currentID := orderid.IncrementAndGet(newOrder.Floor, int(newOrder.Button), id)
				elevators[assignedIdx].Requests[newOrder.Floor][newOrder.Button] = config.OrderInfo{
					State:   config.Order,
					OrderID: currentID,
				}
				fmt.Printf("[ASSIGN] Elevator %s gets (%d,%d) OrderID: %d\n", elevators[assignedIdx].ID, newOrder.Floor, newOrder.Button, currentID)
				if elevators[assignedIdx].ID == id {
					chOrderToLocal <- newOrder
				}

				// Try assigning opposite hall call if it exists
				if newOrder.Button == elevio.BUTTON_HALL_UP || newOrder.Button == elevio.BUTTON_HALL_DOWN {
					oppositeButton := elevio.BUTTON_HALL_DOWN
					if newOrder.Button == elevio.BUTTON_HALL_DOWN {
						oppositeButton = elevio.BUTTON_HALL_UP
					}

					oppState := elevators[assignedIdx].Requests[newOrder.Floor][oppositeButton].State
					if oppState == config.Order || oppState == config.Confirmed {
						assignedOpp := elevatorManager.AssignOrders(elevators, elevio.ButtonEvent{
							Floor:  newOrder.Floor,
							Button: oppositeButton,
						}, elevators[assignedIdx].ID)

						if assignedOpp != -1 && assignedOpp != assignedIdx {
							currentID := orderid.IncrementAndGet(newOrder.Floor, int(oppositeButton), id)
							elevators[assignedOpp].Requests[newOrder.Floor][oppositeButton] = config.OrderInfo{
								State:   config.Order,
								OrderID: currentID,
							}
							fmt.Printf("[ASSIGN] Elevator %s gets (%d,%d) OrderID: %d\n", elevators[assignedOpp].ID, newOrder.Floor, oppositeButton, currentID)
							if elevators[assignedOpp].ID == id {
								chOrderToLocal <- elevio.ButtonEvent{Floor: newOrder.Floor, Button: oppositeButton}
							}
						}
					}
				}

				broadcast(elevators, chMsgToUDP)
			} else {
				fmt.Println("Warning: No elevator available to assign order")
			}

			setHallLights(elevators)

		case newState := <-chNewLocalState:
			if newState.State == config.UNAVAILABLE {
				// Set state BEFORE reassignment
				elevators[localElevatorIndex].Behave = config.Behaviour(config.UNAVAILABLE)
				elevatorManager.ReassignOrders(elevators, chNewLocalOrder, id)
			}

			if newState.Floor != elevators[localElevatorIndex].Floor ||
				newState.State == config.IDLE ||
				newState.State == config.DOOR_OPEN {
				elevators[localElevatorIndex].Behave = config.Behaviour(int(newState.State))
				elevators[localElevatorIndex].Floor = newState.Floor
				elevators[localElevatorIndex].Dirn = config.Direction(int(newState.Dirn))
			}

			for floor := range elevators[localElevatorIndex].Requests {
				for button := range elevators[localElevatorIndex].Requests[floor] {
					if !newState.Requests[floor][button] &&
						elevators[localElevatorIndex].Requests[floor][button].State == config.Confirmed {
						elevators[localElevatorIndex].Requests[floor][button].State = config.Complete
					}
					if elevators[localElevatorIndex].Behave != config.Behaviour(config.UNAVAILABLE) &&
						newState.Requests[floor][button] &&
						elevators[localElevatorIndex].Requests[floor][button].State != config.Confirmed {
						elevators[localElevatorIndex].Requests[floor][button].State = config.Confirmed
					}
				}
			}
			setHallLights(elevators)
			broadcast(elevators, chMsgToUDP)
			removeCompletedOrders(elevators)

		case newElevators := <-chMsgFromUDP:

			updateElevators(elevators, newElevators, localElevatorIndex)
			elevatorManager.ReassignOrders(elevators, chNewLocalOrder, id)
			for _, newElev := range newElevators {
				elevExist := false
				for _, elev := range elevators {
					if elev.ID == newElev.ID {
						elevExist = true
						for f := range elev.Requests {
							for b := range elev.Requests[f] {
								if newElev.Requests[f][b].OrderID > elev.Requests[f][b].OrderID {
									elev.Requests[f][b] = newElev.Requests[f][b]
									orderid.UpdateIfGreater(f, b, newElev.Requests[f][b].OrderID, id)
									fmt.Printf("[SYNC] %s updated (%d,%d) to OrderID %d\n", elev.ID, f, b, newElev.Requests[f][b].OrderID)
								}
							}
						}
						elev.Floor = newElev.Floor
						elev.Dirn = newElev.Dirn
						elev.Behave = newElev.Behave
					}
				}
				if !elevExist {
					addNewElevator(&elevators, newElev)
				}
			}
			extractNewOrder := comfirmNewOrder(elevators[localElevatorIndex])
			setHallLights(elevators)
			removeCompletedOrders(elevators)
			if extractNewOrder != nil {
				tempOrder := elevio.ButtonEvent{
					Button: elevio.ButtonType(extractNewOrder.Button),
					Floor:  extractNewOrder.Floor}
				chOrderToLocal <- tempOrder
				broadcast(elevators, chMsgToUDP)
			}

		case peer := <-chPeerUpdate:

			var currentPeers []string
			for peerUpdate := range chPeerUpdate {
				currentPeers = peerUpdate.Peers
				fmt.Println("=== Alive peers in sync ===")
				for _, peer := range currentPeers {
					fmt.Println(" -", peer)
				}
				fmt.Println("===================")
			}

			if len(peer.Lost) != 0 {
				for _, stringLostID := range peer.Lost {
					for _, elev := range elevators {
						if stringLostID == elev.ID {
							elev.Behave = config.Behaviour(config.UNAVAILABLE)
						}
						elevatorManager.ReassignOrders(elevators, chNewLocalOrder, id)
						for floor := range elev.Requests {
							for button := 0; button < len(elev.Requests[floor])-1; button++ {
								elev.Requests[floor][button].State = config.None
							}
						}
					}
				}
			}
			setHallLights(elevators)
			broadcast(elevators, chMsgToUDP)
		}
	}
}
