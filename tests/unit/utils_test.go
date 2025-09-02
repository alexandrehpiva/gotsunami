package unit

import (
	"testing"
	"time"

	"github.com/alexandredias/gotsunami/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestExpandTemplate(t *testing.T) {
	template := "Hello {{name}}, your token is {{token}}"
	variables := map[string]string{
		"name":  "John",
		"token": "abc123",
	}

	result := utils.ExpandTemplate(template, variables)
	expected := "Hello John, your token is abc123"
	assert.Equal(t, expected, result)
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name       string
		s          string
		substrings []string
		expected   bool
	}{
		{
			name:       "contains one",
			s:          "hello world",
			substrings: []string{"world", "universe"},
			expected:   true,
		},
		{
			name:       "contains none",
			s:          "hello world",
			substrings: []string{"universe", "galaxy"},
			expected:   false,
		},
		{
			name:       "empty substrings",
			s:          "hello world",
			substrings: []string{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ContainsAny(tt.s, tt.substrings)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsAll(t *testing.T) {
	tests := []struct {
		name       string
		s          string
		substrings []string
		expected   bool
	}{
		{
			name:       "contains all",
			s:          "hello world",
			substrings: []string{"hello", "world"},
			expected:   true,
		},
		{
			name:       "contains some",
			s:          "hello world",
			substrings: []string{"hello", "universe"},
			expected:   false,
		},
		{
			name:       "empty substrings",
			s:          "hello world",
			substrings: []string{},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ContainsAll(tt.s, tt.substrings)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			s:        "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "long string",
			s:        "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "exact length",
			s:        "hello",
			maxLen:   5,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.TruncateString(tt.s, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected bool
	}{
		{
			name:     "empty string",
			s:        "",
			expected: true,
		},
		{
			name:     "whitespace only",
			s:        "   ",
			expected: true,
		},
		{
			name:     "non-empty string",
			s:        "hello",
			expected: false,
		},
		{
			name:     "string with whitespace",
			s:        " hello ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsEmpty(tt.s)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinNonEmpty(t *testing.T) {
	result := utils.JoinNonEmpty(", ", "hello", "", "world", "   ", "test")
	expected := "hello, world, test"
	assert.Equal(t, expected, result)
}

func TestParseDurationWithDefault(t *testing.T) {
	tests := []struct {
		name            string
		durationStr     string
		defaultDuration time.Duration
		expected        time.Duration
	}{
		{
			name:            "valid duration",
			durationStr:     "5s",
			defaultDuration: 10 * time.Second,
			expected:        5 * time.Second,
		},
		{
			name:            "empty string",
			durationStr:     "",
			defaultDuration: 10 * time.Second,
			expected:        10 * time.Second,
		},
		{
			name:            "invalid duration",
			durationStr:     "invalid",
			defaultDuration: 10 * time.Second,
			expected:        10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ParseDurationWithDefault(tt.durationStr, tt.defaultDuration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "microseconds",
			duration: 500 * time.Microsecond,
			expected: "500µs",
		},
		{
			name:     "milliseconds",
			duration: 500 * time.Millisecond,
			expected: "500ms",
		},
		{
			name:     "seconds",
			duration: 5 * time.Second,
			expected: "5s",
		},
		{
			name:     "minutes",
			duration: 2 * time.Minute,
			expected: "2m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculatePercentile(t *testing.T) {
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
		400 * time.Millisecond,
		500 * time.Millisecond,
	}

	tests := []struct {
		name       string
		percentile float64
		expected   time.Duration
	}{
		{
			name:       "50th percentile",
			percentile: 50,
			expected:   300 * time.Millisecond,
		},
		{
			name:       "90th percentile",
			percentile: 90,
			expected:   400 * time.Millisecond,
		},
		{
			name:       "0th percentile",
			percentile: 0,
			expected:   100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CalculatePercentile(durations, tt.percentile)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateAverage(t *testing.T) {
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
	}

	result := utils.CalculateAverage(durations)
	expected := 200 * time.Millisecond
	assert.Equal(t, expected, result)

	// Test empty slice
	empty := []time.Duration{}
	result = utils.CalculateAverage(empty)
	assert.Equal(t, time.Duration(0), result)
}

func TestCalculateMinMax(t *testing.T) {
	durations := []time.Duration{
		300 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
		200 * time.Millisecond,
	}

	min, max := utils.CalculateMinMax(durations)
	assert.Equal(t, 100*time.Millisecond, min)
	assert.Equal(t, 500*time.Millisecond, max)

	// Test empty slice
	empty := []time.Duration{}
	min, max = utils.CalculateMinMax(empty)
	assert.Equal(t, time.Duration(0), min)
	assert.Equal(t, time.Duration(0), max)
}
