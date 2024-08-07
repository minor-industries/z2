package heartrate

import (
	"github.com/minor-industries/z2/source"
	"tinygo.org/x/bluetooth"
)

type Source struct{}

func (s *Source) Convert(msg source.Message) []source.Value {
	// TODO: perhaps this filtering should be in the caller
	if msg.Service != s.Services()[0] {
		return nil
	}

	if msg.Characteristic != s.Characteristics()[0] {
		return nil
	}

	return []source.Value{{
		Name:      "heartrate",
		Timestamp: msg.Timestamp,
		Value:     float64(msg.Msg[1]),
	}}
}

func (s *Source) Services() []source.UUID {
	return []source.UUID{
		source.UUID(bluetooth.ServiceUUIDHeartRate.String()),
	}
}

func (s *Source) Characteristics() []source.UUID {
	return []source.UUID{
		source.UUID(bluetooth.CharacteristicUUIDHeartRateMeasurement.String()),
	}
}
