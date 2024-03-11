package handler

import (
	"encoding/hex"
	"fmt"
	"github.com/minor-industries/codelab/cmd/z2/parser"
	"github.com/minor-industries/codelab/cmd/z2/source"
	"time"
	"tinygo.org/x/bluetooth"
)

var t0 time.Time

func init() {
	t0 = time.Now()
}

type BikeSource struct{}

func (b *BikeSource) Services() []bluetooth.UUID {
	return []bluetooth.UUID{bluetooth.ServiceUUIDFitnessMachine}
}

func (b *BikeSource) Characteristics() []bluetooth.UUID {
	return []bluetooth.UUID{bluetooth.CharacteristicUUIDIndoorBikeData}
}

func (b *BikeSource) Convert(msg source.Message) []source.Value {
	dt := msg.Timestamp.Sub(t0).Seconds()

	fmt.Printf("%7.2f bikedata: %s\n", dt, hex.EncodeToString(msg.Msg))
	data := parser.ParseIndoorBikeData(msg.Msg)

	var result []source.Value
	data.AllPresentFields(func(series string, value float64) {
		result = append(result, source.Value{
			Name:      series,
			Timestamp: msg.Timestamp,
			Value:     value,
		})
	})

	return result
}
