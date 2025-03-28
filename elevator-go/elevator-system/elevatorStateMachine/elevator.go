package elevatorStateMachine

import (
	"Driver-go/elevator-system/common"
	"Driver-go/elevator-system/elevio"
	"fmt"
	"time"
)

func initElevator(id int) common.Elevator {
	return  common.Elevator{
		ID:       id,
		Floor:    elevio.GetFloor(),
		Dirn:     common.DIRN_STOP,
		State:    common.IDLE,
		Requests: make([][]bool, common.NUM_FLOORS),
	}
}

func RunElevator(ch common.FsmChannels, id int) {

	elevator := initElevator(id)

	for i := range elevator.Requests {
		elevator.Requests[i] = make([]bool, common.NUM_BUTTONS)
	}

	if elevator.Floor == common.INVALID_FLOOR {
		elevator.Floor = 0
		elevator.Dirn = common.DIRN_DOWN
		elevator.State = common.MOVING
		elevio.SetMotorDirection(elevator.Dirn)
	}

	ch.Elevator <- elevator

	stuckTimer := time.NewTimer(time.Duration(common.STUCK_TIMER) * time.Second)
	stuckTimer.Stop()
	stuckTimerRunning := false

	obstructionTimer := time.NewTimer(time.Duration(common.OBSTRUCTION_TIMER))
	obstructionTimer.Stop()

	for {
		select {
		case NewOrder := <-ch.NewOrder:
			fmt.Printf("RunElevator received order: %+v\n", NewOrder)
			handleRequestButtonPress(&elevator, NewOrder, ch)

			if elevator.State == common.MOVING && !stuckTimerRunning {
				stuckTimer.Reset(time.Duration(common.STUCK_TIMER) * time.Second)
				stuckTimerRunning = true
			}

		case elevator.Floor = <-ch.ArrivedAtFloor:
			//fmt.Printf("Floor sensor triggered: %d\n", elevator.Floor)
			onFloorArrival(elevator.Floor, &elevator, ch)

			if stuckTimerRunning {
				stuckTimer.Stop()
				stuckTimer.Reset(time.Duration(common.STUCK_TIMER)*time.Second)
				stuckTimerRunning = false
			}

		case obstruction := <-ch.Obstruction:
			fmt.Printf("Obstruction event: %t\n", obstruction)

			obstructionTimer.Reset(time.Duration(common.OBSTRUCTION_TIMER) * time.Second)
			handleObstruction(&elevator, obstruction, ch)

		case <-time.After(time.Duration(common.DOOR_OPEN_TIMER) * time.Second):
			if elevator.State == common.DOOR_OPEN {
				fmt.Println("Door timeout, closing doors")
				handleDoorTimeout(&elevator, ch)
			}

		case <-stuckTimer.C:
			if elevator.State == common.MOVING {
				fmt.Printf("Elevator %d is stuck!\n", elevator.ID)
				elevio.SetMotorDirection(elevator.Dirn)
				elevator.State = common.UNAVAILABLE

				clearHallOrder(elevator)

				ch.Elevator <- elevator
				stuckTimerRunning = false
			}

		case <-obstructionTimer.C:
			elevio.SetMotorDirection(common.DIRN_STOP)
			elevator.State = common.UNAVAILABLE

			clearHallOrder(elevator)

			handleDoorTimeout(&elevator, ch)
			ch.Elevator <- elevator
		}
	}
}

func clearHallOrder(elevator common.Elevator) {
	for f := 0; f < common.NUM_FLOORS; f++ {
		for b := 0; b < common.NUM_BUTTONS-1; b++ {
			elevator.Requests[f][b] = false
		}
	}
}
