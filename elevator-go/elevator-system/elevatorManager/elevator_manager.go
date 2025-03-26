package elevatorManager

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
)

func AssignOrders(elevators []*config.SyncElevator, newOrder elevio.ButtonEvent, excludeID string) int {
	bestElevIndex := -1
	bestScore := 100000

	for i, elev := range elevators {
		
		if elev.Behave == config.Behaviour(3) { // UNAVAILABLE
			continue
		}

		score := costFunction(elev, newOrder)
		if elev.ID == excludeID {
    		score += 5 // discourage assigning to the same elevator again
		}

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
		if elev.Behave != config.Behaviour(3) { // UNAVAILABLE
			continue
		}
		for floor := 0; floor < config.NUM_FLOORS; floor++ {
			for button := 0; button < config.NUM_BUTTONS-1; button++ {
				if elev.Requests[floor][button].State == config.Order || elev.Requests[floor][button].State == config.Confirmed {
					newOrder := elevio.ButtonEvent{
						Floor:  floor,
						Button: elevio.ButtonType(button),
					}
					
					assignedIdx := AssignOrders(elevators, newOrder, "")

					if assignedIdx != -1 {
						elevators[assignedIdx].Requests[floor][button].State = config.Order
						if elevators[assignedIdx].ID == localElevID { // local elevator ID
							chNewLocalOrder <- newOrder
						}
					}
					// Clear order freom unavailable elevator
					elev.Requests[floor][button].State = config.None
				}
			}
		}
	}
}

// func costFunction(e *config.SyncElevator, order elevio.ButtonEvent) int {
// 	score := 0
// 	score += abs(order.Floor - e.Floor)

// 	if e.Behave == config.Behaviour(0) { // IDLE
// 		score -= 2
// 	}
// 	if e.Dir == config.Up && order.Floor >= e.Floor {
// 		score -= 1
// 	}
// 	if e.Dir == config.Down && order.Floor <= e.Floor {
// 		score -= 1
// 	}

// 	return score
// }

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}


func costFunction(elev *config.SyncElevator, order elevio.ButtonEvent) int {
	distance := abs(elev.Floor - order.Floor)
	cost := distance

	// Penalize elevators currently moving
	if config.ElevatorState(elev.Behave) == config.MOVING {
		cost += 2
	}

	// Penalize elevators going in the wrong direction
	if (elev.Dir == config.Up && order.Floor < elev.Floor) ||
		(elev.Dir == config.Down && order.Floor > elev.Floor) {
		cost += 4
	}

	// Slight bonus for being idle on same floor
	if config.ElevatorState(elev.Behave) == config.IDLE && elev.Floor == order.Floor {
		cost -= 1
	}

	return cost
}


