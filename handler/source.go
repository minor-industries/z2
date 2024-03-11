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

func (b BikeSource) Convert(msg source.Message) []source.Value {
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

func (b BikeSource) SubscriptionFilter(
	service bluetooth.UUID,
	characteristic bluetooth.UUID,
) bool {
	return service == bluetooth.ServiceUUIDFitnessMachine &&
		characteristic == bluetooth.CharacteristicUUIDIndoorBikeData
}
