package handler

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/minor-industries/codelab/cmd/bike/database"
	"github.com/minor-industries/codelab/cmd/bike/parser"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type BleHandler struct {
	db     *gorm.DB
	series map[string]*database.Series
}

func (h *BleHandler) Handle(t time.Time, req []byte) error {
	datum := parser.ParseIndoorBikeData(req)
	var err error

	datum.AllPresentFields(func(seriesName string, value float64) {
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

	return err
}

func NewBleHandler(db *gorm.DB) (*BleHandler, error) {
	allSeries, err := database.LoadSeries(db)
	if err != nil {
		return nil, errors.Wrap(err, "load series")
	}

	handler := &BleHandler{
		db:     db,
		series: allSeries,
	}

	return handler, nil
}
