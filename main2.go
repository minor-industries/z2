package main

import (
	"context"
	"fmt"
	"github.com/minor-industries/codelab/cmd/bike/source"
	"time"

	"tinygo.org/x/bluetooth"
)

// const address = "FC:38:34:32:0D:69"
const address = "5f953f6e-382c-39f7-c5ce-643b6141c967"

func main() {
	var adapter = bluetooth.DefaultAdapter

	fmt.Println("enabling")

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	t0 := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	handler := &bikeHandler{
		t0:     t0,
		cancel: cancel,
		ctx:    ctx,
	}
	go handler.Monitor()

	err := source.Run(ctx, adapter, address, map[source.CBKey]func([]byte){
		source.CBKey{
			bluetooth.ServiceUUIDFitnessMachine,
			bluetooth.CharacteristicUUIDIndoorBikeData,
		}: handler.Handle,
	})

	fmt.Println("run exited, error:", err)
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
