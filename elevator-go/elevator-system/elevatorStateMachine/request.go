package elevatorStateMachine

import (
	"Driver-go/elevator-system/common"
	"Driver-go/elevator-system/elevio"
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
		if hasRequestsAbove(e) && e.Requests[e.Floor][common.BUTTON_HALL_UP] {
			e.Requests[e.Floor][common.BUTTON_HALL_UP] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_UP, e.Floor, false)
		} else if hasRequestsBelow(e) && e.Requests[e.Floor][common.BUTTON_HALL_DOWN] {
			e.Requests[e.Floor][common.BUTTON_HALL_DOWN] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_DOWN, e.Floor, false)
		} else if e.Requests[e.Floor][common.BUTTON_HALL_UP] {
			e.Requests[e.Floor][common.BUTTON_HALL_UP] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_UP, e.Floor, false)
		} else if e.Requests[e.Floor][common.BUTTON_HALL_DOWN] {
			e.Requests[e.Floor][common.BUTTON_HALL_DOWN] = false
			elevio.SetButtonLamp(common.BUTTON_HALL_DOWN, e.Floor, false)
		}
	}

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

	if e.Requests[e.Floor][common.BUTTON_CAB] {
		return true
	}

	switch e.Dirn {
	case common.DIRN_UP:
		if e.Floor < common.NUM_FLOORS-1 {
			return e.Requests[e.Floor][common.BUTTON_HALL_UP]
		}
	case common.DIRN_DOWN:
		if e.Floor > 0 {
			return e.Requests[e.Floor][common.BUTTON_HALL_DOWN]
		}
	case common.DIRN_STOP:
		if e.Requests[e.Floor][common.BUTTON_HALL_UP] || e.Requests[e.Floor][common.BUTTON_HALL_DOWN] {
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

	if hasCab || hasUp || hasDown {
		return clearRequestsAtCurrentFloor(e)
	}

	return e
}
