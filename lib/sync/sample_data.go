package sync

var namedSampleData = []NamedSeries{
	{
		Name:       "Series A",
		Timestamps: []int64{1627689945, 1627689955, 1627689965},
		Values:     []float64{1.23, 2.34, 3.45},
	},
	{
		Name:       "Series B",
		Timestamps: []int64{1627689975, 1627689985, 1627689995},
		Values:     []float64{4.56, 5.67, 6.78},
	},
	{
		Name:       "Series A",
		Timestamps: []int64{1627689965, 1627690005, 1627690015}, // Overlap with 1627689965, new unique timestamps
		Values:     []float64{7.89, 8.90, 9.01},
	},
}
