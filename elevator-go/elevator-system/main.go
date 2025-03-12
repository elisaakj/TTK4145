package main

import (
	"Driver-go/elevator-system/communication"
	"Driver-go/elevator-system/elevatorStateMachine"
	"Network-go/network/localip"
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	fmt.Println("Started!")

	// Init network and elevator

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// Ensure ID is provided, convert to int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error: Invalid ID format, using default ID 1")
		idInt = 1 // Default ID if conversion fails
	}

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	updateChannel := make(chan communication.ElevatorState)
	communication.InitNetwork(idInt, updateChannel)

	elevatorStateMachine.RunElevator()
}

/*
	go func() {
		for {
			select {
			case newState := <-updateChannel:
				// Handle received state updates
				fmt.Printf("Elevator %d received update: %+v\n", elevatorID, newState)
				if newState.ID != elevatorID {
					fmt.Printf("Updating peer elevator %d state...\n", newState.ID)
				}
			default:
				// Request button handling
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
					fmt.Println("Obstruction detected! Keeping doors open.")
					fsmOnObstruction()
				} else if obstr == 0 && prevObstr != 0 {
					fmt.Println("Obstruction cleared. Resuming operation.")
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

				// Send periodic state updates
				state := ElevatorState{
					ID:        elevatorID,
					floor:     elevator.floor,
					dirn:      elevator.dirn,
					behaviour: elevator.behaviour,
					active:    true,
				}
				sendStateUpdate(state, nil)

				time.Sleep(inputPollRate)
			}
		}
	}()

	// Keep the program running
	select {}
}
*/
