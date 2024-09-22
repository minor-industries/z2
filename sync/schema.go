package sync

//go:generate msgp

type NamedSeries struct {
	Name       string
	Timestamps []int64
	Values     []float64
}

type Marker struct {
	ID        string
	Type      string
	Ref       string
	Timestamp int64
}
