package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/minor-industries/codelab/cmd/bike/parser"
	"github.com/minor-industries/codelab/cmd/bike/schema"
	"github.com/minor-industries/platform/common/broker"
	"time"
)

type bikeHandler struct {
	t0      time.Time
	lastMsg time.Time
	cancel  context.CancelFunc
	ctx     context.Context
	broker  *broker.Broker
}

func (h *bikeHandler) Handle(msg []byte) {
	h.lastMsg = time.Now()
	dt := time.Now().Sub(h.t0).Seconds()
	now := time.Now()
	fmt.Printf("%7.2f bikedata: %s\n", dt, hex.EncodeToString(msg))
	data := parser.ParseIndoorBikeData(msg)

	// perhaps we should send these in bulk to the broker
	data.AllPresentFields(func(series string, value float64) {
		h.broker.Publish(&schema.Series{
			SeriesName: series,
			Timestamp:  now,
			Value:      value,
		})
	})
}

func (h *bikeHandler) Monitor() {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if h.lastMsg.Equal(time.Time{}) {
				break
			}
			if time.Now().Sub(h.lastMsg) > 10*time.Second {
				fmt.Println("timed out waiting for message")
				// TODO: also disable Handle from processing future messages
				h.cancel()
				return
			}
		case <-h.ctx.Done():
			return
		}
	}
}
