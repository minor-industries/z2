package sync

//go:generate msgp

type NamedSeries struct {
	Name       string
	Timestamps []int64
	Values     []float64
}
