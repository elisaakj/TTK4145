package elevatorStateMachine

import (
	"Driver-go/elevator-system/common"
	"Driver-go/elevator-system/elevio"
	"fmt"
)


func hasRequestsAbove(e common.Elevator) bool {
	for f := e.Floor + 1; f < common.NUM_FLOORS; f++ {
		for btn := 0; btn < common.NUM_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func hasRequestsBelow(e common.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < common.NUM_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func determineNextDirection(e common.Elevator) common.DirnBehaviourPair {
	if ifHaveRequestInSameDirection(e) {
		return common.DirnBehaviourPair{common.DIRN_STOP, common.DOOR_OPEN}
	}

	switch e.Dirn {
	case common.DIRN_UP:
		if hasRequestsAbove(e) {
			return common.DirnBehaviourPair{common.DIRN_UP, common.MOVING}
		} else if hasRequestsBelow(e) {
			return common.DirnBehaviourPair{common.DIRN_DOWN, common.MOVING}
		}
	case common.DIRN_DOWN:
		if hasRequestsBelow(e) {
			return common.DirnBehaviourPair{common.DIRN_DOWN, common.MOVING}
		} else if hasRequestsAbove(e) {
			return common.DirnBehaviourPair{common.DIRN_UP, common.MOVING}
		}
	case common.DIRN_STOP:
		if hasRequestsAbove(e) {
			return common.DirnBehaviourPair{common.DIRN_UP, common.MOVING}
		} else if hasRequestsBelow(e) {
			return common.DirnBehaviourPair{common.DIRN_DOWN, common.MOVING}
		}
	}
	return common.DirnBehaviourPair{common.DIRN_STOP, common.IDLE}
}

func stopAtCurrentFloor(e common.Elevator) bool {
	if e.Floor < 0 || e.Floor >= common.NUM_FLOORS {
		return false
	}

	if e.Requests[e.Floor][common.BUTTON_CAB] {
		return true
	}

	switch e.Dirn {
	case common.DIRN_DOWN:
		return !hasRequestsBelow(e)
	case common.DIRN_UP:
		return !hasRequestsAbove(e)
	case common.DIRN_STOP:
		return true
	}
	return false
}

func clearRequestsAtCurrentFloor(e common.Elevator) common.Elevator {
	for btn := 0; btn < common.NUM_BUTTONS; btn++ {
		e.Requests[e.Floor][btn] = false
		elevio.SetButtonLamp(common.ButtonType(btn), e.Floor, false)
	}
	return e
}

func clearHallRequestInDirection(e common.Elevator) common.Elevator {
	switch e.Dirn {
	case common.DIRN_UP:
		if e.Requests[e.Floor][common.BUTTON_HALL_UP] {
			e.Requests[e.Floor][common.BUTTON_HALL_UP] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_UP, e.Floor, false)
		}
	case common.DIRN_DOWN:
		if e.Requests[e.Floor][common.BUTTON_HALL_DOWN] {
			e.Requests[e.Floor][common.BUTTON_HALL_DOWN] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_DOWN, e.Floor, false)
		}
	case common.DIRN_STOP:
		// Clear only one hall call based on future direction preference
		if hasRequestsAbove(e) && e.Requests[e.Floor][common.BUTTON_HALL_UP] {
			e.Requests[e.Floor][common.BUTTON_HALL_UP] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_UP, e.Floor, false)
		} else if hasRequestsBelow(e) && e.Requests[e.Floor][common.BUTTON_HALL_DOWN] {
			e.Requests[e.Floor][common.BUTTON_HALL_DOWN] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_DOWN, e.Floor, false)
		} else if e.Requests[e.Floor][common.BUTTON_HALL_UP] {
			// fallback
			e.Requests[e.Floor][common.BUTTON_HALL_UP] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_UP, e.Floor, false)
		} else if e.Requests[e.Floor][common.BUTTON_HALL_DOWN] {
			e.Requests[e.Floor][common.BUTTON_HALL_DOWN] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_DOWN, e.Floor, false)
		}
	}

	// Always clear the cab call at this floor
	if e.Requests[e.Floor][common.BUTTON_CAB] {
		e.Requests[e.Floor][common.BUTTON_CAB] = false
		elevio.SetButtonLamp(common.BUTTON_CAB, e.Floor, false)
	}

	return e
}

