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



}

func main() {

	go receiving(30000)

	go sending(20015)

}
