package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"fmt"
	// "time"
)

var (
	elevator     Elevator
	outputDevice ElevOutputDevice
)

/*
func init() {
	elevator = elevatorUninitialized()
	// Simulating config loading
	elevator.config.doorOpenDurationS = 3.0 // Example default value
	elevator.config.clearRequestVariant = CV_All
	outputDevice = elevioGetOutputDevice()
}*/

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

	if elevator.Requests[event.Floor][event.Button] {
		return
	}

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

	/*
		if elevator.dirn == elevio.MD_Stop {
			elevator.behaviour = EB_Idle
		} else {
			elevator.behaviour = EB_Moving
			elevio.SetMotorDirection(elevator.dirn)
		}

			if elevator.behaviour == EB_DoorOpen {
				pair := requestsChooseDirection(elevator)
				elevator.dirn = pair.dirn
				elevator.behaviour = pair.behaviour
				switch elevator.behaviour {
				case EB_DoorOpen:
					timerStart(elevator.doorOpenDurationS)
					elevator = requestsClearAtCurrentFloor(elevator)
					setAllLights(elevator, event.Button)
				case EB_Moving, EB_Idle:
					outputDevice.doorLight(false)
					outputDevice.motorDirection(elevator.dirn)
				}
			}
	*/

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
