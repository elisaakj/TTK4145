package elevatorStateMachine

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
	"fmt"
	"time"
)

type ElevatorState int

const (
	IDLE ElevatorState = iota
	DOOR_OPEN
	MOVING
	UNAVAILABLE
)

type ClearRequestMode int

const (
	CLEAR_ALL ClearRequestMode = iota
	CLEAR_DIRECTION
)

type Elevator struct {
	ID                int
	Floor             int
	Dirn              elevio.MotorDirection
	Requests          [][]bool
	State             ElevatorState
	ClearRequestMode  ClearRequestMode
	DoorOpenDurationS float64
	Obstructed        bool
	Active            bool
	LastSeen          time.Time
	OrderID           int
}

// this struct with channels is not implemented properly yet
type FsmChannels struct {
	//OrderComplete  chan int
	Elevator       chan Elevator
	NewOrder       chan elevio.ButtonEvent
	ArrivedAtFloor chan int
	Obstruction    chan bool
}

// REMOVE??
//
// func elevatorPrint(es Elevator) {
// 	fmt.Println("  +--------------------+")
// 	fmt.Printf("  |floor = %-2d          |\n", es.Floor)
// 	fmt.Printf("  |dirn  = %-12.12s|\n", stateToString(es.State))
// 	fmt.Println("  +--------------------+")
// 	fmt.Println("  |  | up  | dn  | cab |")
// 	for f := NUM_FLOORS - 1; f >= 0; f-- {
// 		fmt.Printf("  | %d", f)
// 		for btn := 0; btn < NUM_BUTTONS; btn++ {
// 			if (f == NUM_FLOORS-1 && btn == int(elevio.BT_HallUp)) || (f == 0 && btn == int(elevio.BT_HallDown)) {
// 				fmt.Print("|     ")
// 			} else {
// 				if es.Requests[f][btn] {
// 					fmt.Print("|  #  ")
// 				} else {
// 					fmt.Print("|  -  ")
// 				}
// 			}
// 		}
// 		fmt.Println("|")
// 	}
// 	fmt.Println("  +--------------------+")
// }

func createUninitializedElevator(id int, numFloors int, numButtons int) Elevator {
	return Elevator{
		ID:                id,
		Floor:             elevio.GetFloor(),
		Dirn:              elevio.DIRN_STOP,
		State:             MOVING,
		Requests:          make([][]bool, numFloors),
		DoorOpenDurationS: config.DOOR_OPEN_DURATION,
	}
}

// should init and then run stateMachine
func RunElevator(ch FsmChannels, id int, numFloors int, numButtons int) {

	// Initialize elevator
	//elevator := createUninitializedElevator(id, numFloors, numButtons)

	//if elevio.GetFloor() == -1 {
	//	fsmOnInitBetweenFloors()
	//}

	elevator := Elevator{
		ID:                id,
		State:             IDLE,
		Dirn:              elevio.DIRN_STOP,
		Floor:             elevio.GetFloor(),
		Requests:          make([][]bool, numFloors),
		DoorOpenDurationS: config.DOOR_OPEN_DURATION,
	}

	for i := range elevator.Requests {
		elevator.Requests[i] = make([]bool, numButtons)
	}

	if elevator.Floor == config.INVALID_FLOOR {
		elevator.Floor = 0
		elevator.Dirn = elevio.DIRN_DOWN
		elevator.State = MOVING
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
			handleRequestButtonPress(&elevator, NewOrder, ch)

		case elevator.Floor = <-ch.ArrivedAtFloor:
			fmt.Printf("Floor sensor triggered: %d\n", elevator.Floor) // Debugging
			onFloorArrival(elevator.Floor, &elevator, ch, numButtons)

		case obstruction := <-ch.Obstruction:
			fmt.Printf("Obstruction event: %t\n", obstruction) // Debugging
			handleObstruction(&elevator, obstruction, ch)

		case <-time.After(time.Duration(elevator.DoorOpenDurationS) * time.Second):
			if elevator.State == DOOR_OPEN {
				fmt.Println("Door timeout, closing doors")
				handleDoorTimeout(&elevator, ch)
			}
		}
	}
}
