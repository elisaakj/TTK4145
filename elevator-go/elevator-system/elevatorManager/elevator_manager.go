package elevatorManager

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
	"fmt"
)

func AssignOrders(elevators []*config.SyncElevator, newOrder elevio.ButtonEvent) int {
	bestElevIndex := -1
	bestScore := 100000

	fmt.Printf("We now have %d elevators available\n", len(elevators))
	for i, elev := range elevators {
		fmt.Printf("Elevator ID%s\n", elev.ID)
		if elev.State == config.State(config.UNAVAILABLE) {
			fmt.Printf("Elevator ID%s is UNAVAILABLE\n", elev.ID)
			continue
		}

		score := costFunction(elev, newOrder)

		if score < bestScore {
			bestScore = score
			bestElevIndex = i
		}
	}

	if bestElevIndex != -1 {
		elevators[bestElevIndex].Requests[newOrder.Floor][newOrder.Button].State = config.Order
	}
	return bestElevIndex
}

func ReassignOrders(elevators []*config.SyncElevator, chNewLocalOrder chan<- elevio.ButtonEvent, localElevID string) {
	for _, elev := range elevators {
		if elev.State != config.State(config.UNAVAILABLE) {
			continue
		}

		for floor := 0; floor < config.NUM_FLOORS; floor++ {
			for button := 0; button < config.NUM_BUTTONS-1; button++ {
				if elev.Requests[floor][button].State == config.Order || elev.Requests[floor][button].State == config.Confirmed {
					newOrder := elevio.ButtonEvent{
						Floor:  floor,
						Button: elevio.ButtonType(button),
					}

					assignedIdx := AssignOrders(elevators, newOrder)

					if assignedIdx != -1 {
						elevators[assignedIdx].Requests[floor][button].State = config.Order
						if elevators[assignedIdx].ID == localElevID {
							fmt.Printf("[REASSIGN] Re-issuing order (%d, %v)\n", floor, button)
							chNewLocalOrder <- newOrder
						}
					}

					elev.Requests[floor][button].State = config.None
				}
			}
		}
	}
}

func costFunction(elev *config.SyncElevator, order elevio.ButtonEvent) int {
	distance := abs(elev.Floor - order.Floor)
	cost := distance

	if config.ElevatorState(elev.State) == config.MOVING {
		cost += 2
	}

	if (elev.Dirn == config.Direction(elevio.DIRN_UP) && order.Floor < elev.Floor) ||
		(elev.Dirn == config.Direction(elevio.DIRN_DOWN) && order.Floor > elev.Floor) {
		cost += 3
	}

	if config.ElevatorState(elev.State) == config.IDLE {
		cost -= 2
	}

	if config.ElevatorState(elev.State) == config.IDLE && elev.Floor == order.Floor {
		cost -= 1
	}

	activeCount := 0
	for f := 0; f < config.NUM_FLOORS; f++ {
		for b := 0; b < config.NUM_BUTTONS-1; b++ {
			if elev.Requests[f][b].State == config.Order || elev.Requests[f][b].State == config.Confirmed {
				activeCount++
			}
		}
	}
	cost += activeCount

	return cost
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
