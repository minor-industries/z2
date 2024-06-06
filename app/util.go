package app

import "strings"

func sign(x float64) int {
	if x < 0 {
		return -1
	}
	return 1
}

func seriesBuilder(parts ...string) string {
	return strings.Join(parts, "|")
}
