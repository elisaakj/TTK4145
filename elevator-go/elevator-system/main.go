package main

import (
	"Driver-go/elevator-system/communication"
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

	// 	Ensure ID is provided, convert to int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error: Invalid ID format, using default ID 1")
		idInt = 1 // Default ID if conversion fails
	}

	numFloors := 4
	numButtons := 3

	//simFromHome := "172.26.129.47:20101"
	addr := fmt.Sprintf("172.26.129.47:%d", 20100+idInt)
	//simFromHome := "localhost:15657"
	//addr := fmt.Sprintf("localhost:%d", 15555+idInt)
	elevio.Init(addr, numFloors)

	ch := elevatorStateMachine.FsmChannels{
		OrderComplete:  make(chan int),
		Elevator:       make(chan elevatorStateMachine.Elevator),
		NewOrder:       make(chan elevio.ButtonEvent),
		ArrivedAtFloor: make(chan int),
		Obstruction:    make(chan bool),
	}

	//go elevatorStateMachine.RunElevator(ch, 1, 4, 3)
	go elevatorStateMachine.RunElevator(ch, idInt, numFloors, numButtons)

	// want to have the our looking like something like this below
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	//drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	//go elevio.PollStopButton(drv_stop)

	updateChannel := make(chan communication.ElevatorState)
	communication.InitNetwork(idInt, updateChannel)

	for {
		select {
		case buttonEvent := <-drv_buttons:
			fmt.Printf("📥 Received button event in main: %+v\n", buttonEvent) // Debugging
			ch.NewOrder <- buttonEvent

		case floor := <-drv_floors:
			fmt.Printf("📥 Received floor sensor event: %d\n", floor)
			ch.ArrivedAtFloor <- floor

		case obstruction := <-drv_obstr:
			fmt.Printf("📥 Received obstruction event: %t\n", obstruction)
			ch.Obstruction <- obstruction

		case elevator := <-ch.Elevator:
			fmt.Printf("Elevator state update: %+v\n", elevator)
		}
	}
}
