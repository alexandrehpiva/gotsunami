package engine

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/alexandredias/gotsunami/internal/config"
	"github.com/alexandredias/gotsunami/internal/metrics"
	"github.com/alexandredias/gotsunami/internal/protocols"
	"github.com/alexandredias/gotsunami/internal/protocols/http"
	"github.com/alexandredias/gotsunami/internal/validation"
	"github.com/sirupsen/logrus"
)

// LoadEngine orchestrates the load testing process
type LoadEngine struct {
	config    *config.LoadTestConfig
	scenario  *config.Scenario
	protocol  protocols.Protocol
	collector *metrics.Collector
	validator *validation.ResponseValidator
	workers   []*Worker
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewLoadEngine creates a new load testing engine
func NewLoadEngine(cfg *config.LoadTestConfig, scenario *config.Scenario) (*LoadEngine, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Duration)

	// Create HTTP client
	httpConfig := &http.Config{
		Timeout:        cfg.Timeout,
		KeepAlive:      cfg.KeepAlive,
		MaxConnections: cfg.Connections,
		TLSSkipVerify:  cfg.TLSSkipVerify,
		Proxy:          cfg.Proxy,
		UserAgent:      cfg.UserAgent,
	}

	protocol := http.NewHTTPClient(httpConfig)
	collector := metrics.NewCollector()
	validator := validation.NewResponseValidator(scenario.GetValidationConfig())

	// Determine number of workers
	workers := cfg.Workers
	if workers == 0 {
		workers = runtime.NumCPU()
	}

	engine := &LoadEngine{
		config:    cfg,
		scenario:  scenario,
		protocol:  protocol,
		collector: collector,
		validator: validator,
		workers:   make([]*Worker, workers),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Create workers
	for i := 0; i < workers; i++ {
		engine.workers[i] = NewWorker(i, engine)
	}

	return engine, nil
}

// Run executes the load test
func (e *LoadEngine) Run() (*metrics.Summary, error) {
	logrus.Info("Starting load test...")
	logrus.Infof("Configuration: %d VUs, %v duration, %s pattern",
		e.config.VirtualUsers, e.config.Duration, e.config.Pattern)

	// Start metrics collection
	e.collector.Start()

	// Start workers
	for _, worker := range e.workers {
		e.wg.Add(1)
		go worker.Run(&e.wg)
	}

	// Wait for completion or timeout
	select {
	case <-e.ctx.Done():
		logrus.Info("Load test completed")
	case <-time.After(e.config.Duration + 5*time.Second):
		logrus.Warn("Load test timeout exceeded")
	}

	// Stop metrics collection
	e.collector.Stop()

	// Wait for all workers to finish
	e.wg.Wait()

	// Clean up
	e.protocol.Close()

	// Get final summary
	summary := e.collector.GetSummary()

	logrus.Infof("Load test completed: %d requests, %.2f%% success rate, %.2f req/s",
		summary.TotalRequests, summary.SuccessRate, summary.RequestsPerSecond)

	return summary, nil
}

// Stop gracefully stops the load test
func (e *LoadEngine) Stop() {
	logrus.Info("Stopping load test...")
	e.cancel()
}

// GetCollector returns the metrics collector
func (e *LoadEngine) GetCollector() *metrics.Collector {
	return e.collector
}

// GetContext returns the engine context
func (e *LoadEngine) GetContext() context.Context {
	return e.ctx
}

// GetConfig returns the load test configuration
func (e *LoadEngine) GetConfig() *config.LoadTestConfig {
	return e.config
}

// GetScenario returns the scenario configuration
func (e *LoadEngine) GetScenario() *config.Scenario {
	return e.scenario
}

// GetProtocol returns the protocol instance
func (e *LoadEngine) GetProtocol() protocols.Protocol {
	return e.protocol
}

// GetValidator returns the response validator
func (e *LoadEngine) GetValidator() *validation.ResponseValidator {
	return e.validator
}

// CreateRequest creates a protocol request from the scenario
func (e *LoadEngine) CreateRequest() *protocols.Request {
	// Build full URL
	fullURL := e.scenario.BaseURL + e.scenario.URL

	// Convert body to bytes if needed
	var bodyBytes []byte
	if e.scenario.Body != nil {
		// TODO: Handle different body types (JSON, form data, etc.)
		bodyBytes = []byte(fmt.Sprintf("%v", e.scenario.Body))
	}

	// Convert query params to string map
	queryParams := make(map[string]interface{})
	for key, value := range e.scenario.QueryParams {
		queryParams[key] = value
	}

	return &protocols.Request{
		Method:      e.scenario.Method,
		URL:         fullURL,
		Headers:     e.scenario.Headers,
		Body:        bodyBytes,
		Timeout:     e.scenario.GetTimeout(),
		QueryParams: queryParams,
	}
}

// RecordResponse records a response in the metrics collector
func (e *LoadEngine) RecordResponse(resp *protocols.Response) {
	// Validate response
	validationResult := e.validator.Validate(resp)
	e.collector.RecordValidation(validationResult.Passed, validationResult.ErrorType)

	// Record response metrics
	e.collector.RecordResponse(resp)
}
