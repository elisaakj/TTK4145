package elevatorManager

import (
	"Driver-go/elevator-system/common"
	"fmt"
)

func AssignOrders(elevators []*common.SyncElevator, newOrder common.ButtonEvent) int {
	bestElevIndex := -1
	bestScore := 100000

	fmt.Printf("We now have %d elevators available\n", len(elevators))
	for i, elev := range elevators {
		fmt.Printf("Elevator ID%s\n", elev.ID)
		if elev.State == common.ElevatorState(common.UNAVAILABLE) {
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
		elevators[bestElevIndex].Requests[newOrder.Floor][newOrder.Button].State = common.ORDER
	}
	return bestElevIndex
}

func ReassignOrders(elevators []*common.SyncElevator, chNewLocalOrder chan<- common.ButtonEvent, localElevID string) {
	for _, elev := range elevators {
		if elev.State != common.ElevatorState(common.UNAVAILABLE) {
			continue
		}

		for floor := 0; floor < common.NUM_FLOORS; floor++ {
			for button := 0; button < common.NUM_BUTTONS-1; button++ {
				if elev.Requests[floor][button].State == common.ORDER || elev.Requests[floor][button].State == common.CONFIRMED {
					newOrder := common.ButtonEvent{
						Floor:  floor,
						Button: common.ButtonType(button),
					}

					assignedIdx := AssignOrders(elevators, newOrder)

					if assignedIdx != -1 {
						elevators[assignedIdx].Requests[floor][button].State = common.ORDER
						if elevators[assignedIdx].ID == localElevID {
							fmt.Printf("[REASSIGN] Re-issuing order (%d, %v)\n", floor, button)
							chNewLocalOrder <- newOrder
						}
					}

					elev.Requests[floor][button].State = common.NONE
				}
			}
		}
	}
}

func costFunction(elev *common.SyncElevator, order common.ButtonEvent) int {
	distance := abs(elev.Floor - order.Floor)
	cost := distance

	if common.ElevatorState(elev.State) == common.MOVING {
		cost += 2
	}

	if (elev.Dirn == common.MotorDirection(common.DIRN_UP) && order.Floor < elev.Floor) ||
		(elev.Dirn == common.MotorDirection(common.DIRN_DOWN) && order.Floor > elev.Floor) {
		cost += 3
	}

	if common.ElevatorState(elev.State) == common.IDLE {
		cost -= 2
	}

	if common.ElevatorState(elev.State) == common.IDLE && elev.Floor == order.Floor {
		cost -= 1
	}

	if order.Button == common.BUTTON_HALL_UP {
		if elev.Requests[order.Floor][common.BUTTON_HALL_DOWN].State == common.ORDER ||
			elev.Requests[order.Floor][common.BUTTON_HALL_DOWN].State == common.CONFIRMED {
			cost += 5
		}
	}
	if order.Button == common.BUTTON_HALL_DOWN {
		if elev.Requests[order.Floor][common.BUTTON_HALL_UP].State == common.ORDER ||
			elev.Requests[order.Floor][common.BUTTON_HALL_UP].State == common.CONFIRMED {
			cost += 5
		}
	}

	activeCount := 0
	for f := 0; f < common.NUM_FLOORS; f++ {
		for b := 0; b < common.NUM_BUTTONS-1; b++ {
			if elev.Requests[f][b].State == common.ORDER || elev.Requests[f][b].State == common.CONFIRMED {
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
