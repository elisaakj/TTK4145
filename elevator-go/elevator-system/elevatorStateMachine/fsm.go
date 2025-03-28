package elevatorStateMachine

import (
	"Driver-go/elevator-system/common"
	"Driver-go/elevator-system/elevio"
)

func handleRequestButtonPress(elevator *common.Elevator, event common.ButtonEvent, ch common.FsmChannels) {
	// if already requested, do nothing
	if elevator.Requests[event.Floor][event.Button] {
		return
	}

	elevator.Requests[event.Floor][event.Button] = true
	elevio.SetButtonLamp(event.Button, event.Floor, true)
	// elevator.OrderID = (elevator.OrderID + 1) % 1000 need to implement this probably

	switch elevator.State {
	case common.DOOR_OPEN:
		if elevator.Floor == event.Floor {
			*elevator = handleRequestsAndMaybeReverse(*elevator)
		}

	case common.IDLE:
		choice := determineNextDirection(*elevator)
		elevator.Dirn = choice.Dirn
		elevator.State = choice.State

		if elevator.State == common.MOVING {
			elevio.SetMotorDirection(elevator.Dirn)
		} else {
			elevio.SetDoorOpenLamp(true)
			*elevator = clearHallRequestInDirection(*elevator)
		}
	}

	ch.Elevator <- *elevator
}

func onFloorArrival(newFloor int, elevator *common.Elevator, ch common.FsmChannels) {
	elevator.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)

	if stopAtCurrentFloor(*elevator) {
		elevio.SetMotorDirection(common.DIRN_STOP)
		elevator.State = common.DOOR_OPEN
		elevio.SetDoorOpenLamp(true)

		*elevator = handleRequestsAndMaybeReverse(*elevator)

	} else {
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- *elevator
}

func handleDoorTimeout(elevator *common.Elevator, ch common.FsmChannels) {
	if elevator.Obstructed {
		return
	}

	if ifHaveRequestInSameDirection(*elevator) {
		*elevator = handleRequestsAndMaybeReverse(*elevator)
		return
	}

	elevio.SetDoorOpenLamp(false)

	choice := determineNextDirection(*elevator)
	elevator.Dirn = choice.Dirn
	elevator.State = choice.State

	if elevator.State == common.MOVING {
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- *elevator
}

func handleObstruction(elevator *common.Elevator, obstruction bool, ch common.FsmChannels) {
	elevator.Obstructed = obstruction

	ch.Elevator <- *elevator
}
