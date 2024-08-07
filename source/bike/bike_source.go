package bike

import (
	"github.com/minor-industries/z2/source"
	"github.com/minor-industries/z2/source/bike/parser"
	"time"
)

var t0 time.Time

func init() {
	t0 = time.Now()
}

type BikeSource struct{}

func (b *BikeSource) Services() []source.UUID {
	return []source.UUID{"00001826-0000-1000-8000-00805f9b34fb"}
}

func (b *BikeSource) Characteristics() []source.UUID {
	return []source.UUID{"00002ad2-0000-1000-8000-00805f9b34fb"}
}

func (b *BikeSource) Convert(msg source.Message) []source.Value {
	// TODO: perhaps this filtering should be in the caller
	if msg.Service != b.Services()[0] {
		return nil
	}

	if msg.Characteristic != b.Characteristics()[0] {
		return nil
	}
	//dt := msg.Timestamp.Sub(t0).Seconds()
	//fmt.Printf("%7.2f bikedata: %s\n", dt, hex.EncodeToString(msg.Msg))
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
