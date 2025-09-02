package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alexandredias/gotsunami/internal/protocols"
)

// HTTPClient implements the Protocol interface for HTTP/HTTPS
type HTTPClient struct {
	client    *http.Client
	transport *http.Transport
	config    *Config
	metrics   *Metrics
}

// Config holds HTTP client configuration
type Config struct {
	Timeout        time.Duration
	KeepAlive      bool
	MaxConnections int
	TLSSkipVerify  bool
	Proxy          string
	UserAgent      string
}

// Metrics holds HTTP-specific metrics
type Metrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalBytes         int64
	AverageLatency     time.Duration
	MaxLatency         time.Duration
	MinLatency         time.Duration
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(config *Config) *HTTPClient {
	transport := &http.Transport{
		MaxIdleConns:        config.MaxConnections,
		MaxIdleConnsPerHost: config.MaxConnections / 2,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.TLSSkipVerify,
		},
		DisableKeepAlives: !config.KeepAlive,
	}

	// Configure proxy if provided
	if config.Proxy != "" {
		transport.Proxy = http.ProxyURL(&url.URL{
			Scheme: "http",
			Host:   config.Proxy,
		})
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &HTTPClient{
		client:    client,
		transport: transport,
		config:    config,
		metrics:   &Metrics{},
	}
}

// Name returns the protocol name
func (c *HTTPClient) Name() string {
	return "HTTP"
}

// Version returns the protocol version
func (c *HTTPClient) Version() string {
	return "1.1"
}

// Execute performs an HTTP request
func (c *HTTPClient) Execute(ctx context.Context, req *protocols.Request) (*protocols.Response, error) {
	start := time.Now()

	// Create HTTP request
	httpReq, err := c.createHTTPRequest(ctx, req)
	if err != nil {
		return c.createErrorResponse(err, time.Since(start)), nil
	}

	// Execute request
	httpResp, err := c.client.Do(httpReq)
	responseTime := time.Since(start)

	if err != nil {
		c.metrics.FailedRequests++
		return c.createErrorResponse(err, responseTime), nil
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.metrics.FailedRequests++
		return c.createErrorResponse(err, responseTime), nil
	}

	// Update metrics
	c.updateMetrics(responseTime, len(body), httpResp.StatusCode)

	// Create response
	resp := &protocols.Response{
		StatusCode:    httpResp.StatusCode,
		Headers:       c.extractHeaders(httpResp.Header),
		Body:          body,
		ResponseTime:  responseTime,
		ContentLength: int64(len(body)),
	}

	return resp, nil
}

// createHTTPRequest creates an HTTP request from a protocol request
func (c *HTTPClient) createHTTPRequest(ctx context.Context, req *protocols.Request) (*http.Request, error) {
	// Build URL with query parameters
	url := req.URL
	if len(req.QueryParams) > 0 {
		url = c.buildURLWithParams(url, req.QueryParams)
	}

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, strings.NewReader(string(req.Body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set User-Agent if not provided
	if httpReq.Header.Get("User-Agent") == "" && c.config.UserAgent != "" {
		httpReq.Header.Set("User-Agent", c.config.UserAgent)
	}

	return httpReq, nil
}

// buildURLWithParams builds URL with query parameters
func (c *HTTPClient) buildURLWithParams(baseURL string, params map[string]interface{}) string {
	if len(params) == 0 {
		return baseURL
	}

	query := make([]string, 0, len(params))
	for key, value := range params {
		query = append(query, fmt.Sprintf("%s=%v", key, value))
	}

	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}

	return baseURL + separator + strings.Join(query, "&")
}

// extractHeaders extracts headers from HTTP response
func (c *HTTPClient) extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// createErrorResponse creates a response for an error
func (c *HTTPClient) createErrorResponse(err error, responseTime time.Duration) *protocols.Response {
	return &protocols.Response{
		StatusCode:   0,
		Headers:      make(map[string]string),
		Body:         []byte{},
		ResponseTime: responseTime,
		Error:        err,
	}
}

// updateMetrics updates client metrics
func (c *HTTPClient) updateMetrics(responseTime time.Duration, bodySize int, statusCode int) {
	c.metrics.TotalRequests++
	c.metrics.TotalBytes += int64(bodySize)

	if statusCode >= 200 && statusCode < 400 {
		c.metrics.SuccessfulRequests++
	} else {
		c.metrics.FailedRequests++
	}

	// Update latency metrics
	if c.metrics.MinLatency == 0 || responseTime < c.metrics.MinLatency {
		c.metrics.MinLatency = responseTime
	}
	if responseTime > c.metrics.MaxLatency {
		c.metrics.MaxLatency = responseTime
	}

	// Calculate average latency (simplified)
	if c.metrics.TotalRequests > 0 {
		totalLatency := c.metrics.AverageLatency * time.Duration(c.metrics.TotalRequests-1)
		c.metrics.AverageLatency = (totalLatency + responseTime) / time.Duration(c.metrics.TotalRequests)
	}
}

// ValidateConfig validates HTTP client configuration
func (c *HTTPClient) ValidateConfig(config map[string]interface{}) error {
	// TODO: Implement configuration validation
	return nil
}

// GetMetrics returns HTTP-specific metrics
func (c *HTTPClient) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_requests":      c.metrics.TotalRequests,
		"successful_requests": c.metrics.SuccessfulRequests,
		"failed_requests":     c.metrics.FailedRequests,
		"total_bytes":         c.metrics.TotalBytes,
		"average_latency":     c.metrics.AverageLatency.String(),
		"max_latency":         c.metrics.MaxLatency.String(),
		"min_latency":         c.metrics.MinLatency.String(),
	}
}

// Close cleans up HTTP client resources
func (c *HTTPClient) Close() error {
	if c.transport != nil {
		c.transport.CloseIdleConnections()
	}
	return nil
}
