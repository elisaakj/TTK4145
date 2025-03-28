package syncElev

import (
	"Driver-go/elevator-system/common"
	"Driver-go/elevator-system/elevatorManager"
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/orderid"
	"Network-go/network/peers"
	"fmt"
	"time"
)

func syncElevatorInit(id string) common.SyncElevator {
	requests := make([][]common.OrderInfo, common.NUM_FLOORS)
	for floor := range requests {
		requests[floor] = make([]common.OrderInfo, common.NUM_BUTTONS)
	}
	return common.SyncElevator{Requests: requests, ID: id, Floor: 0, State: common.ElevatorState(common.IDLE)}
}

func broadcast(elevators []*common.SyncElevator, chTx chan<- []common.SyncElevator) {
	tempElev := make([]common.SyncElevator, 0)
	for _, elev := range elevators {
		tempElev = append(tempElev, *elev)
	}
	chTx <- tempElev
	time.Sleep(100 * time.Millisecond)
}

// call function as goroutine in main, sending to network, reciving from network to sync
func SyncElevators(id string, chNewLocalOrder chan common.ButtonEvent, chNewLocalState chan common.Elevator, chMsgFromUDP chan []common.SyncElevator,
	chMsgToUDP chan []common.SyncElevator, chOrderToLocal chan common.ButtonEvent, chPeerUpdate chan peers.PeerUpdate, chClearLocalHallOrders chan bool) {

	err := orderid.Load(id)
	if err != nil {
		fmt.Println("[OrderID] Warning: Failed to load persistent OrderID store:", err)
	}

	var localElevatorIndex int

	elevators := make([]*common.SyncElevator, 0)
	localElevator := new(common.SyncElevator)
	*localElevator = syncElevatorInit(id)
	elevators = append(elevators, localElevator)
	localElevatorIndex = getIndexByID(elevators, id)

	connectionTimer := time.NewTimer(time.Duration(common.CONNECTION_TIMER) * time.Second)
	select {
	case newElevators := <-chMsgFromUDP:
		for _, elev := range newElevators {
			if elev.ID == elevators[localElevatorIndex].ID {
				for floor := range elev.Requests {
					if elev.Requests[floor][common.BUTTON_CAB].State == common.CONFIRMED || elev.Requests[floor][common.BUTTON_CAB].State == common.ORDER {
						chNewLocalOrder <- common.ButtonEvent{
							Floor:  floor,
							Button: common.ButtonType(common.BUTTON_CAB)}
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
			if newOrder.Button == common.BUTTON_CAB {
				currentID := orderid.IncrementAndGet(newOrder.Floor, int(newOrder.Button), id)
				elevators[getIndexByID(elevators, id)].Requests[newOrder.Floor][newOrder.Button] = common.OrderInfo{
					State:   common.CONFIRMED,
					OrderID: currentID,
				}
				chOrderToLocal <- newOrder
				broadcast(elevators, chMsgToUDP)
				break
			}

			fmt.Printf("Assigning hall call: Floor %d, Button %v\n", newOrder.Floor, newOrder.Button)

			assignedIdx := elevatorManager.AssignOrders(elevators, newOrder)

			if assignedIdx != -1 {
				currentID := orderid.IncrementAndGet(newOrder.Floor, int(newOrder.Button), id)
				elevators[assignedIdx].Requests[newOrder.Floor][newOrder.Button] = common.OrderInfo{
					State:   common.ORDER,
					OrderID: currentID,
				}
				fmt.Printf("[ASSIGN] Elevator %s gets (%d,%d) OrderID: %d\n", elevators[assignedIdx].ID, newOrder.Floor, newOrder.Button, currentID)
				if elevators[assignedIdx].ID == id {
					chOrderToLocal <- newOrder
				}

				if newOrder.Button == common.BUTTON_HALL_UP || newOrder.Button == common.BUTTON_HALL_DOWN {
					oppositeButton := common.BUTTON_HALL_DOWN
					if newOrder.Button == common.BUTTON_HALL_DOWN {
						oppositeButton = common.BUTTON_HALL_UP
					}

					oppState := elevators[assignedIdx].Requests[newOrder.Floor][oppositeButton].State
					if oppState == common.ORDER || oppState == common.CONFIRMED {
						assignedOpp := elevatorManager.AssignOrders(elevators, common.ButtonEvent{
							Floor:  newOrder.Floor,
							Button: oppositeButton,
						})

						if assignedOpp != -1 && assignedOpp != assignedIdx {
							currentID := orderid.IncrementAndGet(newOrder.Floor, int(oppositeButton), id)
							elevators[assignedOpp].Requests[newOrder.Floor][oppositeButton] = common.OrderInfo{
								State:   common.ORDER,
								OrderID: currentID,
							}
							fmt.Printf("Elevator %s gets (%d,%d) OrderID: %d\n", elevators[assignedOpp].ID, newOrder.Floor, oppositeButton, currentID)
							if elevators[assignedOpp].ID == id {
								chOrderToLocal <- common.ButtonEvent{Floor: newOrder.Floor, Button: oppositeButton}
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
			if newState.State == common.UNAVAILABLE {
				elevators[localElevatorIndex].State = common.ElevatorState(common.UNAVAILABLE)
				elevatorManager.ReassignOrders(elevators, chNewLocalOrder, id)
			}

			if newState.Floor != elevators[localElevatorIndex].Floor ||
				newState.State == common.IDLE ||
				newState.State == common.DOOR_OPEN {
				elevators[localElevatorIndex].State = common.ElevatorState(int(newState.State))
				elevators[localElevatorIndex].Floor = newState.Floor
				elevators[localElevatorIndex].Dirn = common.MotorDirection(int(newState.Dirn))
			}

			for floor := range elevators[localElevatorIndex].Requests {
				for button := range elevators[localElevatorIndex].Requests[floor] {
					if !newState.Requests[floor][button] &&
						elevators[localElevatorIndex].Requests[floor][button].State == common.CONFIRMED {
						elevators[localElevatorIndex].Requests[floor][button].State = common.COMPLETE
					}
					if elevators[localElevatorIndex].State != common.ElevatorState(common.UNAVAILABLE) &&
						newState.Requests[floor][button] &&
						elevators[localElevatorIndex].Requests[floor][button].State != common.CONFIRMED {
						elevators[localElevatorIndex].Requests[floor][button].State = common.CONFIRMED
					}
				}
			}
			setHallLights(elevators)
			broadcast(elevators, chMsgToUDP)
			removeOrdersCompleted(elevators)

		case newElevators := <-chMsgFromUDP:

			updateElev(elevators, newElevators, localElevatorIndex)
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
								}
							}
						}
						elev.Floor = newElev.Floor
						elev.Dirn = newElev.Dirn
						elev.State = newElev.State
					}
				}
				if !elevExist {
					addNewElev(&elevators, newElev)
				}
			}
			extractNewOrder := newOrderConfirm(elevators[localElevatorIndex])
			setHallLights(elevators)
			removeOrdersCompleted(elevators)
			if extractNewOrder != nil {
				tempOrder := common.ButtonEvent{
					Button: common.ButtonType(extractNewOrder.Button),
					Floor:  extractNewOrder.Floor}
				chOrderToLocal <- tempOrder
				broadcast(elevators, chMsgToUDP)
			}

		case peer := <-chPeerUpdate:

			if len(peer.Lost) != 0 {
				for _, stringLostID := range peer.Lost {
					for _, elev := range elevators {
						if stringLostID == elev.ID {
							elev.State = common.ElevatorState(common.UNAVAILABLE)
						}
						elevatorManager.ReassignOrders(elevators, chNewLocalOrder, id)
						for floor := range elev.Requests {
							for button := 0; button < len(elev.Requests[floor])-1; button++ {
								elev.Requests[floor][button].State = common.NONE
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

func addNewElev(elevators *[]*common.SyncElevator, newElevator common.SyncElevator) {
	temp := new(common.SyncElevator)
	*temp = syncElevatorInit(newElevator.ID)
	(*temp).State = newElevator.State
	(*temp).Dirn = newElevator.Dirn
	(*temp).Floor = newElevator.Floor
	for floor := range temp.Requests {
		for button := range temp.Requests[floor] {
			temp.Requests[floor][button] = newElevator.Requests[floor][button]
		}
	}
	*elevators = append(*elevators, temp)
}

func newOrderConfirm(elev *common.SyncElevator) *common.ButtonEvent {
	for floor := range elev.Requests {
		for button := 0; button < len(elev.Requests[floor]); button++ {
			if elev.Requests[floor][button].State == common.ORDER {
				elev.Requests[floor][button].State = common.CONFIRMED
				temp := new(common.ButtonEvent)
				*temp = common.ButtonEvent{
					Floor:  floor,
					Button: common.ButtonType(button)}
				return temp
			}
		}
	}
	return nil
}

func removeOrdersCompleted(elevators []*common.SyncElevator) {
	for _, elev := range elevators {
		for floor := range elev.Requests {
			for button := range elev.Requests[floor] {
				if elev.Requests[floor][button].State == common.COMPLETE {
					elev.Requests[floor][button].State = common.NONE
				}
			}
		}
	}
}

func updateElev(elevators []*common.SyncElevator, newElevators []common.SyncElevator, localElevatorIndex int) {
	if elevators[localElevatorIndex].ID != newElevators[localElevatorIndex].ID {
		for _, elev := range elevators {
			if elev.ID == newElevators[localElevatorIndex].ID {
				for floor := range elev.Requests {
					for button := range elev.Requests[floor] {
						if !(elev.Requests[floor][button].State == common.CONFIRMED && newElevators[localElevatorIndex].Requests[floor][button].State == common.ORDER) {
							elev.Requests[floor][button] = newElevators[localElevatorIndex].Requests[floor][button]
						}
						elev.Floor = newElevators[localElevatorIndex].Floor
						elev.Dirn = newElevators[localElevatorIndex].Dirn
						elev.State = newElevators[localElevatorIndex].State
					}
				}
			}
		}
		for _, newElev := range newElevators {
			if newElev.ID == elevators[localElevatorIndex].ID {
				for floor := range newElev.Requests {
					for button := range newElev.Requests[floor] {
						if elevators[localElevatorIndex].State != common.ElevatorState(common.UNAVAILABLE) {
							if newElev.Requests[floor][button].State == common.ORDER {
								(*elevators[localElevatorIndex]).Requests[floor][button].State = common.ORDER
							}
						}
					}
				}
			}
		}
	}
}

func setHallLights(elevators []*common.SyncElevator) {
	for button := 0; button < common.NUM_BUTTONS-1; button++ {
		for floor := 0; floor < common.NUM_FLOORS; floor++ {
			isLightOn := false
			for _, elev := range elevators {
				if elev.Requests[floor][button].State == common.CONFIRMED || elev.Requests[floor][button].State == common.ORDER {
					isLightOn = true
				}
			}
			elevio.SetButtonLamp(common.ButtonType(button), floor, isLightOn)
		}
	}
}

func getIndexByID(elevators []*common.SyncElevator, id string) int {
	for i, elev := range elevators {
		if elev.ID == id {
			return i
		}
	}
	return -1
}
