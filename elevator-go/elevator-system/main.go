package main

import (
	"Driver-go/elevator-system/common"
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

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// 	Ensure ID is provided, convert to int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error: Invalid ID format, using default ID 1")
		idInt = common.DEFAULT_ID
	}

	addr := "localhost:15657"
	elevio.Init(addr, common.NUM_FLOORS)

	// channels for syncElev
	chNewLocalOrder := make(chan common.ButtonEvent, 100)
	chMsgFromUDP := make(chan []common.SyncElevator, 100)
	chMsgToUDP := make(chan []common.SyncElevator, 100)
	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerTx := make(chan bool)

	// channels for comm between syncElev and local elevator
	chClearLocalHallOrders := make(chan bool)
	chOrderToLocal := make(chan common.ButtonEvent, 100)
	chNewLocalState := make(chan common.Elevator, 100)

	// channels for local elevator
	chArrivedAtFloor := make(chan int)
	chObstruction := make(chan bool)
	chStuckElevator := make(chan int, 10)

	// goroutines for local elevator
	go elevio.PollButtons(chNewLocalOrder)
	go elevio.PollFloorSensor(chArrivedAtFloor)
	go elevio.PollObstructionSwitch(chObstruction)

	ch := common.FsmChannels{
		Elevator:       chNewLocalState,
		NewOrder:       chOrderToLocal,
		ArrivedAtFloor: chArrivedAtFloor,
		Obstruction:    chObstruction,
		StuckElevator:  chStuckElevator,
	}

	go elevatorStateMachine.RunElevator(ch, idInt)

	go bcast.Transmitter(20100, chMsgToUDP)
	go bcast.Receiver(20100, chMsgFromUDP)
	go peers.Transmitter(20200, id, chPeerTx)
	go peers.Receiver(20200, chPeerUpdate)

	go syncElev.SyncElevators(id, chNewLocalOrder, chNewLocalState, chMsgFromUDP, chMsgToUDP,
		chOrderToLocal, chPeerUpdate, chClearLocalHallOrders)

	select {}
}
