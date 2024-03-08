package main

import (
	"context"
	"fmt"
	"github.com/minor-industries/codelab/cmd/bike/source"
	"github.com/minor-industries/codelab/cmd/bike/source/replay"
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

	var err error
	if false {
		err = source.Run(ctx, address, map[source.CBKey]func([]byte){
			source.CBKey{
				bluetooth.ServiceUUIDFitnessMachine,
				bluetooth.CharacteristicUUIDIndoorBikeData,
			}: handler.Handle,
		})
	} else {
		err = replay.Run(ctx, "raw.txt", handler.Handle)
	}

	fmt.Println("run exited, error:", err)
}
