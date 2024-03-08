package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"
)

type bikeHandler struct {
	t0      time.Time
	lastMsg time.Time
	cancel  context.CancelFunc
	ctx     context.Context
}

func (h *bikeHandler) Handle(msg []byte) {
	h.lastMsg = time.Now()
	dt := time.Now().Sub(h.t0).Seconds()
	fmt.Printf("%7.2f bikedata: %s\n", dt, hex.EncodeToString(msg))
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
