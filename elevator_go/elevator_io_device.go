package main

import (
	// "fmt"
)

type Dirn int
type Button int



type ElevInputDevice struct {
	floorSensor    func() int
	requestButton  func(int, Button) int
	stopButton     func() int
	obstruction    func() int
}

type ElevOutputDevice struct {
	floorIndicator     func(int)
	requestButtonLight func(int, Button, int)
	doorLight          func(int)
	stopButtonLight    func(int)
	motorDirection     func(Dirn)
}

const (
	D_Down Dirn = -1
	D_Stop Dirn = 0
	D_Up   Dirn = 1
)

const (
	B_HallUp Button = iota
	B_HallDown
	B_Cab
)

func elevioDirnToString(d Dirn) string {
	switch d {
	case D_Up:
		return "D_Up"
	case D_Down:
		return "D_Down"
	case D_Stop:
		return "D_Stop"
	default:
		return "D_UNDEFINED"
	}
}

func elevioButtonToString(b Button) string {
	switch b {
	case B_HallUp:
		return "B_HallUp"
	case B_HallDown:
		return "B_HallDown"
	case B_Cab:
		return "B_Cab"
	default:
		return "B_UNDEFINED"
	}
}

func elevioGetInputDevice() ElevInputDevice {
	return ElevInputDevice{
		floorSensor:    elevatorHardwareGetFloorSensorSignal,
		requestButton:  wrapRequestButton,
		stopButton:     elevatorHardwareGetStopSignal,
		obstruction:    elevatorHardwareGetObstructionSignal,
	}
}

func elevioGetOutputDevice() ElevOutputDevice {
	return ElevOutputDevice{
		floorIndicator:     elevatorHardwareSetFloorIndicator,
		requestButtonLight: wrapRequestButtonLight,
		doorLight:          elevatorHardwareSetDoorOpenLamp,
		stopButtonLight:    elevatorHardwareSetStopLamp,
		motorDirection:     wrapMotorDirection,
	}
}

func wrapRequestButton(floor int, btn Button) int {
	return elevatorHardwareGetButtonSignal(btn, floor)
}

func wrapRequestButtonLight(floor int, btn Button, value int) {
	elevatorHardwareSetButtonLamp(btn, floor, value)
}

func wrapMotorDirection(dirn Dirn) {
	elevatorHardwareSetMotorDirection(dirn)
}