package main

import (
	"Driver-go/elevator-system/elevatorStateMachine"
	"Driver-go/elevator-system/elevio"
	"flag"
	"fmt"
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

	numFloors := 4
	numButtons := 3

	addr := fmt.Sprintf("localhost:%d", 15555+idInt)
	elevio.Init(addr, numFloors)

	ch := elevatorStateMachine.stateMachineChannels{
		orderComplete:  make(chan int),
		elevator:       make(chan elevatorStateMachine.Elevator),
		newOrder:       make(chan elevio.ButtonType),
		arrivedAtFloor: make(chan int),
		obstruction:    make(chan bool),
	}

	elevatorStateMachine.RunElevator(ch, idInt, numFloors, numButtons)

	// want to have the our looking like something like this below
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	//updateChannel := make(chan communication.ElevatorState)
	//communication.InitNetwork(idInt, updateChannel)

	for {
		select {
		case buttonEvent := <-drv_buttons:
			ch.newOrder <- buttonEvent

		case floor := <-drv_floors:
			ch.arrivedAtFloor <- floor

		case obstrution := <-drv_obstr:
			ch.obstrution <- obstrution

			//case stop := <-drv_stop:
			// do not need to implement
		}
	}
}
