package rower

import (
	"github.com/minor-industries/z2/source"
	"time"
)

var (
	sv1 = source.UUID("ce060030-43e5-11e4-916c-0800200c9a66")

	ch1 = source.UUID("CE060031-43E5-11E4-916C-0800200C9A66")
	ch2 = source.UUID("CE060032-43E5-11E4-916C-0800200C9A66")
	ch3 = source.UUID("CE060036-43E5-11E4-916C-0800200C9A66")
)

type rowerSource struct {
	status Status
}

func NewRowerSource() source.Source {
	return &rowerSource{}
}

func (r *rowerSource) Services() []source.UUID {
	return []source.UUID{sv1}
}

func (r *rowerSource) Characteristics() []source.UUID {
	return []source.UUID{ch1, ch2, ch3}
}

func (r *rowerSource) Convert(msg source.Message) []source.Value {
	return convert(msg, &r.status)
}

func publish(pm5 *Status, timestamp time.Time) []source.Value {
	return []source.Value{
		{
			Name:      "rower_stroke_count",
			Timestamp: timestamp,
			Value:     float64(pm5.StrokeCount),
		},
		{
			Name:      "rower_power",
			Timestamp: timestamp,
			Value:     float64(pm5.Power),
		},
		{
			Name:      "rower_speed",
			Timestamp: timestamp,
			Value:     float64(pm5.Speed),
		},
		{
			Name:      "rower_spm",
			Timestamp: timestamp,
			Value:     float64(pm5.Spm),
		},
	}
}
