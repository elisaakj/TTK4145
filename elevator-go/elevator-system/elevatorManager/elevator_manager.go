package elevatorManager

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
)

type DirnBehaviourPair struct {
	Dirn  config.Direction
	State config.ElevatorState
}

func AssignOrders(elevators []*config.SyncElevator, newOrder elevio.ButtonEvent, excludeID string) int {
	bestElevIndex := -1
	bestScore := 100000

	for i, elev := range elevators {

		if elev.Behave == config.Behaviour(3) { // UNAVAILABLE
			continue
		}

		score := costFunction(elev, newOrder)
		if elev.ID == excludeID {
			score += 8 // discourage assigning to the same elevator again
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

// func costFunction(elev *config.SyncElevator, order elevio.ButtonEvent) int {
// 	distance := abs(elev.Floor - order.Floor)
// 	cost := distance

// 	// Penalize elevators currently moving
// 	if config.ElevatorState(elev.Behave) == config.MOVING {
// 		cost += 2
// 	}

// 	// Penalize elevators going in the wrong direction
// 	if (elev.Dirn == config.Up && order.Floor < elev.Floor) ||
// 		(elev.Dirn == config.Down && order.Floor > elev.Floor) {
// 		cost += 4
// 	}

// 	// Slight bonus for being idle on same floor
// 	if config.ElevatorState(elev.Behave) == config.IDLE && elev.Floor == order.Floor {
// 		cost -= 1
// 	}

// 	return cost
// }

func costFunction(elev *config.SyncElevator, order elevio.ButtonEvent) int {
	distance := abs(elev.Floor - order.Floor)
	cost := distance

	// Penalize if moving
	if config.ElevatorState(elev.Behave) == config.MOVING {
		cost += 2
	}

	// Penalize if moving in wrong direction
	if (elev.Dirn == config.Up && order.Floor < elev.Floor) ||
		(elev.Dirn == config.Down && order.Floor > elev.Floor) {
		cost += 3
	}

	// Reward for being idle (even if not at the same floor)
	if config.ElevatorState(elev.Behave) == config.IDLE {
		cost -= 2
	}

	// Small bonus if idle and at the correct floor
	if config.ElevatorState(elev.Behave) == config.IDLE && elev.Floor == order.Floor {
		cost -= 1
	}

	// NEW: Add penalty if elevator already has multiple active orders
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











/*
func costFunction(elev *config.SyncElevator, order elevio.ButtonEvent) int {
	if elev.Behave != config.Behaviour(config.UNAVAILABLE) {
		e := new(config.SyncElevator)
		*e = *elev
		e.Requests[order.Floor][order.Button].State = config.Confirmed

		duration := 0

		switch e.Behave {
		case config.Behaviour(config.IDLE):
			determineNextDirection(*e)
			if e.Dirn == config.Stop {
				return duration
			}
		case config.Behaviour(config.MOVING):
			duration += config.TRAVEL_TIME / 2
			e.Floor += int(e.Dirn)
		case config.Behaviour(config.DOOR_OPEN):
			duration -= config.DOOR_OPEN_DURATION / 2
		}

		for {
			if stopAtCurrentFloor(*e) {
				clearRequestsAtCurrentFloor(e)
				duration += config.DOOR_OPEN_DURATION
				determineNextDirection(*e)
				if e.Dirn == config.Stop {
					return duration
				}
			}
			e.Floor += int(e.Dirn)
			duration += config.TRAVEL_TIME
		}
	}
	return 999
}

func determineNextDirection(e config.SyncElevator) DirnBehaviourPair {

	switch e.Dirn {
	case config.Up:
		if hasRequestsAbove(e) {
			return DirnBehaviourPair{config.Up, config.MOVING}
		} else if hasRequestsBelow(e) {
			return DirnBehaviourPair{config.Down, config.MOVING}
		}
	case config.Down:
		if hasRequestsBelow(e) {
			return DirnBehaviourPair{config.Down, config.MOVING}
		} else if hasRequestsAbove(e) {
			return DirnBehaviourPair{config.Up, config.MOVING}
		}
	case config.Stop:
		if hasRequestsAbove(e) {
			return DirnBehaviourPair{config.Up, config.MOVING}
		} else if hasRequestsBelow(e) {
			return DirnBehaviourPair{config.Down, config.MOVING}
		}
	}
	return DirnBehaviourPair{config.Stop, config.IDLE}
}

func stopAtCurrentFloor(e config.SyncElevator) bool {
	if e.Floor < 0 || e.Floor >= len(e.Requests) {
		return false
	}

	if e.Requests[e.Floor][elevio.BUTTON_CAB].State == config.Confirmed ||
		e.Requests[e.Floor][elevio.BUTTON_HALL_UP].State == config.Confirmed ||
		e.Requests[e.Floor][elevio.BUTTON_HALL_DOWN].State == config.Confirmed {
		return true
	}

	switch e.Dirn {
	case config.Down:
		return !hasRequestsBelow(e)
	case config.Up:
		return !hasRequestsAbove(e)
	case config.Stop:
		return true
	}
	return false
}

func clearRequestsAtCurrentFloor(elev *config.SyncElevator) {
	elev.Requests[elev.Floor][int(elevio.BUTTON_CAB)].State = config.None
	switch {
	case elev.Dirn == config.Up:
		elev.Requests[elev.Floor][int(config.Up)].State = config.None
		if !hasRequestsAbove(*elev) {
			elev.Requests[elev.Floor][int(elevio.BUTTON_HALL_DOWN)].State = config.None
		}
	case elev.Dirn == config.Down:
		elev.Requests[elev.Floor][int(elevio.BUTTON_HALL_DOWN)].State = config.None
		if !hasRequestsBelow(*elev) {
			elev.Requests[elev.Floor][int(config.Up)].State = config.None
		}
	}
}

func hasRequestsAbove(e config.SyncElevator) bool {
	for f := e.Floor + 1; f < config.NUM_FLOORS; f++ {
		for btn := 0; btn < config.NUM_BUTTONS; btn++ {
			if e.Requests[f][btn].State == config.Confirmed {
				return true
			}
		}
	}
	return false
}

func hasRequestsBelow(e config.SyncElevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < config.NUM_BUTTONS; btn++ {
			if e.Requests[f][btn].State == config.Confirmed {
				return true
			}
		}
	}
	return false
}
*/
