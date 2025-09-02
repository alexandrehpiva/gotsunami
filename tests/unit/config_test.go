package unit

import (
	"testing"
	"time"

	"github.com/alexandredias/gotsunami/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestScenarioValidation(t *testing.T) {
	tests := []struct {
		name      string
		scenario  *config.Scenario
		wantError bool
	}{
		{
			name: "valid scenario",
			scenario: &config.Scenario{
				Name:    "test",
				Method:  "GET",
				URL:     "/test",
				BaseURL: "https://example.com",
			},
			wantError: false,
		},
		{
			name: "missing name",
			scenario: &config.Scenario{
				Method:  "GET",
				URL:     "/test",
				BaseURL: "https://example.com",
			},
			wantError: true,
		},
		{
			name: "invalid method",
			scenario: &config.Scenario{
				Name:    "test",
				Method:  "INVALID",
				URL:     "/test",
				BaseURL: "https://example.com",
			},
			wantError: true,
		},
		{
			name: "invalid timeout",
			scenario: &config.Scenario{
				Name:    "test",
				Method:  "GET",
				URL:     "/test",
				BaseURL: "https://example.com",
				Timeout: "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scenario.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRetryConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		retry     *config.RetryConfig
		wantError bool
	}{
		{
			name: "valid retry config",
			retry: &config.RetryConfig{
				Attempts: 3,
				Backoff:  "exponential",
				MaxDelay: "5s",
			},
			wantError: false,
		},
		{
			name: "negative attempts",
			retry: &config.RetryConfig{
				Attempts: -1,
				Backoff:  "exponential",
				MaxDelay: "5s",
			},
			wantError: true,
		},
		{
			name: "too many attempts",
			retry: &config.RetryConfig{
				Attempts: 15,
				Backoff:  "exponential",
				MaxDelay: "5s",
			},
			wantError: true,
		},
		{
			name: "invalid backoff",
			retry: &config.RetryConfig{
				Attempts: 3,
				Backoff:  "invalid",
				MaxDelay: "5s",
			},
			wantError: true,
		},
		{
			name: "invalid max delay",
			retry: &config.RetryConfig{
				Attempts: 3,
				Backoff:  "exponential",
				MaxDelay: "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.retry.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationConfigValidation(t *testing.T) {
	tests := []struct {
		name       string
		validation *config.ValidationConfig
		wantError  bool
	}{
		{
			name: "valid validation config",
			validation: &config.ValidationConfig{
				StatusCodes:     []int{200, 201},
				ResponseTimeMax: "2s",
				BodyContains:    []string{"success"},
				BodyNotContains: []string{"error"},
			},
			wantError: false,
		},
		{
			name: "invalid status code",
			validation: &config.ValidationConfig{
				StatusCodes: []int{999},
			},
			wantError: true,
		},
		{
			name: "invalid response time",
			validation: &config.ValidationConfig{
				ResponseTimeMax: "invalid",
			},
			wantError: true,
		},
		{
			name: "invalid response size",
			validation: &config.ValidationConfig{
				MinResponseSize: 100,
				MaxResponseSize: 50,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validation.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScenarioGetTimeout(t *testing.T) {
	scenario := &config.Scenario{
		Timeout: "5s",
	}

	timeout := scenario.GetTimeout()
	assert.Equal(t, 5*time.Second, timeout)

	// Test default timeout
	scenario.Timeout = ""
	timeout = scenario.GetTimeout()
	assert.Equal(t, 30*time.Second, timeout)
}

func TestScenarioGetRetryConfig(t *testing.T) {
	scenario := &config.Scenario{}

	retry := scenario.GetRetryConfig()
	assert.Equal(t, 3, retry.Attempts)
	assert.Equal(t, "exponential", retry.Backoff)
	assert.Equal(t, "5s", retry.MaxDelay)
}

func TestScenarioGetValidationConfig(t *testing.T) {
	scenario := &config.Scenario{}

	validation := scenario.GetValidationConfig()
	assert.Equal(t, []int{200}, validation.StatusCodes)
}
