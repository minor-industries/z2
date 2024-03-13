package handler

import (
	"context"
	"fmt"
	database2 "github.com/minor-industries/codelab/cmd/z2/rtgraph/database"
	"github.com/minor-industries/codelab/cmd/z2/rtgraph/schema"
	"github.com/minor-industries/codelab/cmd/z2/source"
	"github.com/minor-industries/platform/common/broker"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
	"tinygo.org/x/bluetooth"
)

// TODO: might be nice to split into separate handlers, one for DB and one for publishing to broker

type BikeHandler struct {
	db     *gorm.DB
	series map[string]*database2.Series
	source source.Source

	t0      time.Time
	lastMsg time.Time
	cancel  context.CancelFunc
	ctx     context.Context
	broker  *broker.Broker
}

func NewBikeHandler(
	db *gorm.DB,
	source source.Source,
	cancel context.CancelFunc,
	ctx context.Context,
	allSeries map[string]*database2.Series,
	broker *broker.Broker,
) (*BikeHandler, error) {
	return &BikeHandler{
		db:     db,
		source: source,
		series: allSeries,
		t0:     time.Now(),
		cancel: cancel,
		ctx:    ctx,
		broker: broker,
	}, nil
}

func (h *BikeHandler) Handle(
	t time.Time,
	service bluetooth.UUID,
	characteristic bluetooth.UUID,
	msg []byte,
) error {
	h.lastMsg = t

	// store raw messages to database
	tx := h.db.Create(&database2.RawValue{
		ID:               database2.RandomID(),
		ServiceID:        service.String(),
		CharacteristicID: characteristic.String(),
		Timestamp:        t,
		Message:          msg,
	})
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "create raw value")
	}

	srs := h.source.Convert(source.Message{
		Timestamp:      t,
		Service:        service,
		Characteristic: characteristic,
		Msg:            msg,
	})

	for _, s := range srs {
		series, ok := h.series[s.Name]
		if !ok {
			return fmt.Errorf("unknown database series: %s", s.Name)
		}
		tx := h.db.Create(&database2.Value{
			ID:        database2.RandomID(),
			Timestamp: t,
			Value:     s.Value,
			Series:    series,
		})
		if tx.Error != nil {
			return errors.Wrap(tx.Error, "create value")
		}
	}

	for _, s := range srs {
		h.broker.Publish(&schema.Series{
			SeriesName: s.Name,
			Timestamp:  t,
			Value:      s.Value,
		})
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
