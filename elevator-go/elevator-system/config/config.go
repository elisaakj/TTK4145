package config

const (
	NUM_FLOORS  = 4
	NUM_BUTTONS = 3

	INVALID_FLOOR      = -1
	DOOR_OPEN_DURATION = 3

	CONNECTION_TIMER  = 3
	STUCK_TIMER       = 5
	OBSTRUCTION_TIMER = 5

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

type Direction int
type RequestState int
type State int

const (
	None      RequestState = 0
	Order     RequestState = 1
	Confirmed RequestState = 2
	Complete  RequestState = 3
)

type OrderInfo struct {
	State   RequestState
	OrderID int
}

type SyncElevator struct {
	ID       string
	Floor    int
	Dirn     Direction
	Requests [][]OrderInfo
	OrderID  int
	State    State
}
