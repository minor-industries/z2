package handler

import (
	"context"
	"fmt"
	"github.com/minor-industries/codelab/cmd/z2/source"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database"
	"github.com/pkg/errors"
	"time"
	"tinygo.org/x/bluetooth"
)

// TODO: might be nice to split into separate handlers, one for DB and one for publishing to broker

type BikeHandler struct {
	graph  *rtgraph.Graph
	source source.Source
	cancel context.CancelFunc
	ctx    context.Context

	t0      time.Time
	lastMsg time.Time

	dbCh chan *database.RawValue
}

func NewBikeHandler(
	graph *rtgraph.Graph,
	source source.Source,
	cancel context.CancelFunc,
	ctx context.Context,
) (*BikeHandler, error) {
	h := &BikeHandler{
		graph:   graph,
		source:  source,
		cancel:  cancel,
		ctx:     ctx,
		t0:      time.Now(),
		lastMsg: time.Time{},
		dbCh:    make(chan *database.RawValue, 256),
	}

	go h.dbWriter()

	return h, nil
}

func (h *BikeHandler) dbWriter() {
	ticker := time.NewTicker(100 * time.Millisecond)
	var values []*database.RawValue

	for {
		select {
		case <-ticker.C:
			if len(values) == 0 {
				continue
			}
			tx := h.graph.DB().Create(values)
			if tx.Error != nil {
				fmt.Println("error:", errors.Wrap(tx.Error, "writing raw values"))
			}
			values = nil
		case value := <-h.dbCh:
			values = append(values, value)
		}
	}
}

func (h *BikeHandler) Handle(
	t time.Time,
	service bluetooth.UUID,
	characteristic bluetooth.UUID,
	msg []byte,
) error {
	h.lastMsg = t

	// store raw messages to database
	h.dbCh <- &database.RawValue{
		ID:               database.RandomID(),
		ServiceID:        service.String(),
		CharacteristicID: characteristic.String(),
		Timestamp:        t,
		Message:          msg,
	}

	srs := h.source.Convert(source.Message{
		Timestamp:      t,
		Service:        service,
		Characteristic: characteristic,
		Msg:            msg,
	})

	for _, s := range srs {
		err := h.graph.CreateValue(s.Name, s.Timestamp, s.Value)
		if err != nil {
			return errors.Wrap(err, "create value")
		}
	}

	return nil
}

func (h *BikeHandler) Monitor() {
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
