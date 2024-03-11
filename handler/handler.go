package handler

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/minor-industries/codelab/cmd/z2/database"
	"github.com/minor-industries/codelab/cmd/z2/parser"
	"github.com/minor-industries/codelab/cmd/z2/schema"
	"github.com/minor-industries/platform/common/broker"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
	"tinygo.org/x/bluetooth"
)

// TODO: might be nice to split into separate handlers, one for DB and one for publishing to broker

type BikeHandler struct {
	db     *gorm.DB
	series map[string]*database.Series

	t0      time.Time
	lastMsg time.Time
	cancel  context.CancelFunc
	ctx     context.Context
	broker  *broker.Broker
}

func NewBikeHandler(
	db *gorm.DB,
	cancel context.CancelFunc,
	ctx context.Context,
	allSeries map[string]*database.Series,
	broker *broker.Broker,
) (*BikeHandler, error) {
	return &BikeHandler{
		db:     db,
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
	dt := t.Sub(h.t0).Seconds()

	fmt.Printf("%7.2f bikedata: %s\n", dt, hex.EncodeToString(msg))
	data := parser.ParseIndoorBikeData(msg)

	// store raw messages to database
	tx := h.db.Create(&database.RawValue{
		ID:               database.RandomID(),
		ServiceID:        service.String(),
		CharacteristicID: characteristic.String(),
		Timestamp:        t,
		Message:          msg,
	})
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "create raw value")
	}

	// store series to database
	var err error
	data.AllPresentFields(func(seriesName string, value float64) {
		series, ok := h.series[seriesName]
		if !ok {
			panic(fmt.Errorf("unknown database series: %s", seriesName))
		}
		tx := h.db.Create(&database.Value{
			ID:        database.RandomID(),
			Timestamp: t,
			Value:     value,
			Series:    series,
		})
		if tx.Error != nil {
			err = errors.Wrap(tx.Error, "create value")
			return
		}
	})
	if err != nil {
		return errors.Wrap(err, "store db")
	}

	// publish to broker
	// perhaps we should send these in bulk to the broker
	data.AllPresentFields(func(series string, value float64) {
		h.broker.Publish(&schema.Series{
			SeriesName: series,
			Timestamp:  t,
			Value:      value,
		})
	})

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
