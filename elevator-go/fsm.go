package main

import (
	"fmt"
	// "time"
)

var (
	elevator     Elevator
	outputDevice ElevOutputDevice
)

func init() {
	elevator = elevatorUninitialized()
	// Simulating config loading
	elevator.config.doorOpenDurationS = 3.0 // Example default value
	elevator.config.clearRequestVariant = CV_All
	outputDevice = elevioGetOutputDevice()
}

func setAllLights(es Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			outputDevice.requestButtonLight(floor, Button(btn), boolToInt(es.requests[floor][btn]))
		}
	}
}

func fsmOnInitBetweenFloors() {
	outputDevice.motorDirection(D_Down)
	elevator.dirn = D_Down
	elevator.behaviour = EB_Moving
}

func fsmOnRequestButtonPress(btnFloor int, btnType Button) {
	fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %s)\n", btnFloor, elevioButtonToString(btnType))
	elevatorPrint(elevator)

	switch elevator.behaviour {
	case EB_DoorOpen:
		if requestsShouldClearImmediately(elevator, btnFloor, btnType) {
			timerStart(elevator.config.doorOpenDurationS)
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
			outputDevice.doorLight(1)
			timerStart(elevator.config.doorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)
		case EB_Moving:
			outputDevice.motorDirection(elevator.dirn)
		}
	}

	setAllLights(elevator)
	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func fsmOnFloorArrival(newFloor int) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)
	elevatorPrint(elevator)
	elevator.floor = newFloor
	outputDevice.floorIndicator(elevator.floor)

	if elevator.behaviour == EB_Moving && requestsShouldStop(elevator) {
		outputDevice.motorDirection(D_Stop)
		outputDevice.doorLight(1)
		elevator = requestsClearAtCurrentFloor(elevator)
		timerStart(elevator.config.doorOpenDurationS)
		setAllLights(elevator)
		elevator.behaviour = EB_DoorOpen
	}

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func fsmOnDoorTimeout() {
	fmt.Println("\n\nfsmOnDoorTimeout()")
	elevatorPrint(elevator)

	if elevator.behaviour == EB_DoorOpen {
		pair := requestsChooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.behaviour = pair.behaviour
		switch elevator.behaviour {
		case EB_DoorOpen:
			timerStart(elevator.config.doorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case EB_Moving, EB_Idle:
			outputDevice.doorLight(0)
			outputDevice.motorDirection(elevator.dirn)
		}
	}

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func fsmOnObstruction() {
	fmt.Printf("\n\nfsmOnObstruction()\n")
	elevatorPrint(elevator)

	outputDevice.doorLight(1)
	timerStop()

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func fsmOnObstructionCleared() {
	fmt.Printf("\n\nfsmOnObstructionCleared()\n")
	elevatorPrint(elevator)

	outputDevice.doorLight(0)
	timerStart(elevator.config.doorOpenDurationS)

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
