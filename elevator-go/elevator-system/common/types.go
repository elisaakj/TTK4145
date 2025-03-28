package common

type ElevatorState int
const (
	IDLE ElevatorState = iota
	DOOR_OPEN
	MOVING
	UNAVAILABLE
)

type FsmChannels struct {
	Elevator       chan Elevator
	NewOrder       chan ButtonEvent
	ArrivedAtFloor chan int
	Obstruction    chan bool
	StuckElevator  chan int
}

type DirnBehaviourPair struct {
	Dirn  MotorDirection
	State ElevatorState
}

type Elevator struct {
	ID         int
	Floor      int
	Dirn       MotorDirection
	Requests   [][]bool
	State      ElevatorState
	Obstructed bool
	OrderID    int
}

type MotorDirection int
const (
	DIRN_UP   MotorDirection = 1
	DIRN_STOP MotorDirection = 0
	DIRN_DOWN MotorDirection = -1
)

type ButtonType int
const (
	BUTTON_HALL_UP ButtonType = iota
	BUTTON_HALL_DOWN
	BUTTON_CAB
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type RequestState int
const (
	NONE RequestState = iota
	ORDER
	CONFIRMED
	COMPLETE
)

type OrderInfo struct {
	State   RequestState
	OrderID int
}

type SyncElevator struct {
	ID       string
	Floor    int
	Dirn     MotorDirection
	Requests [][]OrderInfo
	OrderID  int
	State    ElevatorState
}