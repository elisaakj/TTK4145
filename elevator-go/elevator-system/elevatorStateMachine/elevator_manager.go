package elevatorStateMachine

/*
import (
	"Driver-go/elevator-system/communication"
	"Driver-go/elevator-system/elevio"
	"fmt"
	"math"
	"time"
)

type ElevatorManager struct {
	ID           int
	MasterID     int
	Elevators    map[int]*Elevator
	IsMaster     bool
	StateUpdated bool
}

// once again almost the same struct

func (em *ElevatorManager) ElectMaster() {
	minID := em.ID
	for id, elevator := range em.Elevators {
		if elevator.Active && id < minID {
			minID = id
		}
	}
	em.MasterID = minID
	em.IsMaster = (em.ID == em.MasterID)

	fmt.Printf("New master elected: %d (self: %t)\n", em.MasterID, em.IsMaster)
}

func (em *ElevatorManager) SyncState() {
	if em.IsMaster {
		for id, elevator := range em.Elevators {
			if id != em.ID && elevator.Active {
				em.StateUpdated = true
				fmt.Printf("Syncing state to slave: %d\n", id)
			}
		}
	}
}

// should probably not do like this, should detect failurs in network, and then elect and redistribute here

// DetectFailure identifies unresponsive elevators
func (em *ElevatorManager) DetectFailure() {
	for id, elevator := range em.Elevators {
		if time.Since(elevator.LastSeen) > 3*time.Second {
			fmt.Printf("Elevator %d unresponsive!\n", id)
			elevator.Active = false

			// Redistribute hall calls
			for f := 0; f < N_FLOORS; f++ {
				if elevator.Requests[f][elevio.BT_HallUp] {
					em.AssignHallCall(f, "up")
				}
				if elevator.Requests[f][elevio.BT_HallDown] {
					em.AssignHallCall(f, "down")
				}
			}

			// If master is down, elect again
			if id == em.MasterID {
				em.ElectMaster()
			}
		}
	}
}

func (em *ElevatorManager) AssignHallCall(floor int, direction string) {
	if !em.IsMaster {
		return
	}

	bestElevator := -1
	bestScore := 999

	for id, elevator := range em.Elevators {
		if elevator.Active {
			score := int(math.Abs(float64(elevator.Floor - floor)))
			if score < bestScore {
				bestScore = score
				bestElevator = id
			}
		}
	}

	if bestElevator != -1 {
		em.Elevators[bestElevator].Requests[floor][elevio.BT_HallUp] = true
	}
}

func (em *ElevatorManager) UpdateElevatorState(state communication.ElevatorState) {
	elevator, exist := em.Elevators[state.ID]
	if !exist {
		elevator = &Elevator{}
		em.Elevators[state.ID] = elevator
	}

	elevator.Floor = state.Floor
	elevator.Dirn = state.Dirn
	elevator.Requests = state.Requests
	elevator.Behaviour = state.Behaviour
	elevator.LastSeen = time.Now()
	elevator.Active = state.Active // Mark as active again if previously inactive

	fmt.Printf("Updated state from elevator %d: %+v\n", state.ID, elevator)

	// if master is down, elect again
	if state.ID == em.MasterID && !state.Active {
		em.DetectFailure()
	}
}
*/
