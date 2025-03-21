package elevatorStateMachine

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
	"fmt"
	// "time"
)

var (
	elevator     Elevator
	outputDevice ElevOutputDevice
)

func setAllLights(es Elevator, button elevio.ButtonType) {
	for floor := 0; floor < config.NUM_FLOORS; floor++ {
		for btn := 0; btn < config.NUM_BUTTONS; btn++ {
			outputDevice.RequestButtonLight(button, floor, es.Requests[floor][btn])
		}
	}
}

func initBetweenFloors() {
	startFloor := elevio.GetFloor()
	if startFloor == -1 {
		startFloor = 0
	}
	elevator.Floor = startFloor
	elevator.Dirn = elevio.DIRN_DOWN
	elevator.State = MOVING
	elevio.SetMotorDirection(elevio.DIRN_DOWN)
}

// btnFloor int, btnType elevio.ButtonType
func handleRequestButtonPress(elevator *Elevator, event elevio.ButtonEvent, ch FsmChannels) {
	//fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %s)\n", btnFloor, elevioButtonToString(btnType))
	//elevatorPrint(elevator)

	// if already requested, do nothing
	if elevator.Requests[event.Floor][event.Button] {
		return
	}

	// else it should update the global state
	// broadcast the new state to all elevators
	// when other elevators receive the new state, they should update their own state and broadcast update number

	//  1. Update local state and increment order ID
	elevator.Requests[event.Floor][event.Button] = true
	elevio.SetButtonLamp(event.Button, event.Floor, true)
	elevator.OrderID = (elevator.OrderID + 1) % 1000 //
	//  2. Broadcast new hall call request
	//go syncElev.BroadcastHallCall(*elevator, event, hallCallTx)

	// setting the request to true, and hall light to be switched on
	elevator.Requests[event.Floor][event.Button] = true
	elevio.SetButtonLamp(event.Button, event.Floor, true)

	switch elevator.State {
	case DOOR_OPEN:
		if elevator.Floor == event.Floor {
			timerStart(elevator.DoorOpenDurationS)
			go func() { ch.OrderComplete <- event.Floor }()
			*elevator = clearRequestsAtCurrentFloor(*elevator, int(event.Button))
		}

	//case EB_Moving:
	//	elevator.requests[event.Floor][event.Button] = true
	case IDLE:
		choice := determineNextDirection(*elevator)
		elevator.Dirn = choice.Dirn
		elevator.State = choice.State

		if elevator.State == MOVING {
			elevio.SetMotorDirection(elevator.Dirn)
			//elevator.behaviour = EB_DoorOpen
			//elevio.SetDoorOpenLamp(true)
			//go func() { ch.OrderComplete <- event.Floor }()
			//*elevator = requestsClearAtCurrentFloor(*elevator)
		} else {
			elevio.SetDoorOpenLamp(true)
			timerStart(elevator.DoorOpenDurationS)
			go func() { ch.OrderComplete <- event.Floor }()
			*elevator = clearRequestsAtCurrentFloor(*elevator, int(event.Button))
		}
	}

	ch.Elevator <- *elevator
}

func onFloorArrival(newFloor int, elevator *Elevator, ch FsmChannels, numButtons int) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)
	elevator.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)

	// elevator.behaviour == EB_Moving && removed
	if stopAtCurrentFloor(*elevator) {
		elevio.SetMotorDirection(elevio.DIRN_STOP)
		elevator.State = DOOR_OPEN
		elevio.SetDoorOpenLamp(true)

		*elevator = clearRequestsAtCurrentFloor(*elevator, numButtons)
		timerStart(elevator.DoorOpenDurationS)

		go func() { ch.OrderComplete <- newFloor }()
	} else {
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- *elevator
}

func handleDoorTimeout(elevator *Elevator, ch FsmChannels) {
	//fmt.Println("\n\nfsmOnDoorTimeout()")W
	//elevatorPrint(*elevator)

	/*
		if elevio.GetObstruction() {
			timerStart(elevator.doorOpenDurationS)
			return
		}*/

	//if !timerTimedOut() {
	//	return
	//}

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
	/*
		if elevator.requests == nil || len(elevator.requests.) == 0 {
			return
		}*/

	if elevator.State == DOOR_OPEN {
		timerStart(elevator.DoorOpenDurationS)
	}
}