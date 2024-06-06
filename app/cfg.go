package app

type Config struct {
	LongTermAverage     string
	LongTermAverageName string

	ShortTermAverage     string
	ShortTermAverageName string

	Target          string
	MaxDriftPct     string
	AllowedErrorPct string

	DriftMin string
	DriftMax string
}

var cfg = Config{
	LongTermAverage: seriesBuilder(
		"bike_instant_speed",
		"mygate bike_target_speed bike_max_drift_pct",
		"gt 20",
		"avg 10m triangle",
	),
	LongTermAverageName: "bike_avg_speed_long",

	ShortTermAverage: seriesBuilder(
		"bike_instant_speed",
		"mygate bike_target_speed bike_max_drift_pct",
		"gt 20",
		"avg 30s triangle",
	),
	ShortTermAverageName: "bike_avg_speed_short",

	DriftMin: "bike_instant_speed_min",
	DriftMax: "bike_instant_speed_max",

	Target:          "bike_target_speed",
	MaxDriftPct:     "bike_max_drift_pct",
	AllowedErrorPct: "bike_allowed_error_pct",
}
