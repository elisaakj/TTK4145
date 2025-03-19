package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/sync"
	"fmt"
	// "time"
)

var (
	elevator     Elevator
	outputDevice ElevOutputDevice
)

func setAllLights(es Elevator, button elevio.ButtonType) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			outputDevice.requestButtonLight(button, floor, es.Requests[floor][btn])
		}
	}
}

func fsmOnInitBetweenFloors() {
	startFloor := elevio.GetFloor()
	if startFloor == -1 {
		startFloor = 0
	}
	elevator.Floor = startFloor
	elevator.Dirn = elevio.MD_Down
	elevator.Behaviour = EB_Moving
	elevio.SetMotorDirection(elevio.MD_Down)
}

// btnFloor int, btnType elevio.ButtonType
func fsmOnRequestButtonPress(elevator *Elevator, event elevio.ButtonEvent, ch FsmChannels) {
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
	go sync.broadcastHallCall(*elevator, event, hallCallTx)

	// setting the request to true, and hall light to be switched on
	elevator.Requests[event.Floor][event.Button] = true
	elevio.SetButtonLamp(event.Button, event.Floor, true)

	switch elevator.Behaviour {
	case EB_DoorOpen:
		if elevator.Floor == event.Floor {
			timerStart(elevator.DoorOpenDurationS)
			go func() { ch.OrderComplete <- event.Floor }()
			*elevator = requestsClearAtCurrentFloor(*elevator, int(event.Button))
		}

	//case EB_Moving:
	//	elevator.requests[event.Floor][event.Button] = true
	case EB_Idle:
		choice := requestsChooseDirection(*elevator)
		elevator.Dirn = choice.dirn
		elevator.Behaviour = choice.behaviour

		if elevator.Behaviour == EB_Moving {
			elevio.SetMotorDirection(elevator.Dirn)
			//elevator.behaviour = EB_DoorOpen
			//elevio.SetDoorOpenLamp(true)
			//go func() { ch.OrderComplete <- event.Floor }()
			//*elevator = requestsClearAtCurrentFloor(*elevator)
		} else {
			elevio.SetDoorOpenLamp(true)
			timerStart(elevator.DoorOpenDurationS)
			go func() { ch.OrderComplete <- event.Floor }()
			*elevator = requestsClearAtCurrentFloor(*elevator, int(event.Button))
		}
	}

	ch.Elevator <- *elevator
}

func fsmOnFloorArrival(newFloor int, elevator *Elevator, ch FsmChannels, numButtons int) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)
	elevator.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)

	// elevator.behaviour == EB_Moving && removed
	if requestsShouldStop(*elevator) {
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevator.Behaviour = EB_DoorOpen
		elevio.SetDoorOpenLamp(true)

		*elevator = requestsClearAtCurrentFloor(*elevator, numButtons)
		timerStart(elevator.DoorOpenDurationS)

		go func() { ch.OrderComplete <- newFloor }()
	} else {
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- *elevator
}

func fsmOnDoorTimeout(elevator *Elevator, ch FsmChannels) {
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

	if requestsHere(*elevator) {
		timerStart(elevator.DoorOpenDurationS)
		*elevator = requestsClearAtCurrentFloor(*elevator, N_BUTTONS) //bytte N_BUTTONS
		return
	}

	elevio.SetDoorOpenLamp(false)
	timerStop()

	choice := requestsChooseDirection(*elevator)
	elevator.Dirn = choice.dirn
	elevator.Behaviour = choice.behaviour

	if elevator.Behaviour == EB_Moving {
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- *elevator
}

func fsmOnObstruction(elevator *Elevator, obstruction bool, ch FsmChannels) {
	elevator.Obstructed = obstruction

	if obstruction {
		if elevator.Behaviour == EB_DoorOpen {
			timerStart(elevator.DoorOpenDurationS)
		}
	} else {
		fsmOnObstructionCleared(elevator)
	}

	ch.Elevator <- *elevator
}

func fsmOnObstructionCleared(elevator *Elevator) {
	/*
		if elevator.requests == nil || len(elevator.requests.) == 0 {
			return
		}*/

	if elevator.Behaviour == EB_DoorOpen {
		timerStart(elevator.DoorOpenDurationS)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
