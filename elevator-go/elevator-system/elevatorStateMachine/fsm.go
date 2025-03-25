package elevatorStateMachine

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
)

type FsmChannels struct {
	Elevator       chan Elevator
	NewOrder       chan elevio.ButtonEvent
	ArrivedAtFloor chan int
	Obstruction    chan bool
}

func handleRequestButtonPress(elevator *Elevator, event elevio.ButtonEvent, ch FsmChannels) {

	// if already requested, do nothing
	if elevator.Requests[event.Floor][event.Button] {
		return
	}

	elevator.Requests[event.Floor][event.Button] = true
	elevio.SetButtonLamp(event.Button, event.Floor, true)
	// elevator.OrderID = (elevator.OrderID + 1) % 1000 need to implement this probably

	switch elevator.State {
	case config.DOOR_OPEN:
		if elevator.Floor == event.Floor {
			*elevator = clearHallRequestInDirection(*elevator)
		}

	case config.IDLE:
		choice := determineNextDirection(*elevator)
		elevator.Dirn = choice.Dirn
		elevator.State = choice.State

		if elevator.State == config.MOVING {
			elevio.SetMotorDirection(elevator.Dirn)
		} else {
			elevio.SetDoorOpenLamp(true)
			*elevator = clearHallRequestInDirection(*elevator)
		}
	}

	ch.Elevator <- *elevator
}

func onFloorArrival(newFloor int, elevator *Elevator, ch FsmChannels) {
	elevator.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)

	if stopAtCurrentFloor(*elevator) {
		elevio.SetMotorDirection(elevio.DIRN_STOP)
		elevator.State = config.DOOR_OPEN
		elevio.SetDoorOpenLamp(true)

		*elevator = clearHallRequestInDirection(*elevator)

	} else {
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- *elevator
}

func handleDoorTimeout(elevator *Elevator, ch FsmChannels) {
	if elevator.Obstructed {
		return
	}


	if hasRequestsAtCurrentFloor(*elevator) {
		*elevator = clearHallRequestInDirection(*elevator)
		return
	}

	elevio.SetDoorOpenLamp(false)

	choice := determineNextDirection(*elevator)
	elevator.Dirn = choice.Dirn
	elevator.State = choice.State

	if elevator.State == config.MOVING {
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- *elevator
}

func handleObstruction(elevator *Elevator, obstruction bool, ch FsmChannels) {
	elevator.Obstructed = obstruction

	ch.Elevator <- *elevator
}



// Check if the cab requests are opposite of hall requests, if so, reverse the direction, clear both lights and return the elevator
// Otherwise, continues as normal

func handleRequestsAndMaybeReverse(elevator Elevator) Elevator {
	floor := elevator.Floor
	hasCab := elevator.Requests[floor][elevio.BUTTON_CAB]
	hasUp := elevator.Requests[floor][elevio.BUTTON_HALL_UP]
	hasDown := elevator.Requests[floor][elevio.BUTTON_HALL_DOWN]

	if hasRequestsAtCurrentFloor(elevator) {
		if hasCab && ((elevator.Dirn == elevio.DIRN_UP && !hasUp) || (elevator.Dirn == elevio.DIRN_DOWN && !hasDown)) {
			elevator.Dirn = oppositeDirection(elevator.Dirn)
			return clearRequestsAtCurrentFloor(elevator)
		}
		return clearHallRequestInDirection(elevator)
	}
	return elevator
}

// To help reverse the elevator direction 

func oppositeDirection(dir elevio.MotorDirection) elevio.MotorDirection {
	if dir == elevio.DIRN_UP {
		return elevio.DIRN_DOWN
	} else if dir == elevio.DIRN_DOWN {
		return elevio.DIRN_UP
	}
	return elevio.DIRN_STOP
}
