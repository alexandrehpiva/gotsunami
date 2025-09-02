package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alexandredias/gotsunami/internal/config"
	"github.com/alexandredias/gotsunami/internal/protocols"
	"github.com/tidwall/gjson"
)

// ResponseValidator validates HTTP responses against configured rules
type ResponseValidator struct {
	config *config.ValidationConfig
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Passed    bool   `json:"passed"`
	ErrorType string `json:"error_type,omitempty"`
	Message   string `json:"message,omitempty"`
}

// NewResponseValidator creates a new response validator
func NewResponseValidator(config *config.ValidationConfig) *ResponseValidator {
	return &ResponseValidator{
		config: config,
	}
}

// Validate validates a response against all configured rules
func (v *ResponseValidator) Validate(resp *protocols.Response) *ValidationResult {
	// Check for request errors first
	if resp.Error != nil {
		return &ValidationResult{
			Passed:    false,
			ErrorType: "request_error",
			Message:   resp.Error.Error(),
		}
	}

	// Validate status code
	if result := v.validateStatusCode(resp.StatusCode); !result.Passed {
		return result
	}

	// Validate response time
	if result := v.validateResponseTime(resp.ResponseTime); !result.Passed {
		return result
	}

	// Validate response size
	if result := v.validateResponseSize(resp.ContentLength); !result.Passed {
		return result
	}

	// Validate body content
	if result := v.validateBody(resp.Body); !result.Passed {
		return result
	}

	// Validate headers
	if result := v.validateHeaders(resp.Headers); !result.Passed {
		return result
	}

	return &ValidationResult{
		Passed: true,
	}
}

// validateStatusCode validates the HTTP status code
func (v *ResponseValidator) validateStatusCode(statusCode int) *ValidationResult {
	if len(v.config.StatusCodes) == 0 {
		return &ValidationResult{Passed: true}
	}

	for _, expectedCode := range v.config.StatusCodes {
		if statusCode == expectedCode {
			return &ValidationResult{Passed: true}
		}
	}

	return &ValidationResult{
		Passed:    false,
		ErrorType: "status_code",
		Message:   fmt.Sprintf("expected status codes %v, got %d", v.config.StatusCodes, statusCode),
	}
}

// validateResponseTime validates the response time
func (v *ResponseValidator) validateResponseTime(responseTime time.Duration) *ValidationResult {
	if v.config.ResponseTimeMax == "" {
		return &ValidationResult{Passed: true}
	}

	maxTime, err := time.ParseDuration(v.config.ResponseTimeMax)
	if err != nil {
		return &ValidationResult{
			Passed:    false,
			ErrorType: "config_error",
			Message:   fmt.Sprintf("invalid response_time_max format: %s", v.config.ResponseTimeMax),
		}
	}

	if responseTime > maxTime {
		return &ValidationResult{
			Passed:    false,
			ErrorType: "response_time",
			Message:   fmt.Sprintf("response time %v exceeds maximum %v", responseTime, maxTime),
		}
	}

	return &ValidationResult{Passed: true}
}

// validateResponseSize validates the response size
func (v *ResponseValidator) validateResponseSize(size int64) *ValidationResult {
	if v.config.MinResponseSize > 0 && size < int64(v.config.MinResponseSize) {
		return &ValidationResult{
			Passed:    false,
			ErrorType: "response_size",
			Message:   fmt.Sprintf("response size %d is below minimum %d", size, v.config.MinResponseSize),
		}
	}

	if v.config.MaxResponseSize > 0 && size > int64(v.config.MaxResponseSize) {
		return &ValidationResult{
			Passed:    false,
			ErrorType: "response_size",
			Message:   fmt.Sprintf("response size %d exceeds maximum %d", size, v.config.MaxResponseSize),
		}
	}

	return &ValidationResult{Passed: true}
}

