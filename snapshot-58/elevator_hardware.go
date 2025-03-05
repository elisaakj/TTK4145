package main

import (
	"log"
	"net"
	"sync"
)

var (
	sock      net.Conn
	sockMutex sync.Mutex
)

type MotorDirection int
type ButtonType int
type LampType int

const (
	DIRN_DOWN MotorDirection = -1
	DIRN_STOP MotorDirection = 0
	DIRN_UP   MotorDirection = 1
)

const (
	BUTTON_HALL_UP ButtonType = iota
	BUTTON_HALL_DOWN
	BUTTON_CAB
)

const (
	LAMP_BUTTON LampType = iota
	LAMP_FLOOR_INDICATOR
	LAMP_DOOR
	LAMP_STOP
)

func hardwareInit() {
	var err error
	sock, err = net.Dial("tcp", "localhost:15657")
	if err != nil {
		// panic("Unable to connect to simulator server")
		log.Fatalf("Unable to connect to simulator server: %v", err)
	}
	sock.Write([]byte{0, 0, 0, 0})
}

func hardwareExecuteCommand(command byte, arg1 byte, arg2 byte, arg3 byte) byte {
	sockMutex.Lock()
	defer sockMutex.Unlock()

	sock.Write([]byte{command, arg1, arg2, arg3})
	buf := make([]byte, 4)
	sock.Read(buf)

	return buf[1]
}

func hardwareSetMotorDirection(dirn MotorDirection) {
	hardwareExecuteCommand(1, byte(dirn), 0, 0)
}

func hardwareGetButtonSignal(floor int, button ButtonType) int {
	return int(hardwareExecuteCommand(6, byte(button), byte(floor), 0))
}

func hardwareGetFloorSensorSignal() int {
	result := hardwareExecuteCommand(7, 0, 0, 0)
	if result != 0 {
		return int(hardwareExecuteCommand(7, 0, 0, 0))
	}
	return -1
}

func hardwareGetStopSignal() int {
	return int(hardwareExecuteCommand(8, 0, 0, 0))
}

func hardwareGetObstructionSignal() int {
	return int(hardwareExecuteCommand(9, 0, 0, 0))
}

func hardwareLampToggle(lampType LampType, button ButtonType, floor int, value int) { // bedre navn for value?
	sockMutex.Lock()
	defer sockMutex.Unlock()

	var command []byte

	switch lampType {
	case LAMP_BUTTON:
		command = []byte{2, byte(button), byte(floor), byte(value)}
	case LAMP_FLOOR_INDICATOR:
		command = []byte{3, byte(floor), 0, 0}
	case LAMP_DOOR:
		command = []byte{4, byte(value), 0, 0}
	case LAMP_STOP:
		command = []byte{5, byte(value), 0, 0}
	default:
		return
	}

	sock.Write(command)
}

/*package main

import (
	// "fmt"
	"log"
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

func hardwareInit() {
	var err error
	sock, err = net.Dial("tcp", "localhost:15657")
	if err != nil {
		panic("Unable to connect to simulator server")
	}

	sock.Write([]byte{0, 0, 0, 0})
}

func hardwareSetMotorDirection(dirn Dirn) {
	sockMutex.Lock()
	sock.Write([]byte{1, byte(dirn), 0, 0})
	sockMutex.Unlock()
}

func hardwareSetButtonLamp(button Button, floor int, value int) {
	sockMutex.Lock()
	sock.Write([]byte{2, byte(button), byte(floor), byte(value)})
	sockMutex.Unlock()
}

func hardwareSetFloorIndicator(floor int) {
	sockMutex.Lock()
	sock.Write([]byte{3, byte(floor), 0, 0})
	sockMutex.Unlock()
}

func hardwareSetDoorOpenLamp(value int) {
	sockMutex.Lock()
	sock.Write([]byte{4, byte(value), 0, 0})
	sockMutex.Unlock()
}

func hardwareSetStopLamp(value int) {
	sockMutex.Lock()
	sock.Write([]byte{5, byte(value), 0, 0})
	sockMutex.Unlock()
}

func hardwareGetButtonSignal(button Button, floor int) int {
	sockMutex.Lock()
	sock.Write([]byte{6, byte(button), byte(floor), 0})
	buf := make([]byte, 4)
	sock.Read(buf)
	sockMutex.Unlock()
	return int(buf[1])
}

func hardwareGetFloorSensorSignal() int {

	if sock == nil  {
		log.Fatal("Socket is nil. Ensure it is initialized before usage")
	}

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

func hardwareGetStopSignal() int {
	sockMutex.Lock()
	sock.Write([]byte{8, 0, 0, 0})
	buf := make([]byte, 4)
	sock.Read(buf)
	sockMutex.Unlock()
	return int(buf[1])
}

func hardwareGetObstructionSignal() int {
	sockMutex.Lock()
	sock.Write([]byte{9, 0, 0, 0})
	buf := make([]byte, 4)
	sock.Read(buf)
	sockMutex.Unlock()
	return int(buf[1])
}
*/
