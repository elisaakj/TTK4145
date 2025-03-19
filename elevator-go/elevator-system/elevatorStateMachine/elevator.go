package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"fmt"
	"time"
)

type ElevatorBehaviour int
type ClearRequestVariant int

// this struct with channels is not implemented properly yet
type FsmChannels struct {
	OrderComplete chan int
	Elevator      chan Elevator
	// state error here?
	NewOrder       chan elevio.ButtonEvent
	ArrivedAtFloor chan int
	Obstruction    chan bool
}

type Elevator struct {
	ID                  int
	Floor               int
	Dirn                elevio.MotorDirection
	Requests            [][]bool
	Behaviour           ElevatorBehaviour
	ClearRequestVariant ClearRequestVariant
	DoorOpenDurationS   float64
	Obstructed          bool
	Active              bool
	LastSeen            time.Time
	OrderID             int 
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
	fmt.Printf("  |floor = %-2d          |\n", es.Floor)
	fmt.Printf("  |dirn  = %-12.12s|\n", ebToString(es.Behaviour))
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")
	for f := N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (f == N_FLOORS-1 && btn == int(elevio.BT_HallUp)) || (f == 0 && btn == int(elevio.BT_HallDown)) {
				fmt.Print("|     ")
			} else {
				if es.Requests[f][btn] {
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
		Floor:             elevio.GetFloor(),
		Dirn:              elevio.MD_Stop,
		Behaviour:         EB_Idle,
		Requests:          make([][]bool, numFloors),
		DoorOpenDurationS: 3.0,
	}
}

// should init and then run stateMachine
func RunElevator(ch FsmChannels, id int, numFloors int, numButtons int) {

	// Initialize elevator
	//elevator := elevatorUninitialized(id, numFloors, numButtons)

	//if elevio.GetFloor() == -1 {
	//	fsmOnInitBetweenFloors()
	//}

	elevator := Elevator{
		ID:                id,
		Behaviour:         EB_Idle,
		Dirn:              elevio.MD_Stop,
		Floor:             elevio.GetFloor(),
		Requests:          make([][]bool, numFloors),
		DoorOpenDurationS: 3.0,
	}

	for i := range elevator.Requests {
		elevator.Requests[i] = make([]bool, numButtons)
	}

	if elevator.Floor == -1 {
		elevator.Floor = 0
		elevator.Dirn = elevio.MD_Down
		elevator.Behaviour = EB_Moving
		elevio.SetMotorDirection(elevator.Dirn)
	}

	// Send initialized elevator to channels
	//ch.Elevator <- elevator
	select {
	case ch.Elevator <- elevator:
		fmt.Println("Elevator state sent to channel")
	default:
		fmt.Println("Warning: No receiver for ch.Elevator!")
	}

	// var lastButtonEvent elevio.ButtonEvent
	fmt.Println("RunElevator started!") // Add this to confirm the function is running

	for {
		select {
		case NewOrder := <-ch.NewOrder:
			fmt.Printf("RunElevator received order: %+v\n", NewOrder) // Debugging
			fsmOnRequestButtonPress(&elevator, NewOrder, ch)

		case elevator.Floor = <-ch.ArrivedAtFloor:
			fmt.Printf("Floor sensor triggered: %d\n", elevator.Floor) // Debugging
			fsmOnFloorArrival(elevator.Floor, &elevator, ch, numButtons)

		case obstruction := <-ch.Obstruction:
			fmt.Printf("Obstruction event: %t\n", obstruction) // Debugging
			fsmOnObstruction(&elevator, obstruction, ch)

		case <-time.After(time.Duration(elevator.DoorOpenDurationS) * time.Second):
			if elevator.Behaviour == EB_DoorOpen {
				fmt.Println("Door timeout, closing doors")
				fsmOnDoorTimeout(&elevator, ch)
			}
		}
	}
	/*
		if input.floorSensor == -1 {
			fsmOnInitBetweenFloors()
		}

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
