package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("Started!")

	// Init network and elevator

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <elevator_id>")
		return
	}

	elevatorID := 0
	fmt.Sscanf(os.Args[1], "&d", elevatorID)

	updateChannel := make(chan ElevatorState)
	initNetwork(elevatorID, updateChannel)

	elevator := Elevator{
		ID:        elevatorID,
		floor:     0,
		dirn:      D_Stop,
		behaviour: EB_Idle,
		active:    true,
	}

	inputPollRate := 25 * time.Millisecond
	input := elevioGetInputDevice()

	hardwareInit()

	if input.floorSensor() == -1 {
		fsmOnInitBetweenFloors()
	}

	prevRequests := make([][]int, N_FLOORS)
	for i := range prevRequests {
		prevRequests[i] = make([]int, N_BUTTONS)
	}

	prevFloor := -1
	prevObstr := 0

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

/*

func main() {
	// Initialize elevator manager
	elevatorManager := ElevatorManager{
		ID:        1, // Change this per elevator
		Elevators: make(map[int]*Elevator),
	}

	// Start network listener
	go network.ListenForUpdates(&elevatorManager)

	// Start periodic state broadcast
	myElevator := network.ElevatorState{
		ID:        elevatorManager.ID,
		Floor:     0,
		Direction: "idle",
		Active:    true,
	}
	go network.retransmitState(myElevator)

	// Main loop to check for failures
	for {
		elevatorManager.DetectFailure()
		time.Sleep(2 * time.Second)
	}
}

go run elevator.go 1  # Starts the first elevator (ID=1)
go run elevator.go 2  # Starts the second elevator (ID=2)
go run elevator.go 3  # Starts the third elevator (ID=3)


*/
