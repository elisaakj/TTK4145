package main

import (

	// "Network-go/network/bcast"
	// "Network-go/network/localip"
	// "Network-go/network/peers"

	"fmt"
	//"net"
	"time"
	// "./elevator_io_device.go"
	// "fsm"
	// "timer"
)

func main() {
	fmt.Println("Started!")

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

	for {
		// Request button
		for f := 0; f < N_FLOORS; f++ {
			for b := 0; b < N_BUTTONS; b++ {
				v := input.requestButton(f, Button(b))
				if v != 0 && v != prevRequests[f][b] {
					fsmOnRequestButtonPress(f, Button(b))
				}
				prevRequests[f][b] = v
			}
		}

		// Obstruction

		obstr := input.obstruction()
		if obstr != 0 && prevObstr == 0 {
			// If obstruction is detected and wasn't detected before
			fmt.Println("Obstruction detected! Keeping doors open.")
			fsmOnObstruction()
		} else if obstr == 0 && prevObstr != 0 {
			// If obstruction was cleared
			fmt.Println("Obstruction cleared. Resuming operation.")
			fsmOnObstructionCleared()
		}
		prevObstr = obstr // Update previous obstruction state

		// Floor sensor
		floor := input.floorSensor()
		if floor != -1 && floor != prevFloor {
			fsmOnFloorArrival(floor)
		}
		prevFloor = floor

		// Timer
		if timerTimedOut() {
			timerStop()
			fsmOnDoorTimeout()
		}

		time.Sleep(inputPollRate)
	}
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
