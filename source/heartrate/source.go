package heartrate

import (
	"github.com/minor-industries/z2/source"
	"tinygo.org/x/bluetooth"
)

type Source struct{}

func (s *Source) Convert(msg source.Message) []source.Value {
	//TODO implement me
	panic("implement me")
}

func (s *Source) Services() []bluetooth.UUID {
	return []bluetooth.UUID{
		bluetooth.ServiceUUIDHeartRate,
	}
}

func (s *Source) Characteristics() []bluetooth.UUID {
	return []bluetooth.UUID{
		bluetooth.CharacteristicUUIDHeartRateMeasurement,
	}
}
