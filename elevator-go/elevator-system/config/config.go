package config

// consts for several modules should be defined here

const (
	NUM_FLOORS  = 4 //MAX_FLOORS?
	NUM_BUTTONS = 3

	INVALID_FLOOR      = -1
	DOOR_OPEN_DURATION = 3.0

	// disse er bare i main, men uryddig å ha de definert der.. hmm
	DEFAULT_ID = 1
	BASE_PORT  = 20100
)

type ElevatorState int

const (
	IDLE ElevatorState = iota
	DOOR_OPEN
	MOVING
	UNAVAILABLE
)

// Below is for syncElev, will fix structure later

type Direction int
type RequestState int
type Behaviour int

const (
	Up   Direction = 1
	Down Direction = -1
	Stop Direction = 0
)

const (
	None      RequestState = 0
	Order     RequestState = 1
	Confirmed RequestState = 2
	Complete  RequestState = 3
)

type SyncElevator struct {
	ID       string
	Floor    int
	Dir      Direction
	Requests [][]RequestState
	OrderID  [][]int
	Behave   Behaviour
}
