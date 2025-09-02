package protocols

import (
	"context"
	"time"
)

// Request represents a protocol request
type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	Body        []byte
	Timeout     time.Duration
	QueryParams map[string]interface{}
}

// Response represents a protocol response
type Response struct {
	StatusCode    int
	Headers       map[string]string
	Body          []byte
	ResponseTime  time.Duration
	ContentLength int64
	Error         error
}

// Protocol defines the interface for different protocols
type Protocol interface {
	// Name returns the protocol name
	Name() string

	// Version returns the protocol version
	Version() string

	// Execute performs a request using this protocol
	Execute(ctx context.Context, req *Request) (*Response, error)

	// ValidateConfig validates protocol-specific configuration
	ValidateConfig(config map[string]interface{}) error

	// GetMetrics returns protocol-specific metrics
	GetMetrics() map[string]interface{}

	// Close cleans up protocol resources
	Close() error
}

// ProtocolFactory creates protocol instances
type ProtocolFactory interface {
	CreateProtocol(config map[string]interface{}) (Protocol, error)
	SupportedProtocols() []string
}
