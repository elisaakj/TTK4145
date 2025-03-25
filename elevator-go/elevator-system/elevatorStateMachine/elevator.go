package elevatorStateMachine

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevio"
	"fmt"
	"time"
)

type Elevator struct {
	ID         int
	Floor      int
	Dirn       elevio.MotorDirection
	Requests   [][]bool
	State      config.ElevatorState
	Obstructed bool
	OrderID    int
}

func initElevator(id int) Elevator {
	return Elevator{
		ID:       id,
		Floor:    elevio.GetFloor(),
		Dirn:     elevio.DIRN_STOP,
		State:    config.IDLE,
		Requests: make([][]bool, config.NUM_FLOORS),
	}
}

func RunElevator(ch FsmChannels, id int) {

	elevator := initElevator(id)

	for i := range elevator.Requests {
		elevator.Requests[i] = make([]bool, config.NUM_BUTTONS)
	}

	// Re-emit old cab calls to FSM
	for f := 0; f < config.NUM_FLOORS; f++ {
		if elevator.Requests[f][elevio.BUTTON_CAB] {
			ch.NewOrder <- elevio.ButtonEvent{
				Floor:  f,
				Button: elevio.BUTTON_CAB,
		}
	}
}

	if elevator.Floor == config.INVALID_FLOOR {
		elevator.Floor = 0
		elevator.Dirn = elevio.DIRN_DOWN
		elevator.State = config.MOVING
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- elevator

	for {
		select {
		case NewOrder := <-ch.NewOrder:
			fmt.Printf("RunElevator received order: %+v\n", NewOrder)
			handleRequestButtonPress(&elevator, NewOrder, ch)

		case elevator.Floor = <-ch.ArrivedAtFloor:
			fmt.Printf("Floor sensor triggered: %d\n", elevator.Floor)
			onFloorArrival(elevator.Floor, &elevator, ch)

		case obstruction := <-ch.Obstruction:
			fmt.Printf("Obstruction event: %t\n", obstruction)
			handleObstruction(&elevator, obstruction, ch)

		case <-time.After(time.Duration(config.DOOR_OPEN_DURATION) * time.Second):
			if elevator.State == config.DOOR_OPEN {
				fmt.Println("Door timeout, closing doors")
				handleDoorTimeout(&elevator, ch)
			}
		}
	}
}
