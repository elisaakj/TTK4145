package elevatorStateMachine

import "Driver-go/elevator-system/elevio"

// import (
// 	. "elevator"
// )

type DirnBehaviourPair struct {
	dirn      elevio.MotorDirection
	behaviour ElevatorBehaviour
}

func requestsAbove(e Elevator) bool {
	for f := e.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e Elevator) bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.requests[e.floor][btn] {
			return true
		}
	}
	return false
}

func requestsChooseDirection(e Elevator) DirnBehaviourPair {
	switch e.dirn {
	case elevio.MD_Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		}
	case elevio.MD_Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		}
	case elevio.MD_Stop:
		if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Stop, EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		}
	}
	return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
}

func requestsShouldStop(e Elevator) bool {
	switch e.dirn {
	case elevio.MD_Down:
		return e.requests[e.floor][elevio.BT_HallDown] || e.requests[e.floor][elevio.BT_Cab] || !requestsBelow(e)
	case elevio.MD_Up:
		return e.requests[e.floor][elevio.BT_HallUp] || e.requests[e.floor][elevio.BT_Cab] || !requestsAbove(e)
	case elevio.MD_Stop:
		return true
	}
	return false
}

func requestsShouldClearImmediately(e Elevator, btnfloor int, btnType elevio.ButtonType) bool {
	switch e.clearRequestVariant {
	case CV_All:
		return e.floor == btnfloor
	case CV_InDirn:
		return e.floor == btnfloor && (e.dirn == elevio.MD_Up && btnType == elevio.BT_HallUp || e.dirn == elevio.MD_Down && btnType == elevio.BT_HallDown || e.dirn == elevio.MD_Stop || btnType == elevio.BT_Cab)
	}
	return false
}

func requestsClearAtCurrentFloor(e Elevator) Elevator {
	switch e.clearRequestVariant {
	case CV_All:
		for btn := 0; btn < N_BUTTONS; btn++ {
			e.requests[e.floor][btn] = false
		}
	case CV_InDirn:
		e.requests[e.floor][elevio.BT_Cab] = false
		switch e.dirn {
		case elevio.MD_Up:
			if !requestsAbove(e) && !e.requests[e.floor][elevio.BT_HallUp] {
				e.requests[e.floor][elevio.BT_HallDown] = false
			}
			e.requests[e.floor][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			if !requestsBelow(e) && !e.requests[e.floor][elevio.BT_HallDown] {
				e.requests[e.floor][elevio.BT_HallUp] = false
			}
			e.requests[e.floor][elevio.BT_HallDown] = false
		default:
			e.requests[e.floor][elevio.BT_HallUp] = false
			e.requests[e.floor][elevio.BT_HallDown] = false
		}
	}
	return e
}
