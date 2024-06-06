package app

import (
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/messages"
	"github.com/minor-industries/rtgraph/subscription"
	"github.com/minor-industries/z2/variables"
	"math"
	"time"
)

type StateChange struct {
	From string
	To   string
}

type App struct {
	Graph        *rtgraph.Graph
	vars         *variables.Cache
	stateChanges chan StateChange
}

func NewApp(graph *rtgraph.Graph, vars *variables.Cache) *App {
	app := &App{
		Graph:        graph,
		vars:         vars,
		stateChanges: make(chan StateChange),
	}

	app.setupGraphFunctions()

	return app
}

func (app *App) ComputeBounds() {
	// error vs drift:
	//  - error: the difference between the long-term average and the target
	//  - drift: the bounds for the short term average to move about without signaling to the user
	// drift will typically be configured to be a bit larger than error

	const (
		errorSteps = 5
	)

	now := time.Now()
	msgCh := make(chan *messages.Data)
	go app.Graph.Subscribe(&subscription.Request{
		Series:      []string{cfg.LongTermAverage},
		WindowSize:  uint64((10 * time.Minute).Milliseconds()),
		LastPointMs: 0,
		MaxGapMs:    uint64((5 * time.Second).Milliseconds()),
	}, now, msgCh)

	for m := range msgCh {
		for _, s := range m.Series {
			ts := time.UnixMilli(s.Timestamps[0])
			avgLong := s.Values[0]

			target, _ := app.vars.GetOne(cfg.Target)
			maxDriftPct, _ := app.vars.GetOne(cfg.MaxDriftPct)
			allowedErrorPct, _ := app.vars.GetOne(cfg.AllowedErrorPct)

			maxDrift := target * maxDriftPct / 100.0
			allowedError := target * allowedErrorPct / 100.0

			outStepSize := maxDrift / errorSteps
			stepSize := allowedError / errorSteps

			e := avgLong - target
			steps := int(math.Round(math.Abs(e) / stepSize))
			steps = min(steps, errorSteps)
			steps = -sign(e) * steps

			outAdjust := float64(steps) * outStepSize

			minTarget := target + outAdjust - maxDrift
			maxTarget := target + maxDrift + outAdjust

			if err := app.Graph.CreateValue(
				cfg.LongTermAverageName,
				ts,
				avgLong,
			); err != nil {
				panic(err)
			}

			if err := app.Graph.CreateValue(
				cfg.DriftMin,
				ts,
				minTarget,
			); err != nil {
				panic(err)
			}

			if err := app.Graph.CreateValue(
				cfg.DriftMax,
				ts,
				maxTarget,
			); err != nil {
				panic(err)
			}
		}
	}
}

func (app *App) ComputePace() {
	now := time.Now()
	msgCh := make(chan *messages.Data)
	go app.Graph.Subscribe(&subscription.Request{
		Series: []string{
			cfg.ShortTermAverage,
			cfg.DriftMin,
			cfg.DriftMax,
		},
		WindowSize:  uint64((30 * time.Second).Milliseconds()),
		LastPointMs: 0,
		MaxGapMs:    uint64((5 * time.Second).Milliseconds()),
	}, now, msgCh)

	state := "undefined"
	var avgShort, minTarget, maxTarget float64

	for m := range msgCh {
		var ts time.Time
		for _, s := range m.Series {
			switch s.Pos {
			case 0:
				avgShort = s.Values[0]
				ts = time.UnixMilli(s.Timestamps[0])
				if err := app.Graph.CreateValue(
					cfg.ShortTermAverageName,
					ts,
					avgShort,
				); err != nil {
					panic(err)
				}
			case 1:
				minTarget = s.Values[0]
				ts = time.UnixMilli(s.Timestamps[0])
			case 2:
				maxTarget = s.Values[0]
				ts = time.UnixMilli(s.Timestamps[0])
			}
		}

		// perhaps we could use a better way to get these values than subscribing and listening for them
		if avgShort == 0 || minTarget == 0 || maxTarget == 0 {
			continue
		}

		var tooFast, tooSlow, fairway float64

		var newState string

		switch {
		case avgShort > maxTarget:
			newState = "too_fast"
			tooFast = 1.0
		case avgShort < minTarget:
			newState = "too_slow"
			tooSlow = 1.0
		default:
			newState = "fairway"
			fairway = 1.0
		}

		if state != newState {
			app.stateChanges <- StateChange{
				From: state,
				To:   newState,
			}
		}

		state = newState

		_ = app.Graph.CreateValue("too_fast", ts, tooFast)
		_ = app.Graph.CreateValue("too_slow", ts, tooSlow)
		_ = app.Graph.CreateValue("fairway", ts, fairway)
	}
}

func (app *App) Run() {
	go app.PlaySounds()
	go app.ComputeBounds()
	go app.ComputePace()
}
