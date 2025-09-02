package utils

import (
	"time"
)

// ParseDurationWithDefault parses a duration string with a default fallback
func ParseDurationWithDefault(durationStr string, defaultDuration time.Duration) time.Duration {
	if durationStr == "" {
		return defaultDuration
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return defaultDuration
	}

	return duration
}

// FormatDuration formats a duration in a human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return d.String()
	}

	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}

	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}

	if d < time.Minute {
		return d.Round(time.Millisecond).String()
	}

	return d.Round(time.Second).String()
}

// CalculatePercentile calculates a percentile from a slice of durations
func CalculatePercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Sort durations
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	// Simple bubble sort (for small datasets)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)-1) * percentile / 100)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// CalculateAverage calculates the average of a slice of durations
func CalculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}

// CalculateMinMax calculates the minimum and maximum durations
func CalculateMinMax(durations []time.Duration) (min, max time.Duration) {
	if len(durations) == 0 {
		return 0, 0
	}

	min = durations[0]
	max = durations[0]

	for _, d := range durations {
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	return min, max
}
