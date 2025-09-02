package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/alexandredias/gotsunami/internal/config"
	"github.com/alexandredias/gotsunami/internal/metrics"
)

// JSONReporter generates JSON reports
type JSONReporter struct {
	config *config.LoadTestConfig
}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter(config *config.LoadTestConfig) *JSONReporter {
	return &JSONReporter{
		config: config,
	}
}

// GenerateReport generates a JSON report from metrics
func (r *JSONReporter) GenerateReport(summary *metrics.Summary, scenario *config.Scenario) (*Report, error) {
	report := &Report{
		Metadata: ReportMetadata{
			Tool:      "GoTsunami",
			Version:   "1.0.0",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Duration:  r.config.Duration.String(),
			Scenario:  scenario.Name,
		},
		Configuration: ReportConfiguration{
			VirtualUsers: r.config.VirtualUsers,
			Duration:     r.config.Duration.String(),
			RampUp:       r.config.RampUp.String(),
			RampDown:     r.config.RampDown.String(),
			Delay:        r.config.Delay.String(),
			Pattern:      r.config.Pattern,
		},
		Summary: ReportSummary{
			TotalRequests:      summary.TotalRequests,
			SuccessfulRequests: summary.SuccessfulRequests,
			FailedRequests:     summary.FailedRequests,
			SuccessRate:        summary.SuccessRate,
			TotalDuration:      r.config.Duration.String(),
		},
		Latency:           r.formatLatency(summary.Latency),
		Throughput:        r.formatThroughput(summary),
		Errors:            r.formatErrors(summary.Errors),
		StatusCodes:       r.formatStatusCodes(summary.StatusCodes),
		ValidationResults: r.formatValidationResults(summary.ValidationResults),
	}

	return report, nil
}

// WriteReport writes the report to a file or stdout
func (r *JSONReporter) WriteReport(report *Report, outfile string) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	if outfile != "" {
		err = os.WriteFile(outfile, jsonData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write report to file: %w", err)
		}
		fmt.Printf("Report written to: %s\n", outfile)
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

// formatLatency formats latency statistics
func (r *JSONReporter) formatLatency(latency *metrics.LatencyStats) ReportLatency {
	if latency == nil {
		return ReportLatency{}
	}

	return ReportLatency{
		Mean:   latency.Mean.String(),
		Median: latency.Median.String(),
		P90:    latency.P90.String(),
		P95:    latency.P95.String(),
		P99:    latency.P99.String(),
		P99_9:  latency.P99_9.String(),
		Min:    latency.Min.String(),
		Max:    latency.Max.String(),
	}
}

// formatThroughput formats throughput statistics
func (r *JSONReporter) formatThroughput(summary *metrics.Summary) ReportThroughput {
	return ReportThroughput{
		RequestsPerSecond: summary.RequestsPerSecond,
		BytesPerSecond:    summary.BytesPerSecond,
	}
}

// formatErrors formats error statistics
func (r *JSONReporter) formatErrors(errors map[string]int64) []ReportError {
	var reportErrors []ReportError
	totalRequests := int64(0)

	// Calculate total for percentage calculation
	for _, count := range errors {
		totalRequests += count
	}

	for errorType, count := range errors {
		percentage := float64(0)
		if totalRequests > 0 {
			percentage = float64(count) / float64(totalRequests) * 100
		}

		reportErrors = append(reportErrors, ReportError{
			Type:       errorType,
			Count:      count,
			Percentage: percentage,
		})
	}

	return reportErrors
}

// formatStatusCodes formats status code distribution
func (r *JSONReporter) formatStatusCodes(statusCodes map[int]int64) map[string]int64 {
	result := make(map[string]int64)
	for code, count := range statusCodes {
		result[fmt.Sprintf("%d", code)] = count
	}
	return result
}

// formatValidationResults formats validation results
func (r *JSONReporter) formatValidationResults(results *metrics.ValidationResults) ReportValidationResults {
	if results == nil {
		return ReportValidationResults{}
	}

	statusCodeValidation := "passed"
	responseTimeValidation := "passed"
	bodyValidation := "passed"

	if results.FailedValidations > 0 {
		bodyValidation = "failed"
	}

	return ReportValidationResults{
		StatusCodeValidation:   statusCodeValidation,
		ResponseTimeValidation: responseTimeValidation,
		BodyValidation:         bodyValidation,
		FailedValidations:      results.FailedValidations,
	}
}

// Report represents the complete test report
type Report struct {
	Metadata          ReportMetadata          `json:"metadata"`
	Configuration     ReportConfiguration     `json:"configuration"`
	Summary           ReportSummary           `json:"summary"`
	Latency           ReportLatency           `json:"latency"`
	Throughput        ReportThroughput        `json:"throughput"`
	Errors            []ReportError           `json:"errors"`
	StatusCodes       map[string]int64        `json:"status_codes"`
	ValidationResults ReportValidationResults `json:"validation_results"`
}

// ReportMetadata contains report metadata
type ReportMetadata struct {
	Tool      string `json:"tool"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Duration  string `json:"duration"`
	Scenario  string `json:"scenario"`
}

// ReportConfiguration contains test configuration
type ReportConfiguration struct {
	VirtualUsers int    `json:"virtual_users"`
	Duration     string `json:"duration"`
	RampUp       string `json:"ramp_up"`
	RampDown     string `json:"ramp_down"`
	Delay        string `json:"delay"`
	Pattern      string `json:"pattern"`
}

// ReportSummary contains test summary
type ReportSummary struct {
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	SuccessRate        float64 `json:"success_rate"`
	TotalDuration      string  `json:"total_duration"`
}

// ReportLatency contains latency statistics
type ReportLatency struct {
	Mean   string `json:"mean"`
	Median string `json:"median"`
	P90    string `json:"p90"`
	P95    string `json:"p95"`
	P99    string `json:"p99"`
	P99_9  string `json:"p99.9"`
	Min    string `json:"min"`
	Max    string `json:"max"`
}

// ReportThroughput contains throughput statistics
type ReportThroughput struct {
	RequestsPerSecond float64 `json:"requests_per_second"`
	BytesPerSecond    float64 `json:"bytes_per_second"`
}

// ReportError contains error information
type ReportError struct {
	Type       string  `json:"type"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ReportValidationResults contains validation results
type ReportValidationResults struct {
	StatusCodeValidation   string `json:"status_code_validation"`
	ResponseTimeValidation string `json:"response_time_validation"`
	BodyValidation         string `json:"body_validation"`
	FailedValidations      int64  `json:"failed_validations"`
}
