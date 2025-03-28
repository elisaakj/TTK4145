package orderid

import (
	"Driver-go/elevator-system/common"
	"encoding/json"
	"fmt"
	"os"
)

const (
	filePath = "orderid_store.json"
)

var store [common.NUM_FLOORS][common.NUM_BUTTONS]int

func Load(id string) error {
	fileWithId := filePath + "_" + id
	data, err := os.ReadFile(fileWithId)
	if err != nil {
		fmt.Println("No existing orderID store found.")
		return nil // First startup is fine
	}
	return json.Unmarshal(data, &store)
}

func saveToFile(id string) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	fileWithId := filePath + "_" + id
	return os.WriteFile(fileWithId, data, 0644)
}

func UpdateIfGreater(floor, button, newID int, id string) {
	if newID > store[floor][button] {
		store[floor][button] = newID
		saveToFile(id)
	}
}

func IncrementAndGet(floor, button int, id string) int {
	store[floor][button]++
	saveToFile(id)
	return store[floor][button]
}

func DebugPrint() {
	fmt.Println("===== OrderID Store =====")
	for f := 0; f < common.NUM_FLOORS; f++ {
		for b := 0; b < common.NUM_BUTTONS; b++ {
			fmt.Printf("(%d,%d): %d\t", f, b, store[f][b])
		}
		fmt.Println()
	}
	fmt.Println("=========================")
}
