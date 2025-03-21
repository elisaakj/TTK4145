package main

import (
	"Driver-go/elevator-system/config"
	"Driver-go/elevator-system/elevatorStateMachine"
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/syncElev"
	"Network-go/network/bcast"
	"Network-go/network/peers"
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
		idInt = config.DEFAULT_ID // Default ID if conversion fails
	}

	//simFromHome := "172.26.129.47:20101"
	addr := fmt.Sprintf("172.26.129.47:%d", 20100+idInt)
	//simFromHome := "localhost:15657"
	//ddr := fmt.Sprintf("localhost:%d", config.BASE_PORT+idInt)
	elevio.Init(addr, config.NUM_FLOORS)

	// channels for syncElev
	chNewLocalOrder := make(chan elevio.ButtonEvent, 100)
	chMsgFromUDP := make(chan []config.SyncElevator, 100)
	chMsgToUDP := make(chan []config.SyncElevator, 100)
	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerTx := make(chan bool)

	// watchdog

	// channels for comm between syncElev and local elevator
	chClearLocalHallOrders := make(chan bool)
	chOrderToLocal := make(chan elevio.ButtonEvent, 100)
	chNewLocalState := make(chan elevatorStateMachine.Elevator, 100)

	// channels for local elevator
	chArrivedAtFloor := make(chan int)
	chObstruction := make(chan bool)

	// goroutines for local elevator
	go elevio.PollButtons(chNewLocalOrder)
	go elevio.PollFloorSensor(chArrivedAtFloor)
	go elevio.PollObstructionSwitch(chObstruction)

	ch := elevatorStateMachine.FsmChannels{
		//OrderComplete: ,
		Elevator:       chNewLocalState,
		NewOrder:       chOrderToLocal,
		ArrivedAtFloor: chArrivedAtFloor,
		Obstruction:    chObstruction,
	}

	go elevatorStateMachine.RunElevator(ch, idInt, config.NUM_FLOORS, config.NUM_BUTTONS)

	// can change InitNetwork
	statePort := 20100
	go bcast.Transmitter(statePort, chMsgToUDP)
	go bcast.Receiver(statePort, chMsgFromUDP)
	go peers.Transmitter(15647, id, chPeerTx)
	go peers.Receiver(15647, chPeerUpdate)

	// go watchdog

	go syncElev.SyncElevators(id, chNewLocalOrder, chNewLocalState, chMsgFromUDP, chMsgToUDP,
		chOrderToLocal, chPeerUpdate, chClearLocalHallOrders)

	select {}

	/*
		ch := elevatorStateMachine.FsmChannels{
			OrderComplete:  make(chan int),
			Elevator:       make(chan elevatorStateMachine.Elevator),
			NewOrder:       make(chan elevio.ButtonEvent),
			ArrivedAtFloor: make(chan int),
			Obstruction:    make(chan bool),
		}

		//go elevatorStateMachine.RunElevator(ch, 1, 4, 3)
		go elevatorStateMachine.RunElevator(ch, idInt, config.NUM_FLOORS, config.NUM_BUTTONS)

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
				fmt.Printf("Received button event in main: %+v\n", buttonEvent) // Debugging
				ch.NewOrder <- buttonEvent

			case floor := <-drv_floors:
				fmt.Printf("Received floor sensor event: %d\n", floor)
				ch.ArrivedAtFloor <- floor

			case obstruction := <-drv_obstr:
				fmt.Printf("Received obstruction event: %t\n", obstruction)
				ch.Obstruction <- obstruction

			case elevator := <-ch.Elevator:
				fmt.Printf("Elevator state update: %+v\n", elevator)
			}
		}*/
}
