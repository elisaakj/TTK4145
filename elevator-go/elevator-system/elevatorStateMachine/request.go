package elevatorStateMachine

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
	"fmt"
)

type DirnBehaviourPair struct {
	Dirn  elevio.MotorDirection
	State config.ElevatorState
}

func hasRequestsAbove(e Elevator) bool {
	for f := e.Floor + 1; f < config.NUM_FLOORS; f++ {
		for btn := 0; btn < config.NUM_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func hasRequestsBelow(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < config.NUM_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}


func determineNextDirection(e Elevator) DirnBehaviourPair {
	if ifHaveRequestInSameDirection(e) {
		return DirnBehaviourPair{elevio.DIRN_STOP, config.DOOR_OPEN}
	}

	switch e.Dirn {
	case elevio.DIRN_UP:
		if hasRequestsAbove(e) {
			return DirnBehaviourPair{elevio.DIRN_UP, config.MOVING}
		} else if hasRequestsBelow(e) {
			return DirnBehaviourPair{elevio.DIRN_DOWN, config.MOVING}
		}
	case elevio.DIRN_DOWN:
		if hasRequestsBelow(e) {
			return DirnBehaviourPair{elevio.DIRN_DOWN, config.MOVING}
		} else if hasRequestsAbove(e) {
			return DirnBehaviourPair{elevio.DIRN_UP, config.MOVING}
		}
	case elevio.DIRN_STOP:
		if hasRequestsAbove(e) {
			return DirnBehaviourPair{elevio.DIRN_UP, config.MOVING}
		} else if hasRequestsBelow(e) {
			return DirnBehaviourPair{elevio.DIRN_DOWN, config.MOVING}
		}
	}
	return DirnBehaviourPair{elevio.DIRN_STOP, config.IDLE}
}

func stopAtCurrentFloor(e Elevator) bool {
	if e.Requests[e.Floor][elevio.BUTTON_CAB] || e.Requests[e.Floor][elevio.BUTTON_HALL_UP] || e.Requests[e.Floor][elevio.BUTTON_HALL_DOWN] {
		return true
	}

	switch e.Dirn {
	case elevio.DIRN_DOWN:
		return !hasRequestsBelow(e)
	case elevio.DIRN_UP:
		return !hasRequestsAbove(e)
	case elevio.DIRN_STOP:
		return true
	}
	return false
}



func clearRequestsAtCurrentFloor(e Elevator) Elevator {
	for btn := 0; btn < config.NUM_BUTTONS; btn++ {
		e.Requests[e.Floor][btn] = false
		elevio.SetButtonLamp(elevio.ButtonType(btn), e.Floor, false)
	}
	return e
}


func clearHallRequestInDirection(e Elevator) Elevator {
	switch e.Dirn {
	case elevio.DIRN_UP:
		if e.Requests[e.Floor][elevio.BUTTON_HALL_UP] {
			e.Requests[e.Floor][elevio.BUTTON_HALL_UP] = false
			elevio.SetButtonLamp(elevio.BUTTON_HALL_UP, e.Floor, false)
		}
	case elevio.DIRN_DOWN:
		if e.Requests[e.Floor][elevio.BUTTON_HALL_DOWN] {
			e.Requests[e.Floor][elevio.BUTTON_HALL_DOWN] = false
			elevio.SetButtonLamp(elevio.BUTTON_HALL_DOWN, e.Floor, false)
		}
	case elevio.DIRN_STOP:
		// When idle, clear both possible hall calls at this floor
		if e.Requests[e.Floor][elevio.BUTTON_HALL_UP] {
			e.Requests[e.Floor][elevio.BUTTON_HALL_UP] = false
			elevio.SetButtonLamp(elevio.BUTTON_HALL_UP, e.Floor, false)
		}
		if e.Requests[e.Floor][elevio.BUTTON_HALL_DOWN] {
			e.Requests[e.Floor][elevio.BUTTON_HALL_DOWN] = false
			elevio.SetButtonLamp(elevio.BUTTON_HALL_DOWN, e.Floor, false)
		}
	}

	// Always clear the cab call at this floor
	if e.Requests[e.Floor][elevio.BUTTON_CAB] {
		e.Requests[e.Floor][elevio.BUTTON_CAB] = false
		elevio.SetButtonLamp(elevio.BUTTON_CAB, e.Floor, false)
	}

	return e
}



func ifHaveRequestInSameDirection(e Elevator) bool {
	if e.Floor < 0 || e.Floor >= config.NUM_FLOORS {
		return false
	}

	// Always honor a CAB call
	if e.Requests[e.Floor][elevio.BUTTON_CAB] {
		return true
	}

	switch e.Dirn {
	case elevio.DIRN_UP:
		// Don't check up button on top floor
		if e.Floor < config.NUM_FLOORS-1 {
			return e.Requests[e.Floor][elevio.BUTTON_HALL_UP]
		}
	case elevio.DIRN_DOWN:
		// Don't check down button on ground floor
		if e.Floor > 0 {
			return e.Requests[e.Floor][elevio.BUTTON_HALL_DOWN]
		}
	case elevio.DIRN_STOP:
		// If idle, respond to any hall call
		if e.Requests[e.Floor][elevio.BUTTON_HALL_UP] || e.Requests[e.Floor][elevio.BUTTON_HALL_DOWN] {
			fmt.Printf("[CHECK] ifHaveRequestInSameDirection: floor=%d dirn=%v up=%v\n", e.Floor, e.Dirn, e.Requests[e.Floor][elevio.BUTTON_HALL_UP]) //DEBUG
			return true
		}
	}

	return false
}

func handleRequestsAndMaybeReverse(e Elevator) Elevator {
	floor := e.Floor
	hasCab := e.Requests[floor][elevio.BUTTON_CAB]
	hasUp := e.Requests[floor][elevio.BUTTON_HALL_UP]
	hasDown := e.Requests[floor][elevio.BUTTON_HALL_DOWN]

	switch e.Dirn {
	case elevio.DIRN_UP:
		if hasCab && !hasUp && hasDown {
			e.Dirn = elevio.DIRN_DOWN
			return clearRequestsAtCurrentFloor(e)
		}
	case elevio.DIRN_DOWN:
		if hasCab && !hasDown && hasUp {
			e.Dirn = elevio.DIRN_UP
			return clearRequestsAtCurrentFloor(e)
		}
	}

	// Clear all calls at current floor if no direction-specific call exists
	if hasCab || hasUp || hasDown {
		return clearRequestsAtCurrentFloor(e)
	}

	// Fallback if nothing to clear (safety net)
	return e
}


// func handleRequestsAndMaybeReverse(e Elevator) Elevator {
	
// 	floor := e.Floor
// 	hasCab := e.Requests[floor][elevio.BUTTON_CAB]
// 	hasUp := e.Requests[floor][elevio.BUTTON_HALL_UP]
// 	hasDown := e.Requests[floor][elevio.BUTTON_HALL_DOWN]

// 	switch e.Dirn {
// 	case elevio.DIRN_UP:
// 		if hasCab && !hasUp && hasDown {
// 			e.Dirn = elevio.DIRN_DOWN
// 			return clearRequestsAtCurrentFloor(e)
// 		}
// 	case elevio.DIRN_DOWN:
// 		if hasCab && !hasDown && hasUp {
// 			e.Dirn = elevio.DIRN_UP
// 			return clearRequestsAtCurrentFloor(e)
// 		}
// 	case elevio.DIRN_STOP:
// 		// If idle, just clear all valid calls
// 		if hasCab || hasUp || hasDown {
// 			return clearRequestsAtCurrentFloor(e)
// 		}
// 	}

// 	// Default: just clear what's valid
// 	if ifHaveRequestInSameDirection(e) {
// 		return clearHallRequestInDirection(e)
// 	}

// 	return e
// }





// // To help reverse the elevator direction 

// func oppositeDirection(dir elevio.MotorDirection) elevio.MotorDirection {
// 	if dir == elevio.DIRN_UP {
// 		return elevio.DIRN_DOWN
// 	} else if dir == elevio.DIRN_DOWN {
// 		return elevio.DIRN_UP
// 	}
// 	return elevio.DIRN_STOP
// }


/*
func clearRequestsImmediately(e Elevator, btnfloor int, btnType elevio.ButtonType) bool {
	switch e.ClearRequestMode {
	case CLEAR_ALL:
		return e.Floor == btnfloor
	case CLEAR_DIRECTION:
		return e.Floor == btnfloor && (e.Dirn == elevio.DIRN_UP && btnType == elevio.BUTTON_HALL_UP || e.Dirn == elevio.DIRN_DOWN && btnType == elevio.BUTTON_HALL_DOWN || e.Dirn == elevio.DIRN_STOP || btnType == elevio.BUTTON_CAB)
	}
	return false
}*/


// Check if the cab requests are opposite of hall requests, if so, reverse the direction, clear both lights and return the elevator
// Otherwise, continues as normal


// func hasRequestsAtCurrentFloor(e Elevator) bool {
// 	if e.Floor < 0 || e.Floor >= config.NUM_FLOORS {
// 		return false
// 	}
// 	for btn := 0; btn < config.NUM_BUTTONS; btn++ {
// 		if e.Requests[e.Floor][btn] {
// 			return true
// 		}
// 	}
// 	return false
// }