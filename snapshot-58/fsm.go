package main

var (
	elevator     Elevator
	outputDevice ElevOutputDevice
)

func init() {
	elevator = elevatorUninitialized()
	// Simulating config loading
	elevator.config.doorOpenDurationS = 3.0 // Example default value
	elevator.config.clearRequests = CLEAR_ALL
	outputDevice = elevioGetOutputDevice()
}

func setAllLights(e Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			outputDevice.requestButtonLight(floor, ButtonType(btn), boolToInt(e.requests[floor][btn]))
		}
	}
}

func fsmOnInitBetweenFloors() {
	outputDevice.motorDirection(DIRN_DOWN)
	elevator.dirn = DIRN_DOWN
	elevator.state = ELEVSTATE_MOVING
}

func fsmOnRequestButtonPress(btnFloor int, btnType ButtonType) {
	// fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %s)\n", btnFloor, elevioButtonToString(btnType))

	switch elevator.state {
	case ELEVSTATE_DOOR_OPEN:
		if requestsShouldClearImmediately(elevator, btnFloor, btnType) {
			timerStart(elevator.config.doorOpenDurationS)
		} else {
			elevator.requests[btnFloor][btnType] = true
		}
	case ELEVSTATE_MOVING:
		elevator.requests[btnFloor][btnType] = true
	case ELEVSTATE_IDLE:
		elevator.requests[btnFloor][btnType] = true
		pair := requestsChooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.state = pair.state
		switch pair.state {
		case ELEVSTATE_DOOR_OPEN:
			outputDevice.doorLight(1)
			timerStart(elevator.config.doorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)
		case ELEVSTATE_MOVING:
			outputDevice.motorDirection(elevator.dirn)
		}
	}

	setAllLights(elevator)
}

func fsmOnFloorArrival(newFloor int) {
	// fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)
	elevator.floor = newFloor
	outputDevice.floorIndicator(elevator.floor)

	if elevator.state == ELEVSTATE_MOVING && requestsShouldStop(elevator) {
		outputDevice.motorDirection(DIRN_STOP)
		outputDevice.doorLight(1)
		elevator = requestsClearAtCurrentFloor(elevator)
		timerStart(elevator.config.doorOpenDurationS)
		setAllLights(elevator)
		elevator.state = ELEVSTATE_DOOR_OPEN
	}
}

func fsmOnDoorTimeout() {
	// fmt.Println("\n\nfsmOnDoorTimeout()")

	if hardwareGetObstructionSignal() != 0 {
		timerStart(elevator.config.doorOpenDurationS)
		return
	}

	if elevator.state == ELEVSTATE_DOOR_OPEN {
		pair := requestsChooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.state = pair.state
		switch elevator.state {
		case ELEVSTATE_DOOR_OPEN:
			timerStart(elevator.config.doorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case ELEVSTATE_MOVING, ELEVSTATE_IDLE:
			outputDevice.doorLight(0)
			outputDevice.motorDirection(elevator.dirn)
		}
	}
}

func fsmOnObstruction() {
	if elevator.state == ELEVSTATE_DOOR_OPEN {
		outputDevice.doorLight(1) // Keep door open
		timerStop()               // Stop the timer while obstructed
	}
}

func fsmOnObstructionCleared() {
	if elevator.state == ELEVSTATE_DOOR_OPEN {
		outputDevice.doorLight(1)                     // Keep door open
		timerStart(elevator.config.doorOpenDurationS) // Restart door timer
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

/*package main

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

	if hardwareGetObstructionSignal() != 0 {
		timerStart(elevator.config.doorOpenDurationS)
		return
	}

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
	fmt.Println("\n\nfsmOnObstruction()")
	elevatorPrint(elevator)

	if elevator.behaviour == EB_DoorOpen {
		outputDevice.doorLight(1) // Keep door open
		timerStop()               // Stop the timer while obstructed
	}

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

func fsmOnObstructionCleared() {
	fmt.Println("\n\nfsmOnObstructionCleared()")
	elevatorPrint(elevator)

	if elevator.behaviour == EB_DoorOpen {
		outputDevice.doorLight(1)                     // Keep door open
		timerStart(elevator.config.doorOpenDurationS) // Restart door timer
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
*/
