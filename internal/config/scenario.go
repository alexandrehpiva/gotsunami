package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Scenario represents a load test scenario configuration
type Scenario struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	BaseURL     string                 `json:"base_url"`
	Headers     map[string]string      `json:"headers,omitempty"`
	QueryParams map[string]interface{} `json:"query_params,omitempty"`
	Body        interface{}            `json:"body,omitempty"`
	Timeout     string                 `json:"timeout,omitempty"`
	Retry       *RetryConfig           `json:"retry,omitempty"`
	Validation  *ValidationConfig      `json:"validation,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	Variables   map[string]string      `json:"variables,omitempty"`
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	Attempts int    `json:"attempts"`
	Backoff  string `json:"backoff"`
	MaxDelay string `json:"max_delay"`
}

// ValidationConfig defines response validation rules
type ValidationConfig struct {
	StatusCodes     []int             `json:"status_codes,omitempty"`
	ResponseTimeMax string            `json:"response_time_max,omitempty"`
	BodyContains    []string          `json:"body_contains,omitempty"`
	BodyNotContains []string          `json:"body_not_contains,omitempty"`
	BodyRegex       string            `json:"body_regex,omitempty"`
	BodyJSONPath    string            `json:"body_json_path,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	MinResponseSize int               `json:"min_response_size,omitempty"`
	MaxResponseSize int               `json:"max_response_size,omitempty"`
}

// LoadTestConfig represents the complete load test configuration
type LoadTestConfig struct {
	Scenario     *Scenario     `json:"scenario"`
	VirtualUsers int           `json:"virtual_users"`
	Duration     time.Duration `json:"duration"`
	RampUp       time.Duration `json:"ramp_up"`
	RampDown     time.Duration `json:"ramp_down"`
	Delay        time.Duration `json:"delay"`
	MaxRequests  int           `json:"max_requests"`
	Timeout      time.Duration `json:"timeout"`
	Pattern      string        `json:"pattern"`

	// Output configuration
	Live         bool   `json:"live"`
	ReportFormat string `json:"report_format"`
	Outfile      string `json:"outfile"`
	Stdout       bool   `json:"stdout"`

	// Validation overrides
	ExpectStatus       []int         `json:"expect_status,omitempty"`
	ExpectBody         string        `json:"expect_body,omitempty"`
	ExpectBodyNot      string        `json:"expect_body_not,omitempty"`
	ExpectResponseTime time.Duration `json:"expect_response_time,omitempty"`

	// Advanced configuration
	Workers       int    `json:"workers"`
	Connections   int    `json:"connections"`
	KeepAlive     bool   `json:"keep_alive"`
	TLSSkipVerify bool   `json:"tls_skip_verify"`
	Proxy         string `json:"proxy,omitempty"`
	UserAgent     string `json:"user_agent,omitempty"`
}

// LoadScenarioFromFile loads a scenario configuration from a JSON file
func LoadScenarioFromFile(filename string) (*Scenario, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	var scenario Scenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse scenario JSON: %w", err)
	}

	if err := scenario.Validate(); err != nil {
		return nil, fmt.Errorf("scenario validation failed: %w", err)
	}

	return &scenario, nil
}

// Validate validates the scenario configuration
func (s *Scenario) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("scenario name is required")
	}

	if s.Method == "" {
		return fmt.Errorf("scenario method is required")
	}

	if s.URL == "" {
		return fmt.Errorf("scenario URL is required")
	}

	if s.BaseURL == "" {
		return fmt.Errorf("scenario base_url is required")
	}

	// Validate method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "HEAD": true, "OPTIONS": true,
	}
	if !validMethods[s.Method] {
		return fmt.Errorf("invalid HTTP method: %s", s.Method)
	}

	// Validate timeout if provided
	if s.Timeout != "" {
		if _, err := time.ParseDuration(s.Timeout); err != nil {
			return fmt.Errorf("invalid timeout format: %s", s.Timeout)
		}
	}

	// Validate retry config if provided
	if s.Retry != nil {
		if err := s.Retry.Validate(); err != nil {
			return fmt.Errorf("retry config validation failed: %w", err)
		}
	}

	// Validate validation config if provided
	if s.Validation != nil {
		if err := s.Validation.Validate(); err != nil {
			return fmt.Errorf("validation config validation failed: %w", err)
		}
	}

	return nil
}

// Validate validates the retry configuration
func (r *RetryConfig) Validate() error {
	if r.Attempts < 0 {
		return fmt.Errorf("retry attempts must be non-negative")
	}

	if r.Attempts > 10 {
		return fmt.Errorf("retry attempts cannot exceed 10")
	}

	validBackoffs := map[string]bool{
		"linear": true, "exponential": true, "fixed": true,
	}
	if r.Backoff != "" && !validBackoffs[r.Backoff] {
		return fmt.Errorf("invalid backoff strategy: %s", r.Backoff)
	}

	if r.MaxDelay != "" {
		if _, err := time.ParseDuration(r.MaxDelay); err != nil {
			return fmt.Errorf("invalid max_delay format: %s", r.MaxDelay)
		}
	}

	return nil
}

// Validate validates the validation configuration
func (v *ValidationConfig) Validate() error {
	if len(v.StatusCodes) > 0 {
		for _, code := range v.StatusCodes {
			if code < 100 || code > 599 {
				return fmt.Errorf("invalid status code: %d", code)
			}
		}
	}

	if v.ResponseTimeMax != "" {
		if _, err := time.ParseDuration(v.ResponseTimeMax); err != nil {
			return fmt.Errorf("invalid response_time_max format: %s", v.ResponseTimeMax)
		}
	}

	if v.MinResponseSize < 0 {
		return fmt.Errorf("min_response_size must be non-negative")
	}

	if v.MaxResponseSize > 0 && v.MinResponseSize > v.MaxResponseSize {
		return fmt.Errorf("min_response_size cannot be greater than max_response_size")
	}

	return nil
}

// GetTimeout returns the timeout as a time.Duration
func (s *Scenario) GetTimeout() time.Duration {
	if s.Timeout == "" {
		return 30 * time.Second
	}

	duration, err := time.ParseDuration(s.Timeout)
	if err != nil {
		return 30 * time.Second
	}

	return duration
}

// GetRetryConfig returns the retry configuration with defaults
func (s *Scenario) GetRetryConfig() *RetryConfig {
	if s.Retry == nil {
		return &RetryConfig{
			Attempts: 3,
			Backoff:  "exponential",
			MaxDelay: "5s",
		}
	}
	return s.Retry
}

// GetValidationConfig returns the validation configuration with defaults
func (s *Scenario) GetValidationConfig() *ValidationConfig {
	if s.Validation == nil {
		return &ValidationConfig{
			StatusCodes: []int{200},
		}
	}
	return s.Validation
}
