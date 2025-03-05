package main

import (
	"time"
)

type ElevState int
type ClearRequests int
type Elevator struct {
	ID       int
	floor    int
	dirn     MotorDirection
	requests [N_FLOORS][N_BUTTONS]bool
	state    ElevState
	config   struct {
		clearRequests     ClearRequests
		doorOpenDurationS float64
	}
	active   bool
	lastSeen time.Time
}

const (
	ELEVSTATE_IDLE ElevState = iota
	ELEVSTATE_DOOR_OPEN
	ELEVSTATE_MOVING
)

const (
	CLEAR_ALL ClearRequests = iota
	CLEAR_DIRECTION
)

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

func ebToString(eb ElevState) string {
	switch eb {
	case ELEVSTATE_IDLE:
		return "ELEVSTATE_IDLE"
	case ELEVSTATE_DOOR_OPEN:
		return "ELEVSTATE_DOOR_OPEN"
	case ELEVSTATE_MOVING:
		return "ELEVSTATE_MOVING"
	default:
		return "ELEVSTATE_UNDEFINED"
	}
}

func elevatorUninitialized() Elevator {
	return Elevator{
		floor: -1,
		dirn:  DIRN_STOP,
		state: ELEVSTATE_IDLE,
		config: struct {
			clearRequests     ClearRequests
			doorOpenDurationS float64
		}{
			clearRequests:     CLEAR_ALL,
			doorOpenDurationS: 3.0,
		},
	}
}

/*package main

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
*/
