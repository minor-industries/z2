package heartrate

import (
	"github.com/minor-industries/z2/source"
)

type Source struct{}

func (s *Source) Convert(msg source.Message) []source.Value {
	return []source.Value{{
		Name:      "heartrate",
		Timestamp: msg.Timestamp,
		Value:     float64(msg.Msg[1]),
	}}
}

func (s *Source) Services() []source.UUID {
	return []source.UUID{"0000180d-0000-1000-8000-00805f9b34fb"}
}

func (s *Source) Characteristics() []source.UUID {
	return []source.UUID{"00002a37-0000-1000-8000-00805f9b34fb"}
}
