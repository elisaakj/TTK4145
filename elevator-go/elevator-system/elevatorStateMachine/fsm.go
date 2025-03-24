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
			*elevator = clearRequestsAtCurrentFloor(*elevator)
		}

	case config.IDLE:
		choice := determineNextDirection(*elevator)
		elevator.Dirn = choice.Dirn
		elevator.State = choice.State

		if elevator.State == config.MOVING {
			elevio.SetMotorDirection(elevator.Dirn)
		} else {
			elevio.SetDoorOpenLamp(true)
			*elevator = clearRequestsAtCurrentFloor(*elevator)
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

		*elevator = clearRequestsAtCurrentFloor(*elevator)

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
		*elevator = clearRequestsAtCurrentFloor(*elevator)
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
