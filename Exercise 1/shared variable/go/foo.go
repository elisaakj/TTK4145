// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
	"time"
)

var i = 0

type messageType int

const (
    increment messageType = iota
    decrement
    get
)

func incrementing(ch chan<- messageType, done chan<- bool) {
    //TODO: increment i 1000000 times
    for j := 0; j < 1000000; j++ {
        ch <- increment
    }
    done <- true
}

func decrementing(ch chan<- messageType, done chan<- bool) {
    //TODO: decrement i 1000000 times
    for j := 0; j < 1000000; j++ {
        ch <- decrement
    }
    done <- true
}

func server(ch <-chan messageType, resultCh chan<- int) {
    for msg := range ch {
        switch msg{
        case increment:
            i++
        case decrement:
            i--
        case get:
            resultCh <- i
        }
    }
}

func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    runtime.GOMAXPROCS(2)
    // GOMAXPROCS limits the number of threads that can execute goroutines simultaneously
	// Setting it to 1 -> only does the last goroutine

    // Definerer en channel for kommunikasjon med server
    ch := make(chan messageType)
    resultCh := make(chan int)
    done := make(chan bool, 2)

    go server(ch, resultCh)

    // TODO: Spawn both functions as goroutines
	go incrementing(ch, done)
    go decrementing(ch, done)

    // venter på fullføring
    <-done
    <-done

    ch <- get
    finalResult := <-resultCh

    Println("The final result is:", finalResult)

    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    time.Sleep(500*time.Millisecond)
    //Println("The magic number is:", i)
}