package handler

import (
	"context"
	"fmt"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/data"
	"github.com/minor-industries/z2/source"
	"github.com/pkg/errors"
	"time"
)

type Backends struct {
	Samples   *sqlite.Backend
	RawValues *sqlite.Backend
}

type Handler struct {
	graph     *rtgraph.Graph
	rawValues *sqlite.Backend
	source    source.Source
	cancel    context.CancelFunc
	ctx       context.Context

	t0             time.Time
	lastMsg        time.Time
	writeRawValues bool
}

func NewHandler(
	graph *rtgraph.Graph,
	rawValues *sqlite.Backend,
	source source.Source,
	writeRawValues bool,
	cancel context.CancelFunc,
	ctx context.Context,
) *Handler {
	return &Handler{
		graph:          graph,
		rawValues:      rawValues,
		source:         source,
		cancel:         cancel,
		ctx:            ctx,
		t0:             time.Now(),
		lastMsg:        time.Time{},
		writeRawValues: writeRawValues,
	}
}

func (h *Handler) Handle(
	t time.Time,
	service source.UUID,
	characteristic source.UUID,
	msg []byte,
) error {
	h.lastMsg = t

	if h.writeRawValues {
		h.rawValues.Insert(&data.RawValue{
			ID:               sqlite.RandomID(),
			ServiceID:        string(service),
			CharacteristicID: string(characteristic),
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
