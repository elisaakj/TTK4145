package elevatorStateMachine

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
)

// import (
// 	. "elevator"
// )

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	State 	  ElevatorState
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

func hasRequestsAtCurrentFloor(e Elevator) bool {
	if e.Floor < 0 || e.Floor >= config.NUM_FLOORS {
		return false
	}
	for btn := 0; btn < config.NUM_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func determineNextDirection(e Elevator) DirnBehaviourPair {
	if hasRequestsAtCurrentFloor(e) {
		return DirnBehaviourPair{elevio.DIRN_STOP, DOOR_OPEN}
	}

	switch e.Dirn {
	case elevio.DIRN_UP:
		if hasRequestsAbove(e) {
			return DirnBehaviourPair{elevio.DIRN_UP, MOVING}
		} else if hasRequestsBelow(e) {
			return DirnBehaviourPair{elevio.DIRN_DOWN, MOVING}
		}
	case elevio.DIRN_DOWN:
		if hasRequestsBelow(e) {
			return DirnBehaviourPair{elevio.DIRN_DOWN, MOVING}
		} else if hasRequestsAbove(e) {
			return DirnBehaviourPair{elevio.DIRN_UP, MOVING}
		}
	case elevio.DIRN_STOP:
		if hasRequestsAbove(e) {
			return DirnBehaviourPair{elevio.DIRN_UP, MOVING}
		} else if hasRequestsBelow(e) {
			return DirnBehaviourPair{elevio.DIRN_DOWN, MOVING}
		}
	}
	return DirnBehaviourPair{elevio.DIRN_STOP, IDLE}
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

func clearRequestsImmediately(e Elevator, btnfloor int, btnType elevio.ButtonType) bool {
	switch e.ClearRequestMode {
	case CLEAR_ALL:
		return e.Floor == btnfloor
	case CLEAR_DIRECTION:
		return e.Floor == btnfloor && (e.Dirn == elevio.DIRN_UP && btnType == elevio.BUTTON_HALL_UP || e.Dirn == elevio.DIRN_DOWN && btnType == elevio.BUTTON_HALL_DOWN || e.Dirn == elevio.DIRN_STOP || btnType == elevio.BUTTON_CAB)
	}
	return false
}

func clearRequestsAtCurrentFloor(e Elevator, numButtons int) Elevator {
	for btn := 0; btn < numButtons; btn++ {
		//if e.requests[e.floor][btn] {
		e.Requests[e.Floor][btn] = false
		elevio.SetButtonLamp(elevio.ButtonType(btn), e.Floor, false)
		//}
	}
	return e
}
