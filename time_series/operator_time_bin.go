package time_series

import (
	"github.com/minor-industries/rtgraph/schema"
	"time"
)

type OpTimeBin struct {
	tCurrent time.Time
}

func (o *OpTimeBin) ProcessNewValues(values []schema.Value) []schema.Value {
	result := make([]schema.Value, 0, len(values))

	for _, v := range values {
		tTrunc := v.Timestamp.Truncate(time.Second) // TODO: configurable
		if tTrunc != o.tCurrent {
			result = append(result, schema.Value{
				Timestamp: tTrunc,
				Value:     v.Value,
			})
		}
		o.tCurrent = tTrunc
	}

	return result
}
