package handler

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/minor-industries/codelab/cmd/bike/database"
	"github.com/minor-industries/codelab/cmd/bike/parser"
	"github.com/minor-industries/codelab/cmd/bike/schema"
	"github.com/minor-industries/platform/common/broker"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
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
	broker *broker.Broker,
) (*BikeHandler, error) {
	allSeries, err := database.LoadAllSeries(db)
	if err != nil {
		return nil, errors.Wrap(err, "load series")
	}

	return &BikeHandler{
		db:     db,
		series: allSeries,
		t0:     time.Now(),
		cancel: cancel,
		ctx:    ctx,
		broker: broker,
	}, nil
}

func (h *BikeHandler) Handle(t time.Time, msg []byte) error {
	h.lastMsg = t
	dt := t.Sub(h.t0).Seconds()

	fmt.Printf("%7.2f bikedata: %s\n", dt, hex.EncodeToString(msg))
	data := parser.ParseIndoorBikeData(msg)

	// store to database
	var err error
	data.AllPresentFields(func(seriesName string, value float64) {
		series, ok := h.series[seriesName]
		if !ok {
			panic(fmt.Errorf("unknown database series: %s", seriesName))
		}
		tx := h.db.Create(&database.Value{
			ID:        uuid.New(),
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
