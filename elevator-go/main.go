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
	//  prevObstr := 0

	if input.floorSensor() == -1 {
		fsmOnInitBetweenFloors()
	}

	prevRequests := make([][]int, N_FLOORS)
	for i := range prevRequests {
		prevRequests[i] = make([]int, N_BUTTONS)
	}

	prevFloor := -1

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

		// Obstruction sensor
		/*
			obstr := input.obstruction()
			if obstr != 0 && prevObstr == 0 { // Trigger only when obstruction starts
				fmt.Println("Obstruction detected! Keeping doors open.")
				fsmOnObstruction()
			} else if obstr == 0 && prevObstr != 0 { // Detect when obstruction is cleared
				fmt.Println("Obstruction cleared. Resuming operation.")
				fsmOnObstructionCleared()
			}
			prevObstr = obstr
		*/

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
