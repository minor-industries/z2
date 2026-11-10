package time_series

import (
	"github.com/minor-industries/rtgraph/schema"
	"time"
)

// OpWarmup drops values until they have stayed at or above threshold for
// sustain, then until delay has passed since they first rose above it.
type OpWarmup struct {
	threshold float64
	delay     time.Duration
	sustain   time.Duration
	tAbove    time.Time
	latched   bool
}

func (o *OpWarmup) ProcessNewValues(values []schema.Value) []schema.Value {
	result := make([]schema.Value, 0, len(values))

	for _, v := range values {
		if !o.latched {
			if v.Value < o.threshold {
				o.tAbove = time.Time{}
				continue
			}
			if o.tAbove.IsZero() {
				o.tAbove = v.Timestamp
			}
			if v.Timestamp.Sub(o.tAbove) < o.sustain {
				continue
			}
			o.latched = true
		}

		if v.Timestamp.Sub(o.tAbove) >= o.delay {
			result = append(result, v)
		}
	}

	return result
}
