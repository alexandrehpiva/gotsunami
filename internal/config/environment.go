package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Environment manages environment variables and configuration
type Environment struct {
	variables map[string]string
}

// NewEnvironment creates a new environment instance
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]string),
	}
}

// LoadFromFile loads environment variables from a .env file
func (e *Environment) LoadFromFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("environment file not found: %s", filename)
	}

	if err := godotenv.Load(filename); err != nil {
		return fmt.Errorf("failed to load environment file: %w", err)
	}

	// Load all environment variables
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			e.variables[pair[0]] = pair[1]
		}
	}

	return nil
}

// Get retrieves an environment variable value
func (e *Environment) Get(key string) (string, bool) {
	// First check custom variables
	if value, exists := e.variables[key]; exists {
		return value, true
	}

	// Then check system environment
	if value := os.Getenv(key); value != "" {
		return value, true
	}

	return "", false
}

// Set sets a custom environment variable
func (e *Environment) Set(key, value string) {
	e.variables[key] = value
}

// ExpandVariables expands template variables in a string
func (e *Environment) ExpandVariables(template string) string {
	result := template

	// Replace {{env.VARIABLE}} patterns
	for key, value := range e.variables {
		pattern := fmt.Sprintf("{{env.%s}}", key)
		result = strings.ReplaceAll(result, pattern, value)
	}

	// Replace system environment variables
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			pattern := fmt.Sprintf("{{env.%s}}", pair[0])
			result = strings.ReplaceAll(result, pattern, pair[1])
		}
	}

	return result
}

// ExpandMap expands template variables in a map
func (e *Environment) ExpandMap(data map[string]string) map[string]string {
	result := make(map[string]string)

	for key, value := range data {
		expandedKey := e.ExpandVariables(key)
		expandedValue := e.ExpandVariables(value)
		result[expandedKey] = expandedValue
	}

	return result
}

// GetDefaultConfig returns default environment configuration
func GetDefaultConfig() map[string]string {
	return map[string]string{
		"DEFAULT_VUS":      "10",
		"DEFAULT_DURATION": "30s",
		"DEFAULT_TIMEOUT":  "30s",
		"LOG_LEVEL":        "info",
		"LOG_FORMAT":       "text",
		"REPORT_FORMAT":    "json",
		"USER_AGENT":       "GoTsunami/1.0",
		"KEEP_ALIVE":       "true",
		"CONNECTIONS":      "100",
		"WORKERS":          "0", // 0 = CPU cores
	}
}
