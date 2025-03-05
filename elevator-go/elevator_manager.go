package main

import (
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

// ones again almost the same struct

func (em *ElevatorManager) ElectMaster() {
	minID := em.ID
	for id, elevator := range em.Elevators {
		if elevator.active && id < minID {
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
			if id != em.ID && elevator.active {
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
        if time.Since(elevator.lastSeen) > 3*time.Second {
            fmt.Printf("Elevator %d unresponsive!\n", id)
            elevator.active = false

            // Redistribute hall calls
            for f := 0; f < N_FLOORS; f++ {
                if elevator.requests[f][B_HallUp] {
                    em.AssignHallCall(f, "up")
                }
                if elevator.requests[f][B_HallDown] {
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
		if elevator.active {
			// quite ugly because of the 'math.Abs', can probably be fixed somehow
			var floorCalc = elevator.floor - floor
			var floorCalcDone float64 = float64(floorCalc)
			score := math.Abs(floorCalcDone) // Simple distance calc
			var scoreInt int = int(score)
			if scoreInt < bestScore {
				bestScore = scoreInt
				bestElevator = id
			}
		}
	}

	if bestElevator != -1 {
		fmt.Printf("Master assigning floor %d to elevator %d\n", floor, bestElevator)
	}
}

func (em *ElevatorManager) UpdateElevatorState(state ElevatorState) {
	elevator, exist := em.Elevators[state.ID]
	if !exist {
		elevator = &Elevator{}
		em.Elevators[state.ID] = elevator
	}

	elevator.floor = state.floor
	elevator.dirn = state.dirn
	elevator.requests = state.requests
	elevator.behaviour = state.behaviour
	elevator.lastSeen = time.Now()
	elevator.active = true  // Mark as active again if previously inactive


	fmt.Printf("Updated state from elevator %d: %+v\n", state.ID, elevator)

	// if master is down, elect again
	if state.ID == em.MasterID && !state.active {
		em.ElectMaster()
	}
}
