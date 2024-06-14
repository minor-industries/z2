package app

import (
	"fmt"
	"github.com/minor-industries/rtgraph/computed_series"
	"github.com/minor-industries/rtgraph/schema"
	"github.com/minor-industries/z2/variables"
	"github.com/pkg/errors"
	"time"
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

func (app *App) setupGraphFunctions() {
	app.Graph.Parser.AddFunction("mygate", func(start time.Time, args []string) (computed_series.Operator, error) {
		if len(args) != 2 {
			return nil, errors.New("mygate function requires 2 arguments")
		}

		vars := app.vars.Get(args)
		for _, v := range vars {
			if !v.Present {
				return nil, fmt.Errorf("variable %s not found", v.Name)
			}
		}

		return &OpGate{
			target:   args[0],
			driftPct: args[1],
			vars:     app.vars,
		}, nil
	})

}
