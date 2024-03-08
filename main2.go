package main

import (
	"context"
	"fmt"
	"github.com/minor-industries/codelab/cmd/bike/source"
	"time"

	"tinygo.org/x/bluetooth"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	handler := &bikeHandler{
		t0:     time.Now(),
		cancel: cancel,
		ctx:    ctx,
	}
	go handler.Monitor()

	err := source.Run(ctx, address, map[source.CBKey]func([]byte){
		source.CBKey{
			bluetooth.ServiceUUIDFitnessMachine,
			bluetooth.CharacteristicUUIDIndoorBikeData,
		}: handler.Handle,
	})

	fmt.Println("run exited, error:", err)
}
