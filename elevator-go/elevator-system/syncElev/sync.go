package syncElev

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevatorManager"
	"Driver-go/elevator-system/elevatorStateMachine"
	"Driver-go/elevator-system/elevio"
	"Network-go/network/peers"
	"fmt"
	"time"
)

func syncElevatorInit(id string) config.SyncElevator {
	requests := make([][]config.RequestState, config.NUM_FLOORS)
	//orderID := make([][]int, config.NUM_FLOORS)
	for floor := range requests {
		requests[floor] = make([]config.RequestState, config.NUM_BUTTONS)
		//	orderID[floor] = make([]int, config.NUM_BUTTONS)
	}
	return config.SyncElevator{Requests: requests /*OrderID: orderID,*/, ID: id, Floor: 0, Behave: config.Behaviour(config.IDLE)}
}

func broadcast(elevators []*config.SyncElevator, chTx chan<- []config.SyncElevator) {
	tempElev := make([]config.SyncElevator, 0)
	for _, elev := range elevators {
		tempElev = append(tempElev, *elev)
	}
	chTx <- tempElev
	time.Sleep(100 * time.Millisecond)
}

// call function as goroutine in main, sending to network, reciving from network to sync
func SyncElevators(id string, chNewLocalOrder chan elevio.ButtonEvent, chNewLocalState chan elevatorStateMachine.Elevator, chMsgFromUDP chan []config.SyncElevator,
	chMsgToUDP chan []config.SyncElevator, chOrderToLocal chan elevio.ButtonEvent, chPeerUpdate chan peers.PeerUpdate, chClearLocalHallOrders chan bool) {

	var LOCAL_ELEVATOR int

	elevators := make([]*config.SyncElevator, 0)
	localElevator := new(config.SyncElevator)
	*localElevator = syncElevatorInit(id)
	elevators = append(elevators, localElevator)

	LOCAL_ELEVATOR = getIndexByID(elevators, id)

	connectionTimer := time.NewTimer(time.Duration(3) * time.Second)
	select {
	case newElevators := <-chMsgFromUDP:
		for _, elev := range newElevators {
			if elev.ID == elevators[LOCAL_ELEVATOR].ID {
				for floor := range elev.Requests {
					if elev.Requests[floor][elevio.BUTTON_CAB] == config.Confirmed || elev.Requests[floor][elevio.BUTTON_CAB] == config.Order {
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
		case newOrder := <-chNewLocalOrder:
			if newOrder.Button == elevio.BUTTON_CAB {
				elevators[getIndexByID(elevators, id)].Requests[newOrder.Floor][newOrder.Button] = config.Confirmed
				chOrderToLocal <- newOrder
				broadcast(elevators, chMsgToUDP)
				break
			}

			fmt.Printf("Assigning hall call: Floor %d, Button %v\n", newOrder.Floor, newOrder.Button)
			assignedIdx := elevatorManager.AssignOrders(elevators, newOrder)

			if assignedIdx != -1 {
				elevators[assignedIdx].Requests[newOrder.Floor][newOrder.Button] = config.Order
				//elevators[assignedIdx].OrderID[newOrder.Floor][newOrder.Button]++
				fmt.Printf("Assigned to elevator ID: %s\n", elevators[assignedIdx].ID)
				if elevators[assignedIdx].ID == id {
					chOrderToLocal <- newOrder
				}
				broadcast(elevators, chMsgToUDP)
			} else {
				fmt.Println("Warning: No elevator available to assign order")
			}
			setHallLights(elevators)

		case newState := <-chNewLocalState:
			if newState.Floor != elevators[LOCAL_ELEVATOR].Floor ||
				newState.State == config.IDLE ||
				newState.State == config.DOOR_OPEN {
				elevators[LOCAL_ELEVATOR].Behave = config.Behaviour(int(newState.State))
				elevators[LOCAL_ELEVATOR].Floor = newState.Floor
				elevators[LOCAL_ELEVATOR].Dir = config.Direction(int(newState.Dirn))
			}
			for floor := range elevators[LOCAL_ELEVATOR].Requests {
				for button := range elevators[LOCAL_ELEVATOR].Requests[floor] {
					if !newState.Requests[floor][button] &&
						elevators[LOCAL_ELEVATOR].Requests[floor][button] == config.Confirmed {
						elevators[LOCAL_ELEVATOR].Requests[floor][button] = config.Complete
					}
					if elevators[LOCAL_ELEVATOR].Behave != config.Behaviour(config.UNAVAILABLE) &&
						newState.Requests[floor][button] &&
						elevators[LOCAL_ELEVATOR].Requests[floor][button] != config.Confirmed {
						elevators[LOCAL_ELEVATOR].Requests[floor][button] = config.Confirmed
					}
				}
			}
			setHallLights(elevators)
			broadcast(elevators, chMsgToUDP)
			removeCompletedOrders(elevators)

		case newElevators := <-chMsgFromUDP:
			// need to increment and have orderID do the correct thing here as well

			updateElevators(elevators, newElevators, LOCAL_ELEVATOR)
			elevatorManager.ReassignOrders(elevators, chNewLocalOrder, id)
			for _, newElev := range newElevators {
				elevExist := false
				for _, elev := range elevators {
					if elev.ID == newElev.ID {
						elevExist = true
						break
					}
				}
				if !elevExist {
					addNewElevator(&elevators, newElev)
				}
			}
			extractNewOrder := comfirmNewOrder(elevators[LOCAL_ELEVATOR])
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
			if len(peer.Lost) != 0 {
				for _, stringLostID := range peer.Lost {
					for _, elev := range elevators {
						if stringLostID == elev.ID {
							elev.Behave = config.Behaviour(config.UNAVAILABLE)
						}
						elevatorManager.ReassignOrders(elevators, chNewLocalOrder, id)
						for floor := range elev.Requests {
							for button := 0; button < len(elev.Requests[floor])-1; button++ {
								elev.Requests[floor][button] = config.None
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

func removeCompletedOrders(elevators []*config.SyncElevator) {
	for _, elev := range elevators {
		for floor := range elev.Requests {
			for button := range elev.Requests[floor] {
				if elev.Requests[floor][button] == config.Complete {
					elev.Requests[floor][button] = config.None
				}
			}
		}
	}
}

// Updates local elevator array from received elevator array from network
func updateElevators(elevators []*config.SyncElevator, newElevators []config.SyncElevator, LOCAL_ELEVATOR int) {
	if elevators[LOCAL_ELEVATOR].ID != newElevators[LOCAL_ELEVATOR].ID {
		for _, elev := range elevators {
			if elev.ID == newElevators[LOCAL_ELEVATOR].ID {
				for floor := range elev.Requests {
					for button := range elev.Requests[floor] {
						if !(elev.Requests[floor][button] == config.Confirmed && newElevators[LOCAL_ELEVATOR].Requests[floor][button] == config.Order) {
							elev.Requests[floor][button] = newElevators[LOCAL_ELEVATOR].Requests[floor][button]
						}
						elev.Floor = newElevators[LOCAL_ELEVATOR].Floor
						elev.Dir = newElevators[LOCAL_ELEVATOR].Dir
						elev.Behave = newElevators[LOCAL_ELEVATOR].Behave
					}
				}
			}
		}
		for _, newElev := range newElevators {
			if newElev.ID == elevators[LOCAL_ELEVATOR].ID {
				for floor := range newElev.Requests {
					for button := range newElev.Requests[floor] {
						if elevators[LOCAL_ELEVATOR].Behave != config.Behaviour(config.UNAVAILABLE) {
							if newElev.Requests[floor][button] == config.Order {
								(*elevators[LOCAL_ELEVATOR]).Requests[floor][button] = config.Order
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
	(*tempElev).Dir = newElevator.Dir
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
			if elev.Requests[floor][button] == config.Order {
				elev.Requests[floor][button] = config.Confirmed
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
				if elev.Requests[floor][button] == config.Confirmed || elev.Requests[floor][button] == config.Order {
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
