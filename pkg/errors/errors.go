package errors

import (
	"fmt"
)

// Error types for GoTsunami
var (
	ErrInvalidConfig        = New("invalid configuration")
	ErrScenarioNotFound     = New("scenario file not found")
	ErrInvalidScenario      = New("invalid scenario configuration")
	ErrProtocolNotSupported = New("protocol not supported")
	ErrValidationFailed     = New("validation failed")
	ErrTimeoutExceeded      = New("timeout exceeded")
	ErrConnectionFailed     = New("connection failed")
	ErrInvalidResponse      = New("invalid response")
)

// GoTsunamiError represents a GoTsunami-specific error
type GoTsunamiError struct {
	Type    string
	Message string
	Cause   error
}

// Error implements the error interface
func (e *GoTsunamiError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *GoTsunamiError) Unwrap() error {
	return e.Cause
}

// New creates a new GoTsunami error
func New(message string) *GoTsunamiError {
	return &GoTsunamiError{
		Type:    "GoTsunamiError",
		Message: message,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string) *GoTsunamiError {
	return &GoTsunamiError{
		Type:    "GoTsunamiError",
		Message: message,
		Cause:   err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, format string, args ...interface{}) *GoTsunamiError {
	return &GoTsunamiError{
		Type:    "GoTsunamiError",
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}

// NewConfigError creates a configuration error
func NewConfigError(message string) *GoTsunamiError {
	return &GoTsunamiError{
		Type:    "ConfigError",
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *GoTsunamiError {
	return &GoTsunamiError{
		Type:    "ValidationError",
		Message: message,
	}
}

// NewProtocolError creates a protocol error
func NewProtocolError(message string) *GoTsunamiError {
	return &GoTsunamiError{
		Type:    "ProtocolError",
		Message: message,
	}
}
