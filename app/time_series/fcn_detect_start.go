package time_series

import (
	"fmt"
	"github.com/gammazero/deque"
	"github.com/minor-industries/rtgraph/schema"
	"github.com/minor-industries/z2/lib/variables"
)

type FcnDetectWorkout struct {
	target        string
	driftPct      string
	vars          *variables.Cache
	workoutActive bool
}

func (f *FcnDetectWorkout) AddValue(v schema.Value) {}

func (f *FcnDetectWorkout) RemoveValue(v schema.Value) {}

func (f *FcnDetectWorkout) Compute(values *deque.Deque[schema.Value]) (float64, bool) {
	if values.Len() == 0 {
		f.toggle(false, schema.Value{}) // TODO: second argument
		return 0, false
	}

	target, ok := f.vars.GetOne(f.target)
	if !ok {
		panic("unknown variable") // TODO
	}

	driftPct, ok := f.vars.GetOne(f.driftPct)
	if !ok {
		panic("unknown variable") // TODO
	}

	driftedTarget := target * (1 - driftPct/100.0)

	for i := 0; i < values.Len(); i++ {
		s := values.At(i)
		value := s.Value
		if value >= driftedTarget {
			f.toggle(true, s)
		}
	}

	return 0, false // hmm, this just swallows all points, is that cool (maybe for now anyway)?
}

func (f *FcnDetectWorkout) toggle(b bool, s schema.Value) {
	if b {
		if !f.workoutActive {
			fmt.Println("detected start of workout", s.Timestamp, s.Value)
		}
	} else {
		if f.workoutActive {
			fmt.Println("detected end of workout", s.Timestamp, s.Value)
		}
	}
	f.workoutActive = b
}
