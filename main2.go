package main

import (
	"context"
	"fmt"
	"github.com/minor-industries/codelab/cmd/bike/schema"
	"github.com/minor-industries/codelab/cmd/bike/source"
	"github.com/minor-industries/codelab/cmd/bike/source/replay"
	"github.com/minor-industries/platform/common/broker"
	"time"

	"tinygo.org/x/bluetooth"
)

func main() {
	br := broker.NewBroker()
	go br.Start()

	ctx, cancel := context.WithCancel(context.Background())
	handler := &bikeHandler{
		t0:     time.Now(),
		cancel: cancel,
		ctx:    ctx,
		broker: br,
	}
	go handler.Monitor()

	go func() {
		ch := br.Subscribe()
		for msg := range ch {
			switch _ := msg.(type) {
			case *schema.Series:
				//fmt.Println(m.SeriesName, m.Timestamp, m.Value)
			}
		}
	}()

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
