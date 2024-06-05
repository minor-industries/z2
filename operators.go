package main

import (
	"github.com/minor-industries/rtgraph/schema"
	"github.com/minor-industries/z2/variables"
)

type OpGate struct {
	target   string
	driftPct string
	vars     *variables.Cache
	open     bool
}

func (o *OpGate) ProcessNewValues(values []schema.Value) []schema.Value {
	result := make([]schema.Value, 0, len(values))

	target, ok := o.vars.GetOne(o.target)
	if !ok {
		// TODO: perhaps need a way to signal errors
		return nil
	}

	driftPct, ok := o.vars.GetOne(o.driftPct)
	if !ok {
		// TODO: perhaps need a way to signal errors
		return nil
	}

	driftedTarget := target * (1 - driftPct/100.0)

	// TODO: close when appropriate

	for _, v := range values {
		if v.Value > driftedTarget {
			o.open = true
		}
		if o.open {
			result = append(result, v)
		}
	}

	return result
}
