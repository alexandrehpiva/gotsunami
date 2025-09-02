package engine

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Worker represents a load testing worker
type Worker struct {
	id       int
	engine   *LoadEngine
	requests int
	mu       sync.Mutex
}

// NewWorker creates a new worker
func NewWorker(id int, engine *LoadEngine) *Worker {
	return &Worker{
		id:     id,
		engine: engine,
	}
}

// Run executes the worker's load testing loop
func (w *Worker) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	logrus.Debugf("Worker %d started", w.id)

	// Calculate load pattern
	pattern := w.calculateLoadPattern()

	// Execute requests according to pattern
	for {
		select {
		case <-w.engine.GetContext().Done():
			logrus.Debugf("Worker %d stopping", w.id)
			return
		default:
			// Check if we've reached max requests
			if w.engine.GetConfig().MaxRequests > 0 && w.requests >= w.engine.GetConfig().MaxRequests {
				logrus.Debugf("Worker %d reached max requests (%d)", w.id, w.requests)
				return
			}

			// Calculate delay based on pattern
			delay := w.calculateDelay(pattern)
			if delay > 0 {
				time.Sleep(delay)
			}

			// Execute request
			w.executeRequest()

			// Apply delay between requests
			if w.engine.GetConfig().Delay > 0 {
				time.Sleep(w.engine.GetConfig().Delay)
			}
		}
	}
}

// calculateLoadPattern calculates the load pattern for this worker
func (w *Worker) calculateLoadPattern() *LoadPattern {
	config := w.engine.GetConfig()
	pattern := &LoadPattern{
		Type: config.Pattern,
	}

	switch config.Pattern {
	case "spike":
		pattern = w.calculateSpikePattern()
	case "steady":
		pattern = w.calculateSteadyPattern()
	case "ramp-up":
		pattern = w.calculateRampUpPattern()
	case "stress":
		pattern = w.calculateStressPattern()
	default:
		pattern = w.calculateSteadyPattern()
	}

	return pattern
}

// calculateSpikePattern calculates spike load pattern
func (w *Worker) calculateSpikePattern() *LoadPattern {
	config := w.engine.GetConfig()
	duration := config.Duration

	return &LoadPattern{
		Type: "spike",
		Phases: []LoadPhase{
			{
				Duration:  duration / 4,
				Intensity: 0.2, // 20% of max load
			},
			{
				Duration:  duration / 4,
				Intensity: 1.0, // 100% of max load (spike)
			},
			{
				Duration:  duration / 2,
				Intensity: 0.2, // Back to 20%
			},
		},
	}
}

// calculateSteadyPattern calculates steady load pattern
func (w *Worker) calculateSteadyPattern() *LoadPattern {
	config := w.engine.GetConfig()

	return &LoadPattern{
		Type: "steady",
		Phases: []LoadPhase{
			{
				Duration:  config.RampUp,
				Intensity: 0.0, // Ramp up from 0
			},
			{
				Duration:  config.Duration - config.RampUp - config.RampDown,
				Intensity: 1.0, // Full load
			},
			{
				Duration:  config.RampDown,
				Intensity: 0.0, // Ramp down to 0
			},
		},
	}
}

// calculateRampUpPattern calculates ramp-up load pattern
func (w *Worker) calculateRampUpPattern() *LoadPattern {
	config := w.engine.GetConfig()
	duration := config.Duration

	return &LoadPattern{
		Type: "ramp-up",
		Phases: []LoadPhase{
			{
				Duration:  duration,
				Intensity: 0.0, // Linear ramp from 0 to 1
			},
		},
	}
}

// calculateStressPattern calculates stress test pattern
func (w *Worker) calculateStressPattern() *LoadPattern {
	config := w.engine.GetConfig()
	duration := config.Duration

	return &LoadPattern{
		Type: "stress",
		Phases: []LoadPhase{
			{
				Duration:  duration / 3,
				Intensity: 0.5, // 50% load
			},
			{
				Duration:  duration / 3,
				Intensity: 1.0, // 100% load
			},
			{
				Duration:  duration / 3,
				Intensity: 1.5, // 150% load (stress)
			},
		},
	}
}

// calculateDelay calculates the delay between requests based on load pattern
func (w *Worker) calculateDelay(pattern *LoadPattern) time.Duration {
	config := w.engine.GetConfig()
	elapsed := time.Since(time.Now().Add(-config.Duration))

	// Find current phase
	var currentPhase *LoadPhase
	var phaseStart time.Duration

	for _, phase := range pattern.Phases {
		if elapsed < phaseStart+phase.Duration {
			currentPhase = &phase
			break
		}
		phaseStart += phase.Duration
	}

	if currentPhase == nil {
		return 0 // No delay if no active phase
	}

	// Calculate intensity for current time
	intensity := w.calculateIntensity(currentPhase, elapsed-phaseStart)

	// Convert intensity to delay (higher intensity = lower delay)
	baseDelay := 100 * time.Millisecond
	delay := time.Duration(float64(baseDelay) / intensity)

	return delay
}

// calculateIntensity calculates the current intensity based on phase and time
func (w *Worker) calculateIntensity(phase *LoadPhase, elapsed time.Duration) float64 {
	if phase.Duration == 0 {
		return phase.Intensity
	}

	// Linear interpolation for ramp phases
	progress := float64(elapsed) / float64(phase.Duration)
	if progress > 1.0 {
		progress = 1.0
	}

	// For ramp-up pattern, intensity increases linearly
	if w.engine.GetConfig().Pattern == "ramp-up" {
		return progress
	}

	// For other patterns, use phase intensity
	return phase.Intensity
}

// executeRequest executes a single request
func (w *Worker) executeRequest() {
	w.mu.Lock()
	w.requests++
	requestNum := w.requests
	w.mu.Unlock()

	// Create request
	req := w.engine.CreateRequest()

	// Execute request
	ctx, cancel := context.WithTimeout(w.engine.GetContext(), req.Timeout)
	defer cancel()

	resp, err := w.engine.GetProtocol().Execute(ctx, req)
	if err != nil {
		logrus.WithError(err).Debugf("Worker %d request %d failed", w.id, requestNum)
	}

	// Record response
	w.engine.RecordResponse(resp)
}

// GetRequestCount returns the number of requests executed by this worker
func (w *Worker) GetRequestCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.requests
}

// LoadPattern represents a load testing pattern
type LoadPattern struct {
	Type   string      `json:"type"`
	Phases []LoadPhase `json:"phases"`
}

// LoadPhase represents a phase in a load pattern
type LoadPhase struct {
	Duration  time.Duration `json:"duration"`
	Intensity float64       `json:"intensity"` // 0.0 to 2.0 (0% to 200% of base load)
}
