package app

import (
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/messages"
	"github.com/minor-industries/rtgraph/subscription"
	"github.com/minor-industries/z2/lib/variables"
	"math"
	"time"
)

type StateChange struct {
	From string
	To   string
}

type PlaySound struct {
	Sound string
}

func (p *PlaySound) Name() string {
	return "PlaySound"
}

type App struct {
	Graph        *rtgraph.Graph
	vars         *variables.Cache
	stateChanges chan StateChange
	cfg          Config
	broker       *broker.Broker
	audioPlayer  string
}

func NewApp(
	graph *rtgraph.Graph,
	vars *variables.Cache,
	br *broker.Broker,
	kind string,
	audioPlayer string,
) *App {
	cfg, ok := Configs[kind]
	if !ok {
		panic("unknown kind")
	}

	switch audioPlayer {
	case "browser", "backend":
	default:
		panic("unknown audio player")
	}

	app := &App{
		Graph:        graph,
		vars:         vars,
		stateChanges: make(chan StateChange),
		cfg:          cfg,
		broker:       br,
		audioPlayer:  audioPlayer,
	}

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
		Series:      []string{app.cfg.LongTermAverage},
		WindowSize:  uint64((10 * time.Minute).Milliseconds()),
		LastPointMs: 0,
	}, now, msgCh)

	for m := range msgCh {
		for _, s := range m.Series {
			ts := time.UnixMilli(s.Timestamps[0])
			avgLong := s.Values[0]

			target, _ := app.vars.GetOne(app.cfg.Target)
			maxDriftPct, _ := app.vars.GetOne(app.cfg.MaxDriftPct)
			allowedErrorPct, _ := app.vars.GetOne(app.cfg.AllowedErrorPct)

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
				app.cfg.LongTermAverageName,
				ts,
				avgLong,
			); err != nil {
				panic(err)
			}

			if err := app.Graph.CreateValue(
				app.cfg.Target,
				ts,
				target,
			); err != nil {
				panic(err)
			}

			if err := app.Graph.CreateValue(
				app.cfg.DriftMin,
				ts,
				minTarget,
			); err != nil {
				panic(err)
			}

			if err := app.Graph.CreateValue(
				app.cfg.DriftMax,
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
			app.cfg.ShortTermAverage,
			app.cfg.DriftMin,
			app.cfg.DriftMax,
		},
		WindowSize:  uint64((30 * time.Second).Milliseconds()),
		LastPointMs: 0,
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
					app.cfg.ShortTermAverageName,
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

		var newState string

		switch {
		case avgShort > maxTarget:
			newState = "too_fast"
		case avgShort < minTarget:
			newState = "too_slow"
		default:
			newState = "fairway"
		}

		if state != newState {
			app.stateChanges <- StateChange{
				From: state,
				To:   newState,
			}
		}

		state = newState
	}
}

func (app *App) DetectWorkouts() {
	now := time.Now()
	msgCh := make(chan *messages.Data)
	go app.Graph.Subscribe(&subscription.Request{
		Series: []string{
			seriesBuilder(app.cfg.PaceMetric, "detect-workout bike_target_speed bike_max_drift_pct"),
		},
		WindowSize:  uint64((90 * time.Second).Milliseconds()),
		LastPointMs: 0,
	}, now, msgCh)

	for range msgCh {
	}
}

func (app *App) Run() {
	go app.PlaySounds()
	go app.ComputeBounds()
	go app.ComputePace()
	go app.DetectWorkouts()
}
