package main

import (
	"github.com/minor-industries/rtgraph/schema"
	"github.com/minor-industries/z2/variables"
)

type OpGate struct {
	target string
	vars   *variables.Cache
	open   bool
}

func (o *OpGate) ProcessNewValues(values []schema.Value) []schema.Value {
	result := make([]schema.Value, 0, len(values))

	target, ok := o.vars.GetOne(o.target)
	if !ok {
		// TODO: perhaps need a way to signal errors
		return nil
	}

	// TODO: close when appropriate

	for _, v := range values {
		if v.Value > target {
			o.open = true
		}
		if o.open {
			result = append(result, v)
		}
	}

	return result
}
