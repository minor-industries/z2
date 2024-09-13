package time_series

import (
	"fmt"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/computed_series"
	"github.com/minor-industries/z2/variables"
	"github.com/pkg/errors"
	"time"
)

func SetupGraphFunctions(graph *rtgraph.Graph, vars *variables.Cache) {
	graph.Parser.AddFunction("mygate", func(start time.Time, args []string) (computed_series.Operator, error) {
		if len(args) != 2 {
			return nil, errors.New("mygate function requires 2 arguments")
		}

		got := vars.Get(args)
		for _, v := range got {
			if !v.Present {
				return nil, fmt.Errorf("variable %s not found", v.Name)
			}
		}

		return &OpGate{
			target:   args[0],
			driftPct: args[1],
			vars:     vars,
		}, nil
	})

	graph.Parser.AddFunction("detect-workout", func(start time.Time, args []string) (computed_series.Operator, error) {
		if len(args) != 2 {
			return nil, errors.New("detect-workout requires 2 arguments")
		}

		got := vars.Get(args)
		for _, v := range got {
			if !v.Present {
				return nil, fmt.Errorf("variable %s not found", v.Name)
			}
		}

		// TODO: perhaps start has to be part of Fcn, not "NewComputedSeries"
		fcn := computed_series.NewComputedSeries(&FcnDetectWorkout{
			vars:     vars,
			target:   args[0],
			driftPct: args[1],
		}, 90*time.Second, time.Time{})

		return fcn, nil
	})

	graph.Parser.AddFunction("time-bin", func(start time.Time, args []string) (computed_series.Operator, error) {
		if len(args) != 0 {
			return nil, errors.New("time-bin function requires no arguments")
		}

		return &OpTimeBin{}, nil
	})
}
