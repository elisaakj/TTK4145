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

	if elevator.Floor == config.INVALID_FLOOR {
		elevator.Floor = 0
		elevator.Dirn = elevio.DIRN_DOWN
		elevator.State = config.MOVING
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- elevator

	stuckTimer := time.NewTimer(time.Duration(config.STUCK_TIMER) * time.Second)
	stuckTimer.Stop()
	stuckTimerRunning := false

	obstructionTimer := time.NewTimer(time.Duration(config.OBSTRUCTION_TIMER))
	obstructionTimer.Stop()

	for {
		select {
		case NewOrder := <-ch.NewOrder:
			fmt.Printf("RunElevator received order: %+v\n", NewOrder)
			handleRequestButtonPress(&elevator, NewOrder, ch)

			if elevator.State == config.MOVING && !stuckTimerRunning {
				stuckTimer.Reset(time.Duration(config.STUCK_TIMER) * time.Second)
				stuckTimerRunning = true
			}

		case elevator.Floor = <-ch.ArrivedAtFloor:
			//fmt.Printf("Floor sensor triggered: %d\n", elevator.Floor)
			onFloorArrival(elevator.Floor, &elevator, ch)

			if stuckTimerRunning {
				stuckTimer.Stop()
				stuckTimerRunning = false
			}

		case obstruction := <-ch.Obstruction:
			fmt.Printf("Obstruction event: %t\n", obstruction)

			obstructionTimer.Reset(time.Duration(config.OBSTRUCTION_TIMER) * time.Second)
			handleObstruction(&elevator, obstruction, ch)

		case <-time.After(time.Duration(config.DOOR_OPEN_DURATION) * time.Second):
			if elevator.State == config.DOOR_OPEN {
				fmt.Println("Door timeout, closing doors")
				handleDoorTimeout(&elevator, ch)
			}

		case <-stuckTimer.C:

			if elevator.State == config.MOVING {
				fmt.Printf("Elevator %d is stuck!\n", elevator.ID)
				elevio.SetMotorDirection(elevator.Dirn)
				elevator.State = config.UNAVAILABLE

				clearHallOrder(elevator)

				ch.Elevator <- elevator
				stuckTimerRunning = false
			}

		case <-obstructionTimer.C:
			elevio.SetMotorDirection(elevio.DIRN_STOP)
			elevator.State = config.UNAVAILABLE

			clearHallOrder(elevator)

			handleDoorTimeout(&elevator, ch)
			ch.Elevator <- elevator
		}
	}
}

func clearHallOrder(elevator Elevator) {
	for f := 0; f < config.NUM_FLOORS; f++ {
		for b := 0; b < config.NUM_BUTTONS-1; b++ {
			elevator.Requests[f][b] = false
		}
	}
}
