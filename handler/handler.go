package handler

import (
	"context"
	"fmt"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/data"
	"github.com/minor-industries/z2/source"
	"github.com/pkg/errors"
	"time"
	"tinygo.org/x/bluetooth"
)

type Backends struct {
	Samples   *database.Backend
	RawValues *database.Backend
}

type Handler struct {
	graph    *rtgraph.Graph
	backends Backends
	source   source.Source
	cancel   context.CancelFunc
	ctx      context.Context

	t0             time.Time
	lastMsg        time.Time
	writeRawValues bool
}

func NewBikeHandler(
	graph *rtgraph.Graph,
	backends Backends,
	source source.Source,
	writeRawValues bool,
	cancel context.CancelFunc,
	ctx context.Context,
) (*Handler, error) {
	h := &Handler{
		graph:          graph,
		backends:       backends,
		source:         source,
		cancel:         cancel,
		ctx:            ctx,
		t0:             time.Now(),
		lastMsg:        time.Time{},
		writeRawValues: writeRawValues,
	}

	return h, nil
}

func (h *Handler) Handle(
	t time.Time,
	service bluetooth.UUID,
	characteristic bluetooth.UUID,
	msg []byte,
) error {
	h.lastMsg = t

	if h.writeRawValues {
		h.backends.RawValues.Insert(&data.RawValue{
			ID:               database.RandomID(),
			ServiceID:        service.String(),
			CharacteristicID: characteristic.String(),
			Timestamp:        t,
			Message:          msg,
		})
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

func (h *Handler) Monitor() {
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
