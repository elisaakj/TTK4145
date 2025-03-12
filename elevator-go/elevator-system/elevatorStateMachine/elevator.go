package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"fmt"
	"time"
)

type ElevatorBehaviour int
type ClearRequestVariant int

// this struct with channels is not implemented properly yet
type stateMachineChannels struct {
	orderComplete chan int
	elevator      chan Elevator
	// state error here?
	newOrder       chan elevio.ButtonType
	arrivedAtFloor chan int
	obstruction    chan bool
}

type Elevator struct {
	ID                  int
	floor               int
	dirn                elevio.MotorDirection
	requests            [][]bool
	behaviour           ElevatorBehaviour
	clearRequestVariant ClearRequestVariant
	doorOpenDurationS   float64
	active              bool
	lastSeen            time.Time
}

type ElevInputDevice struct {
	floorSensor   int
	requestButton bool
	stopButton    bool
	obstruction   bool
}

type ElevOutputDevice struct {
	floorIndicator     func(int)
	requestButtonLight func(elevio.ButtonType, int, bool)
	doorLight          func(bool)
	stopButtonLight    func(bool)
	motorDirection     func(elevio.MotorDirection)
}

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

func ebToString(eb ElevatorBehaviour) string {
	switch eb {
	case EB_Idle:
		return "EB_Idle"
	case EB_DoorOpen:
		return "EB_DoorOpen"
	case EB_Moving:
		return "EB_Moving"
	default:
		return "EB_UNDEFINED"
	}
}

func elevatorPrint(es Elevator) {
	fmt.Println("  +--------------------+")
	fmt.Printf("  |floor = %-2d          |\n", es.floor)
	fmt.Printf("  |dirn  = %-12.12s|\n", ebToString(es.behaviour))
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")
	for f := N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (f == N_FLOORS-1 && btn == int(elevio.BT_HallUp)) || (f == 0 && btn == int(elevio.BT_HallDown)) {
				fmt.Print("|     ")
			} else {
				if es.requests[f][btn] {
					fmt.Print("|  #  ")
				} else {
					fmt.Print("|  -  ")
				}
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}

func elevioGetInputDevice(button elevio.ButtonType, floor int) ElevInputDevice {
	return ElevInputDevice{
		floorSensor:   elevio.GetFloor(),
		requestButton: elevio.GetButton(button, floor),
		stopButton:    elevio.GetStop(),
		obstruction:   elevio.GetObstruction(),
	}
}

func elevioGetOutputDevice() ElevOutputDevice {
	return ElevOutputDevice{
		floorIndicator:     elevio.SetFloorIndicator,
		requestButtonLight: elevio.SetButtonLamp,
		doorLight:          elevio.SetDoorOpenLamp,
		stopButtonLight:    elevio.SetStopLamp,
		motorDirection:     elevio.SetMotorDirection,
	}
}

// change name probably
func elevatorUninitialized(id int, numFloors int, numButtons int) Elevator {
	return Elevator{
		ID:                id,
		floor:             elevio.GetFloor(),
		dirn:              elevio.MD_Stop,
		behaviour:         EB_Idle,
		requests:          make([][]bool, numFloors),
		doorOpenDurationS: 3.0,
	}
}

// should init and then run stateMachine
func RunElevator(ch stateMachineChannels, id int, numFloors int, numButtons int) {

	// Init elevator
	elevatorUninitialized(id, numFloors, numButtons)

	for i := range elevator.requests {
		elevator.requests[i] = make([]bool, numButtons)
	}

	ch.elevator <- elevator

	for {
		select {
		// run elevator in different cases
		case newOrder := <-ch.newOrder:
			fsmOnRequestButtonPress(elevator.floor, newOrder)
		case elevator.floor = <-ch.arrivedAtFloor:
			fsmOnFloorArrival(elevator.floor, <-ch.newOrder)
		case obstruction := <-ch.obstruction:
			//rewrite obstruction function in fsm
		case 
		}

		// needed? not used now as far as I can see
		if input.floorSensor == -1 {
			fsmOnInitBetweenFloors()
		}

		/*
		inputPollRate := 25 * time.Millisecond
		input := elevioGetInputDevice(button, floor)

		prevRequests := make([][]int, N_FLOORS)
		for i := range prevRequests {
			prevRequests[i] = make([]int, N_BUTTONS)
		}

		prevFloor := -1
		prevObstr := 0

		for {
			for f := 0; f < N_FLOORS; f++ {
				for b := 0; b < N_BUTTONS; b++ {
					v := input.requestButton(button, f)
					if v != 0 && v != prevRequests[f][b] {
						fsmOnRequestButtonPress(f, Button(b))
					}
					prevRequests[f][b] = v
				}
			}

			// Obstruction handling
			// Make as own function
			obstr := input.obstruction
			if obstr != 0 && prevObstr == 0 {
				fsmOnObstruction()
			} else if obstr == 0 && prevObstr != 0 {
				fsmOnObstructionCleared()
			}
			prevObstr = obstr

			// Floor sensor handling
			// Makes as own function
			floor := input.floorSensor
			if floor != -1 && floor != prevFloor {
				fsmOnFloorArrival(floor)
			}
			prevFloor = floor

			// Timer handling
			if timerTimedOut() {
				timerStop()
				fsmOnDoorTimeout(button)
			}

			time.Sleep(inputPollRate)
			*/
		}
	}
