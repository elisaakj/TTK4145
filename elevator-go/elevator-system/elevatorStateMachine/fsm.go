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
			outputDevice.requestButtonLight(button, floor, es.requests[floor][btn])
		}
	}
}

func fsmOnInitBetweenFloors() {
	outputDevice.motorDirection(elevio.MD_Down)
	elevator.dirn = elevio.MD_Down
	elevator.behaviour = EB_Moving
}

func fsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType) {
	//fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %s)\n", btnFloor, elevioButtonToString(btnType))
	//elevatorPrint(elevator)

	switch elevator.behaviour {
	case EB_DoorOpen:
		if requestsShouldClearImmediately(elevator, btnFloor, btnType) {
			timerStart(elevator.doorOpenDurationS)
		} else {
			elevator.requests[btnFloor][btnType] = true
		}
	case EB_Moving:
		elevator.requests[btnFloor][btnType] = true
	case EB_Idle:
		elevator.requests[btnFloor][btnType] = true
		pair := requestsChooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.behaviour = pair.behaviour
		switch pair.behaviour {
		case EB_DoorOpen:
			outputDevice.doorLight(true)
			timerStart(elevator.doorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)
		case EB_Moving:
			outputDevice.motorDirection(elevator.dirn)
		}
	}

	setAllLights(elevator, btnType)
	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func fsmOnFloorArrival(newFloor int, button elevio.ButtonType) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)
	elevatorPrint(elevator)
	elevator.floor = newFloor
	outputDevice.floorIndicator(elevator.floor)

	if elevator.behaviour == EB_Moving && requestsShouldStop(elevator) {
		outputDevice.motorDirection(elevio.MD_Stop)
		outputDevice.doorLight(true)
		elevator = requestsClearAtCurrentFloor(elevator)
		timerStart(elevator.doorOpenDurationS)
		setAllLights(elevator, button)
		elevator.behaviour = EB_DoorOpen
	}

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func fsmOnDoorTimeout(button elevio.ButtonType) {
	fmt.Println("\n\nfsmOnDoorTimeout()")
	elevatorPrint(elevator)

	if elevio.GetObstruction() {
		timerStart(elevator.doorOpenDurationS)
		return
	}

	if elevator.behaviour == EB_DoorOpen {
		pair := requestsChooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.behaviour = pair.behaviour
		switch elevator.behaviour {
		case EB_DoorOpen:
			timerStart(elevator.doorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)
			setAllLights(elevator, button)
		case EB_Moving, EB_Idle:
			outputDevice.doorLight(false)
			outputDevice.motorDirection(elevator.dirn)
		}
	}

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func fsmOnObstruction() {
	fmt.Println("\n\nfsmOnObstruction()")
	elevatorPrint(elevator)

	if elevator.behaviour == EB_DoorOpen {
		outputDevice.doorLight(true) // Keep door open
		timerStop()                  // Stop the timer while obstructed
	}

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func fsmOnObstructionCleared() {
	fmt.Println("\n\nfsmOnObstructionCleared()")
	elevatorPrint(elevator)

	if elevator.behaviour == EB_DoorOpen {
		outputDevice.doorLight(true)           // Keep door open
		timerStart(elevator.doorOpenDurationS) // Restart door timer
	}

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
