package elevatorStateMachine

import (
	"time"
)

var (
	timerEndTime time.Time
	timerIsActive  bool
)

func timerStart(duration float64) {
	timerEndTime = time.Now().Add(time.Duration(duration * float64(time.Second)))
	timerIsActive = true
}

func timerStop() {
	timerIsActive = false
}

func timerTimedOut() bool {
	if timerIsActive && time.Now().After(timerEndTime) {
		timerIsActive = false
		return true
	}
	return false
	//return timerActive && getWallTime().After(timerEndTime)
}
