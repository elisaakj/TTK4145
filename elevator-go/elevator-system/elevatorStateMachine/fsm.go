package elevatorStateMachine

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
	"fmt"
	// "time"
)

// btnFloor int, btnType elevio.ButtonType
func handleRequestButtonPress(elevator *Elevator, event elevio.ButtonEvent, ch FsmChannels) {
	//fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %s)\n", btnFloor, elevioButtonToString(btnType))
	//elevatorPrint(elevator)

	// if already requested, do nothing
	if elevator.Requests[event.Floor][event.Button] {
		return
	}

	elevator.Requests[event.Floor][event.Button] = true
	elevio.SetButtonLamp(event.Button, event.Floor, true)
	// elevator.OrderID = (elevator.OrderID + 1) % 1000 need to implement this probably

	elevator.Requests[event.Floor][event.Button] = true
	elevio.SetButtonLamp(event.Button, event.Floor, true)

	switch elevator.State {
	case DOOR_OPEN:
		if elevator.Floor == event.Floor {
			timerStart(elevator.DoorOpenDurationS)
			//go func() { ch.OrderComplete <- event.Floor }()
			*elevator = clearRequestsAtCurrentFloor(*elevator, int(event.Button))
		}

	case IDLE:
		choice := determineNextDirection(*elevator)
		elevator.Dirn = choice.Dirn
		elevator.State = choice.State

		if elevator.State == MOVING {
			elevio.SetMotorDirection(elevator.Dirn)
		} else {
			elevio.SetDoorOpenLamp(true)
			timerStart(elevator.DoorOpenDurationS)
			//go func() { ch.OrderComplete <- event.Floor }()
			*elevator = clearRequestsAtCurrentFloor(*elevator, int(event.Button))
		}
	}

	ch.Elevator <- *elevator
}

func onFloorArrival(newFloor int, elevator *Elevator, ch FsmChannels, numButtons int) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)
	elevator.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)

	if stopAtCurrentFloor(*elevator) {
		elevio.SetMotorDirection(elevio.DIRN_STOP)
		elevator.State = DOOR_OPEN
		elevio.SetDoorOpenLamp(true)

		*elevator = clearRequestsAtCurrentFloor(*elevator, numButtons)
		timerStart(elevator.DoorOpenDurationS)

		//go func() { ch.OrderComplete <- newFloor }()
	} else {
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- *elevator
}

func handleDoorTimeout(elevator *Elevator, ch FsmChannels) {
	if elevator.Obstructed {
		timerStart(elevator.DoorOpenDurationS)
		return
	}

	if hasRequestsAtCurrentFloor(*elevator) {
		timerStart(elevator.DoorOpenDurationS)
		*elevator = clearRequestsAtCurrentFloor(*elevator, config.NUM_BUTTONS)
		return
	}

	elevio.SetDoorOpenLamp(false)
	timerStop()

	choice := determineNextDirection(*elevator)
	elevator.Dirn = choice.Dirn
	elevator.State = choice.State

	if elevator.State == MOVING {
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- *elevator
}

func handleObstruction(elevator *Elevator, obstruction bool, ch FsmChannels) {
	elevator.Obstructed = obstruction

	if obstruction {
		if elevator.State == DOOR_OPEN {
			timerStart(elevator.DoorOpenDurationS)
		}
	} else {
		handleObstructionCleared(elevator)
	}

	ch.Elevator <- *elevator
}

func handleObstructionCleared(elevator *Elevator) {

	if elevator.State == DOOR_OPEN {
		timerStart(elevator.DoorOpenDurationS)
	}
}
