package app

type Config struct {
	LongTermAverage     string
	LongTermAverageName string

	ShortTermAverage     string
	ShortTermAverageName string

	PaceMetric      string
	Target          string
	MaxDriftPct     string
	AllowedErrorPct string

	DriftMin string
	DriftMax string
}

var bikeConfig = Config{
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

	PaceMetric:      "bike_instant_speed",
	Target:          "bike_target_speed",
	MaxDriftPct:     "bike_max_drift_pct",
	AllowedErrorPct: "bike_allowed_error_pct",
}

var rowerConfig = Config{
	LongTermAverage: seriesBuilder(
		"rower_power",
		"mygate rower_target_power rower_max_drift_pct",
		"gt 60",
		"avg 10m triangle",
	),
	LongTermAverageName: "rower_avg_power_long",

	ShortTermAverage: seriesBuilder(
		"rower_power",
		"mygate rower_target_power rower_max_drift_pct",
		"gt 60",
		"avg 30s triangle",
	),
	ShortTermAverageName: "rower_avg_power_short",

	DriftMin: "rower_power_min",
	DriftMax: "rower_power_max",

	PaceMetric:      "rower_power",
	Target:          "rower_target_power",
	MaxDriftPct:     "rower_max_drift_pct",
	AllowedErrorPct: "rower_allowed_error_pct",
}

var configs = map[string]Config{
	"bike":  bikeConfig,
	"rower": rowerConfig,
}
