package elevatorStateMachine

import (
	"fmt"
	"time"
)

type ElevatorBehaviour int

type ClearRequestVariant int

type Elevator struct {
	ID        int
	floor     int
	dirn      Dirn
	requests  [N_FLOORS][N_BUTTONS]bool
	behaviour ElevatorBehaviour
	config    struct {
		clearRequestVariant ClearRequestVariant
		doorOpenDurationS   float64
	}
	active   bool
	lastSeen time.Time
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

//

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
			if (f == N_FLOORS-1 && btn == int(B_HallUp)) || (f == 0 && btn == int(B_HallDown)) {
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

func elevatorUninitialized() Elevator {
	return Elevator{
		floor:     -1,
		dirn:      D_Stop,
		behaviour: EB_Idle,
		config: struct {
			clearRequestVariant ClearRequestVariant
			doorOpenDurationS   float64
		}{
			clearRequestVariant: CV_All,
			doorOpenDurationS:   3.0,
		},
	}
}

func RunElevator() {
	inputPollRate := 25 * time.Millisecond
	input := elevioGetInputDevice()

	hardwareInit(elevator.ID)

	if input.floorSensor() == -1 {
		fsmOnInitBetweenFloors()
	}

	prevRequests := make([][]int, N_FLOORS)
	for i := range prevRequests {
		prevRequests[i] = make([]int, N_BUTTONS)
	}

	prevFloor := -1
	prevObstr := 0

	for {
		for f := 0; f < N_FLOORS; f++ {
			for b := 0; b < N_BUTTONS; b++ {
				v := input.requestButton(f, Button(b))
				if v != 0 && v != prevRequests[f][b] {
					fsmOnRequestButtonPress(f, Button(b))
				}
				prevRequests[f][b] = v
			}
		}

		// Obstruction handling
		obstr := input.obstruction()
		if obstr != 0 && prevObstr == 0 {
			fsmOnObstruction()
		} else if obstr == 0 && prevObstr != 0 {
			fsmOnObstructionCleared()
		}
		prevObstr = obstr

		// Floor sensor handling
		floor := input.floorSensor()
		if floor != -1 && floor != prevFloor {
			fsmOnFloorArrival(floor)
		}
		prevFloor = floor

		// Timer handling
		if timerTimedOut() {
			timerStop()
			fsmOnDoorTimeout()
		}

		time.Sleep(inputPollRate)
	}
}

///////////////// REDCLARED //////////////////////
// const (
// 	D_Down Dirn = -1
// 	D_Stop Dirn = 0
// 	D_Up   Dirn = 1
// )

// const (
// 	B_HallUp Button = iota
// 	B_HallDown
// 	B_Cab
// )

// type Dirn int

// type Button int
