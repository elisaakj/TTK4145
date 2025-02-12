package main

import (
	"encoding/binary"
	"log"
	"net"
	"os/exec"
	"time"
)

var counter uint64
var buffer = make([]byte, 16)

func spawnBackupForWindows() {
	// Command for windows

	/*err := exec.Command("cmd", "/C", "start", "powershell", "go", "run", "spawnWindow.go").Run()

	if err != nil {
		log.Fatal(err)
	}*/

	// Command for linux

	exec.Command("gnome-terminal", "--", "go", "run", "spawnWindow.go").Run()
}

func main() {

	isPrimary := false

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9999")
	if err != nil {
		log.Println("No connection")
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("I'm backup")
	}

	// backup loop
	for !isPrimary {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			isPrimary = true
		} else {
			counter = binary.LittleEndian.Uint64(buffer[:n]) // counter should have the value as the buffer

		}
	}

	conn.Close()

	spawnBackupForWindows()
	log.Println("I'm primary")
	bcastConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Printf("Failed to create UDP connection: %v", err)
	}

	for i := 0; i < 5; i++ {

		if i == 0 && counter == 0 {
			log.Println("\t| Starting at: ", counter, "\t|")
		} else if i == 0 {
			log.Println("\t| Continuing from number: ", counter, "\t|")
		} else {
			log.Println("\t| Number: ", counter, "\t|")
		}

		// writes incremented counter in the correct format
		counter++
		binary.LittleEndian.PutUint64(buffer, counter)
		_, err = bcastConn.Write([]byte(buffer))

		time.Sleep(1 * time.Second)

	}
}