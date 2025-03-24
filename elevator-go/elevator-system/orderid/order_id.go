package orderid

import (
	"Driver-go/elevator-system/config"
	"encoding/json"
	"fmt"
	"os"
)

const (
	filePath = "orderid_store.json"
)

var store [config.NUM_FLOORS][config.NUM_BUTTONS]int

func Load() error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("No existing orderID store found.")
		return nil // First startup is fine
	}
	return json.Unmarshal(data, &store)
}

func saveToFile() error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func UpdateIfGreater(floor, button, newID int) {
	if newID > store[floor][button] {
		store[floor][button] = newID
		saveToFile()
	}
}

func IncrementAndGet(floor, button int) int {
	store[floor][button]++
	saveToFile()
	return store[floor][button]
}

func DebugPrint() {
	fmt.Println("===== OrderID Store =====")
	for f := 0; f < config.NUM_FLOORS; f++ {
		for b := 0; b < config.NUM_BUTTONS; b++ {
			fmt.Printf("(%d,%d): %d\t", f, b, store[f][b])
		}
		fmt.Println()
	}
	fmt.Println("=========================")
}
