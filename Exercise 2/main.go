package main

// IP: 10.100.23.204

import (
	"fmt"
	"log"
	"net"
)

func receiving(portNum int) {

	address := fmt.Sprintf(":%d", portNum)
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil{
		log.Fatal("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil{
		log.Fatalf("Failed to start UDP listener: %v", err)
	}
	defer conn.Close()

	log.Printf("Listening on UDP port %d...", portNum)

	buffer := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Print("Error reading from UDP: %v", err)
			continue
		}

		log.Print("Received message '%s' from %s", string(buffer[:n]),addr)
	}
}

func sending(portNum int){

	// format port number to string in proper format
	address := fmt.Sprintf(":%d", portNum)

	// Resolve the UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	// About to send a message
	log.Printf("Sending message to UDP port %d...", portNum)

	// Message to send
	message := "Hello from sending(table 15)!"

	// Send the message
	_, err = conn.Write([]byte(message))
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	} else {
		log.Printf("Message sent: '%s'", message)
	}


}

func main() {

	go receiving(30000)

	go sending(20015)

}