// validateBody validates the response body content
func (v *ResponseValidator) validateBody(body []byte) *ValidationResult {
	bodyStr := string(body)

	// Check body contains required strings
	for _, required := range v.config.BodyContains {
		if !strings.Contains(bodyStr, required) {
			return &ValidationResult{
				Passed:    false,
				ErrorType: "body_content",
				Message:   fmt.Sprintf("response body does not contain required string: %s", required),
			}
		}
	}

	// Check body does not contain forbidden strings
	for _, forbidden := range v.config.BodyNotContains {
		if strings.Contains(bodyStr, forbidden) {
			return &ValidationResult{
				Passed:    false,
				ErrorType: "body_content",
				Message:   fmt.Sprintf("response body contains forbidden string: %s", forbidden),
			}
		}
	}

	// Check body regex pattern
	if v.config.BodyRegex != "" {
		matched, err := regexp.MatchString(v.config.BodyRegex, bodyStr)
		if err != nil {
			return &ValidationResult{
				Passed:    false,
				ErrorType: "config_error",
				Message:   fmt.Sprintf("invalid body regex pattern: %s", v.config.BodyRegex),
			}
		}
		if !matched {
			return &ValidationResult{
				Passed:    false,
				ErrorType: "body_regex",
				Message:   fmt.Sprintf("response body does not match regex pattern: %s", v.config.BodyRegex),
			}
		}
	}

	// Check JSON path
	if v.config.BodyJSONPath != "" {
		if !v.validateJSONPath(body, v.config.BodyJSONPath) {
			return &ValidationResult{
				Passed:    false,
				ErrorType: "body_json_path",
				Message:   fmt.Sprintf("JSON path not found or invalid: %s", v.config.BodyJSONPath),
			}
		}
	}

	return &ValidationResult{Passed: true}
}

// validateJSONPath validates a JSON path in the response body
func (v *ResponseValidator) validateJSONPath(body []byte, jsonPath string) bool {
	if len(body) == 0 {
		return false
	}

	// Use gjson to parse JSON path
	result := gjson.GetBytes(body, jsonPath)
	return result.Exists()
}

// validateHeaders validates response headers
func (v *ResponseValidator) validateHeaders(headers map[string]string) *ValidationResult {
	if len(v.config.Headers) == 0 {
		return &ValidationResult{Passed: true}
	}

	for expectedHeader, expectedValue := range v.config.Headers {
		actualValue, exists := headers[expectedHeader]
		if !exists {
			return &ValidationResult{
				Passed:    false,
				ErrorType: "header_missing",
				Message:   fmt.Sprintf("required header missing: %s", expectedHeader),
			}
		}

		if actualValue != expectedValue {
			return &ValidationResult{
				Passed:    false,
				ErrorType: "header_value",
				Message:   fmt.Sprintf("header %s has unexpected value: expected %s, got %s", expectedHeader, expectedValue, actualValue),
			}
		}
	}

	return &ValidationResult{Passed: true}
}

// ValidateWithOverrides validates a response with CLI flag overrides
func (v *ResponseValidator) ValidateWithOverrides(resp *protocols.Response, overrides *ValidationOverrides) *ValidationResult {
	// Create temporary config with overrides
	tempConfig := *v.config

	if len(overrides.ExpectStatus) > 0 {
		tempConfig.StatusCodes = overrides.ExpectStatus
	}

	if overrides.ExpectResponseTime > 0 {
		tempConfig.ResponseTimeMax = overrides.ExpectResponseTime.String()
	}

	if overrides.ExpectBody != "" {
		tempConfig.BodyContains = []string{overrides.ExpectBody}
	}

	if overrides.ExpectBodyNot != "" {
		tempConfig.BodyNotContains = []string{overrides.ExpectBodyNot}
	}

	// Create temporary validator
	tempValidator := &ResponseValidator{config: &tempConfig}
	return tempValidator.Validate(resp)
}

// ValidationOverrides represents CLI flag overrides for validation
type ValidationOverrides struct {
	ExpectStatus       []int
	ExpectResponseTime time.Duration
	ExpectBody         string
	ExpectBodyNot      string
}
