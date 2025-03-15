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
	for f := e.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e Elevator) bool {
	if e.Floor < 0 || e.Floor >= N_FLOORS {
		return false
	}
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func requestsChooseDirection(e Elevator) DirnBehaviourPair {
	if requestsHere(e) {
		return DirnBehaviourPair{elevio.MD_Stop, EB_DoorOpen}
	}

	switch e.Dirn {
	case elevio.MD_Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		}
	case elevio.MD_Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		}
	case elevio.MD_Stop:
		if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		}
	}
	return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
}

func requestsShouldStop(e Elevator) bool {
	if e.Requests[e.Floor][elevio.BT_Cab] || e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_HallDown] {
		return true
	}

	switch e.Dirn {
	case elevio.MD_Down:
		return !requestsBelow(e)
	case elevio.MD_Up:
		return !requestsAbove(e)
	case elevio.MD_Stop:
		return true
	}
	return false
}

func requestsShouldClearImmediately(e Elevator, btnfloor int, btnType elevio.ButtonType) bool {
	switch e.ClearRequestVariant {
	case CV_All:
		return e.Floor == btnfloor
	case CV_InDirn:
		return e.Floor == btnfloor && (e.Dirn == elevio.MD_Up && btnType == elevio.BT_HallUp || e.Dirn == elevio.MD_Down && btnType == elevio.BT_HallDown || e.Dirn == elevio.MD_Stop || btnType == elevio.BT_Cab)
	}
	return false
}

func requestsClearAtCurrentFloor(e Elevator, numButtons int) Elevator {
	for btn := 0; btn < numButtons; btn++ {
		//if e.requests[e.floor][btn] {
		e.Requests[e.Floor][btn] = false
		elevio.SetButtonLamp(elevio.ButtonType(btn), e.Floor, false)
		//}
	}
	return e
	/*
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
	*/
}
