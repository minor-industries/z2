package app

import (
	"context"
	"fmt"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/lib/data"
	"github.com/minor-industries/z2/lib/source"
	"github.com/pkg/errors"
	"time"
)

type BluetoothIsConnected struct {
	DeviceInfo source.DeviceInfo
}

func (b *BluetoothIsConnected) Name() string {
	return "BluetoothIsConnected"
}

type BTHandler struct {
	graph     *rtgraph.Graph
	broker    *broker.Broker
	rawValues *sqlite.Backend
	source    source.Source
	cancel    context.CancelFunc

	ctx            context.Context
	t0             time.Time
	lastMsg        time.Time
	writeRawValues bool
}

func NewBTHandler(graph *rtgraph.Graph, br *broker.Broker, rawValues *sqlite.Backend, source source.Source, writeRawValues bool, cancel context.CancelFunc, ctx context.Context) *BTHandler {
	return &BTHandler{
		graph:          graph,
		broker:         br,
		rawValues:      rawValues,
		source:         source,
		cancel:         cancel,
		ctx:            ctx,
		t0:             time.Now(),
		lastMsg:        time.Time{},
		writeRawValues: writeRawValues,
	}
}

func (h *BTHandler) Handle(
	t time.Time,
	info source.DeviceInfo,
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

	if len(srs) > 0 {
		m := &BluetoothIsConnected{DeviceInfo: info}
		h.broker.Publish(m)
	}

	for _, s := range srs {
		err := h.graph.CreateValue(s.Name, s.Timestamp, s.Value)
		if err != nil {
			return errors.Wrap(err, "create value")
		}
	}

	return nil
}

func (h *BTHandler) Monitor(disconnect chan struct{}) {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			//TODO: h.lastMsg access here is not threadsafe
			if h.lastMsg.Equal(time.Time{}) {
				break
			}
			if time.Now().Sub(h.lastMsg) > 10*time.Second {
				fmt.Println("timed out waiting for message")
				// TODO: also disable Handle from processing future messages
				h.cancel()
				return
			}
		case <-disconnect:
			return
		case <-h.ctx.Done():
			return
		}
	}
}
