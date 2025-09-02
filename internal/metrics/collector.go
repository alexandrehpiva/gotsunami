package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/alexandredias/gotsunami/internal/protocols"
)

// Collector collects and aggregates metrics during load testing
type Collector struct {
	mu sync.RWMutex

	// Request metrics
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64
	totalBytes         int64

	// Latency metrics
	latencies    []time.Duration
	minLatency   time.Duration
	maxLatency   time.Duration
	totalLatency time.Duration

	// Status code distribution
	statusCodes map[int]int64

	// Error tracking
	errors map[string]int64

	// Time tracking
	startTime time.Time
	endTime   time.Time

	// Validation results
	validationResults *ValidationResults
}

// ValidationResults tracks validation outcomes
type ValidationResults struct {
	TotalValidations  int64
	PassedValidations int64
	FailedValidations int64
	ValidationErrors  map[string]int64
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		statusCodes: make(map[int]int64),
		errors:      make(map[string]int64),
		validationResults: &ValidationResults{
			ValidationErrors: make(map[string]int64),
		},
	}
}

// Start begins metrics collection
func (c *Collector) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startTime = time.Now()
}

// Stop ends metrics collection
func (c *Collector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.endTime = time.Now()
}

// RecordResponse records a response and its metrics
func (c *Collector) RecordResponse(resp *protocols.Response) {
	atomic.AddInt64(&c.totalRequests, 1)
	atomic.AddInt64(&c.totalBytes, resp.ContentLength)

	// Update latency metrics
	c.updateLatency(resp.ResponseTime)

	// Update status code distribution
	c.updateStatusCode(resp.StatusCode)

	// Update success/failure counts
	if resp.Error != nil || resp.StatusCode >= 400 {
		atomic.AddInt64(&c.failedRequests, 1)
		c.recordError(resp.Error)
	} else {
		atomic.AddInt64(&c.successfulRequests, 1)
	}
}

// updateLatency updates latency-related metrics
func (c *Collector) updateLatency(latency time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.latencies = append(c.latencies, latency)
	c.totalLatency += latency

	if c.minLatency == 0 || latency < c.minLatency {
		c.minLatency = latency
	}
	if latency > c.maxLatency {
		c.maxLatency = latency
	}
}

// updateStatusCode updates status code distribution
func (c *Collector) updateStatusCode(statusCode int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statusCodes[statusCode]++
}

// recordError records an error occurrence
func (c *Collector) recordError(err error) {
	if err == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.errors[err.Error()]++
}

// RecordValidation records a validation result
func (c *Collector) RecordValidation(passed bool, errorType string) {
	atomic.AddInt64(&c.validationResults.TotalValidations, 1)

	if passed {
		atomic.AddInt64(&c.validationResults.PassedValidations, 1)
	} else {
		atomic.AddInt64(&c.validationResults.FailedValidations, 1)
		if errorType != "" {
			c.mu.Lock()
			c.validationResults.ValidationErrors[errorType]++
			c.mu.Unlock()
		}
	}
}

// GetSummary returns a summary of collected metrics
func (c *Collector) GetSummary() *Summary {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := &Summary{
		TotalRequests:      atomic.LoadInt64(&c.totalRequests),
		SuccessfulRequests: atomic.LoadInt64(&c.successfulRequests),
		FailedRequests:     atomic.LoadInt64(&c.failedRequests),
		TotalBytes:         atomic.LoadInt64(&c.totalBytes),
		StatusCodes:        make(map[int]int64),
		Errors:             make(map[string]int64),
		ValidationResults:  c.validationResults,
	}

	// Copy status codes
	for code, count := range c.statusCodes {
		summary.StatusCodes[code] = count
	}

	// Copy errors
	for err, count := range c.errors {
		summary.Errors[err] = count
	}

	// Calculate latency statistics
	if len(c.latencies) > 0 {
		summary.Latency = c.calculateLatencyStats()
	}

	// Calculate success rate
	if summary.TotalRequests > 0 {
		summary.SuccessRate = float64(summary.SuccessfulRequests) / float64(summary.TotalRequests) * 100
	}

	// Calculate throughput
	if !c.startTime.IsZero() && !c.endTime.IsZero() {
		duration := c.endTime.Sub(c.startTime)
		if duration > 0 {
			summary.RequestsPerSecond = float64(summary.TotalRequests) / duration.Seconds()
			summary.BytesPerSecond = float64(summary.TotalBytes) / duration.Seconds()
		}
	}

	return summary
}

// calculateLatencyStats calculates latency statistics
func (c *Collector) calculateLatencyStats() *LatencyStats {
	if len(c.latencies) == 0 {
		return &LatencyStats{}
	}

	// Sort latencies for percentile calculation
	sortedLatencies := make([]time.Duration, len(c.latencies))
	copy(sortedLatencies, c.latencies)

	// Simple sort (in production, use a more efficient algorithm)
	for i := 0; i < len(sortedLatencies); i++ {
		for j := i + 1; j < len(sortedLatencies); j++ {
			if sortedLatencies[i] > sortedLatencies[j] {
				sortedLatencies[i], sortedLatencies[j] = sortedLatencies[j], sortedLatencies[i]
			}
		}
	}

	stats := &LatencyStats{
		Min:    c.minLatency,
		Max:    c.maxLatency,
		Mean:   c.totalLatency / time.Duration(len(c.latencies)),
		Median: c.calculatePercentile(sortedLatencies, 50),
		P90:    c.calculatePercentile(sortedLatencies, 90),
		P95:    c.calculatePercentile(sortedLatencies, 95),
		P99:    c.calculatePercentile(sortedLatencies, 99),
		P99_9:  c.calculatePercentile(sortedLatencies, 99.9),
	}

	return stats
}

// calculatePercentile calculates a percentile from sorted latencies
func (c *Collector) calculatePercentile(sortedLatencies []time.Duration, percentile float64) time.Duration {
	if len(sortedLatencies) == 0 {
		return 0
	}

	index := int(float64(len(sortedLatencies)-1) * percentile / 100)
	if index >= len(sortedLatencies) {
		index = len(sortedLatencies) - 1
	}

	return sortedLatencies[index]
}

// Summary represents aggregated metrics
type Summary struct {
	TotalRequests      int64              `json:"total_requests"`
	SuccessfulRequests int64              `json:"successful_requests"`
	FailedRequests     int64              `json:"failed_requests"`
	SuccessRate        float64            `json:"success_rate"`
	TotalBytes         int64              `json:"total_bytes"`
	RequestsPerSecond  float64            `json:"requests_per_second"`
	BytesPerSecond     float64            `json:"bytes_per_second"`
	Latency            *LatencyStats      `json:"latency"`
	StatusCodes        map[int]int64      `json:"status_codes"`
	Errors             map[string]int64   `json:"errors"`
	ValidationResults  *ValidationResults `json:"validation_results"`
}

// LatencyStats represents latency statistics
type LatencyStats struct {
	Min    time.Duration `json:"min"`
	Max    time.Duration `json:"max"`
	Mean   time.Duration `json:"mean"`
	Median time.Duration `json:"median"`
	P90    time.Duration `json:"p90"`
	P95    time.Duration `json:"p95"`
	P99    time.Duration `json:"p99"`
	P99_9  time.Duration `json:"p99_9"`
}
