package main

import (
	// "fmt"
	"net"
	"sync"
)

var (
	sock       net.Conn
	sockMutex  sync.Mutex
)

// const (
// 	N_FLOORS  = 4
// 	N_BUTTONS = 3
// )

type MotorDirection int
type ButtonType int

const (
	DIRN_DOWN MotorDirection = -1
	DIRN_STOP MotorDirection = 0
	DIRN_UP   MotorDirection = 1
)

const (
	BUTTON_CALL_UP   ButtonType = 0
	BUTTON_CALL_DOWN ButtonType = 1
	BUTTON_COMMAND   ButtonType = 2
)

func elevatorHardwareInit() {
	var err error
	sock, err = net.Dial("tcp", "localhost:15657")
	if err != nil {
		panic("Unable to connect to simulator server")
	}

	sock.Write([]byte{0, 0, 0, 0})
}

func elevatorHardwareSetMotorDirection(dirn Dirn) {
	sockMutex.Lock()
	sock.Write([]byte{1, byte(dirn), 0, 0})
	sockMutex.Unlock()
}

func elevatorHardwareSetButtonLamp(button Button, floor int, value int) {
	sockMutex.Lock()
	sock.Write([]byte{2, byte(button), byte(floor), byte(value)})
	sockMutex.Unlock()
}

func elevatorHardwareSetFloorIndicator(floor int) {
	sockMutex.Lock()
	sock.Write([]byte{3, byte(floor), 0, 0})
	sockMutex.Unlock()
}

func elevatorHardwareSetDoorOpenLamp(value int) {
	sockMutex.Lock()
	sock.Write([]byte{4, byte(value), 0, 0})
	sockMutex.Unlock()
}

func elevatorHardwareSetStopLamp(value int) {
	sockMutex.Lock()
	sock.Write([]byte{5, byte(value), 0, 0})
	sockMutex.Unlock()
}

func elevatorHardwareGetButtonSignal(button Button, floor int) int {
	sockMutex.Lock()
	sock.Write([]byte{6, byte(button), byte(floor), 0})
	buf := make([]byte, 4)
	sock.Read(buf)
	sockMutex.Unlock()
	return int(buf[1])
}

func elevatorHardwareGetFloorSensorSignal() int {
	sockMutex.Lock()
	sock.Write([]byte{7, 0, 0, 0})
	buf := make([]byte, 4)
	sock.Read(buf)
	sockMutex.Unlock()
	if buf[1] != 0 {
		return int(buf[2])
	}
	return -1
}

func elevatorHardwareGetStopSignal() int {
	sockMutex.Lock()
	sock.Write([]byte{8, 0, 0, 0})
	buf := make([]byte, 4)
	sock.Read(buf)
	sockMutex.Unlock()
	return int(buf[1])
}

func elevatorHardwareGetObstructionSignal() int {
	sockMutex.Lock()
	sock.Write([]byte{9, 0, 0, 0})
	buf := make([]byte, 4)
	sock.Read(buf)
	sockMutex.Unlock()
	return int(buf[1])
}