func ifHaveRequestInSameDirection(e common.Elevator) bool {
	if e.Floor < 0 || e.Floor >= common.NUM_FLOORS {
		return false
	}

	// Always honor a CAB call
	if e.Requests[e.Floor][common.BUTTON_CAB] {
		return true
	}

	switch e.Dirn {
	case common.DIRN_UP:
		// Don't check up button on top floor
		if e.Floor < common.NUM_FLOORS-1 {
			return e.Requests[e.Floor][common.BUTTON_HALL_UP]
		}
	case common.DIRN_DOWN:
		// Don't check down button on ground floor
		if e.Floor > 0 {
			return e.Requests[e.Floor][common.BUTTON_HALL_DOWN]
		}
	case common.DIRN_STOP:
		// If idle, respond to any hall call
		if e.Requests[e.Floor][common.BUTTON_HALL_UP] || e.Requests[e.Floor][common.BUTTON_HALL_DOWN] {
			fmt.Printf("[CHECK] ifHaveRequestInSameDirection: floor=%d dirn=%v up=%v\n", e.Floor, e.Dirn, e.Requests[e.Floor][common.BUTTON_HALL_UP]) //DEBUG
			return true
		}
	}

	return false
}

func handleRequestsAndMaybeReverse(e common.Elevator) common.Elevator {
	floor := e.Floor
	hasCab := e.Requests[floor][common.BUTTON_CAB]
	hasUp := e.Requests[floor][common.BUTTON_HALL_UP]
	hasDown := e.Requests[floor][common.BUTTON_HALL_DOWN]

	switch e.Dirn {
	case common.DIRN_UP:
		if hasCab && !hasUp && hasDown {
			e.Dirn = common.DIRN_DOWN
			return clearRequestsAtCurrentFloor(e)
		}
	case common.DIRN_DOWN:
		if hasCab && !hasDown && hasUp {
			e.Dirn = common.DIRN_UP
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

// func handleRequestsAndMaybeReverse(e common.Elevator) common.Elevator {

// 	floor := e.Floor
// 	hasCab := e.Requests[floor][common.BUTTON_CAB]
// 	hasUp := e.Requests[floor][common.BUTTON_HALL_UP]
// 	hasDown := e.Requests[floor][common.BUTTON_HALL_DOWN]

// 	switch e.Dirn {
// 	case common.DIRN_UP:
// 		if hasCab && !hasUp && hasDown {
// 			e.Dirn = common.DIRN_DOWN
// 			return clearRequestsAtCurrentFloor(e)
// 		}
// 	case common.DIRN_DOWN:
// 		if hasCab && !hasDown && hasUp {
// 			e.Dirn = common.DIRN_UP
// 			return clearRequestsAtCurrentFloor(e)
// 		}
// 	case common.DIRN_STOP:
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

// // To help reverse the common.elevator direction

// func oppositeDirection(dir elevio.MotorDirection) elevio.MotorDirection {
// 	if dir == common.DIRN_UP {
// 		return common.DIRN_DOWN
// 	} else if dir == common.DIRN_DOWN {
// 		return common.DIRN_UP
// 	}
// 	return common.DIRN_STOP
// }

/*
func clearRequestsImmediately(e common.Elevator, btnfloor int, btnType common.ButtonType) bool {
	switch e.ClearRequestMode {
	case CLEAR_ALL:
		return e.Floor == btnfloor
	case CLEAR_DIRECTION:
		return e.Floor == btnfloor && (e.Dirn == common.DIRN_UP && btnType == common.BUTTON_HALL_UP || e.Dirn == common.DIRN_DOWN && btnType == common.BUTTON_HALL_DOWN || e.Dirn == common.DIRN_STOP || btnType == common.BUTTON_CAB)
	}
	return false
}*/

// Check if the cab requests are opposite of hall requests, if so, reverse the direction, clear both lights and return the common.elevator
// Otherwise, continues as normal

// func hasRequestsAtCurrentFloor(e common.Elevator) bool {
// 	if e.Floor < 0 || e.Floor >= common.NUM_FLOORS {
// 		return false
// 	}
// 	for btn := 0; btn < common.NUM_BUTTONS; btn++ {
// 		if e.Requests[e.Floor][btn] {
// 			return true
// 		}
// 	}
// 	return false
// }
