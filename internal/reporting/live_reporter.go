package reporting

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandredias/gotsunami/internal/metrics"
)

// LiveReporter displays real-time metrics during load testing
type LiveReporter struct {
	collector *metrics.Collector
	interval  time.Duration
	stopChan  chan bool
}

// NewLiveReporter creates a new live reporter
func NewLiveReporter(collector *metrics.Collector, interval time.Duration) *LiveReporter {
	return &LiveReporter{
		collector: collector,
		interval:  interval,
		stopChan:  make(chan bool),
	}
}

// Start begins live reporting
func (r *LiveReporter) Start() {
	go r.reportLoop()
}

// Stop stops live reporting
func (r *LiveReporter) Stop() {
	r.stopChan <- true
}

// reportLoop runs the reporting loop
func (r *LiveReporter) reportLoop() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	// Clear screen and show initial header
	r.clearScreen()
	r.printHeader()

	for {
		select {
		case <-ticker.C:
			r.updateDisplay()
		case <-r.stopChan:
			r.printFinalSummary()
			return
		}
	}
}

// clearScreen clears the terminal screen
func (r *LiveReporter) clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// printHeader prints the live report header
func (r *LiveReporter) printHeader() {
	fmt.Println("┌─────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                           GoTsunami Live Report                              │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│  Time: " + time.Now().Format("15:04:05") + strings.Repeat(" ", 55) + "│")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
}

// updateDisplay updates the live display with current metrics
func (r *LiveReporter) updateDisplay() {
	summary := r.collector.GetSummary()

	// Move cursor to beginning of metrics area
	fmt.Print("\033[5;1H")

	// Print metrics
	fmt.Printf("┌─ Requests ──────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│  Total: %-10d  │  Success: %-10d  │  Failed: %-10d  │  Rate: %6.2f%% │\n",
		summary.TotalRequests, summary.SuccessfulRequests, summary.FailedRequests, summary.SuccessRate)
	fmt.Printf("└─────────────────────────────────────────────────────────────────────────────┘\n")

	if summary.Latency != nil {
		fmt.Printf("┌─ Latency ──────────────────────────────────────────────────────────────────┐\n")
		fmt.Printf("│  Mean: %-8s  │  P90: %-8s  │  P95: %-8s  │  P99: %-8s  │\n",
			summary.Latency.Mean.String(), summary.Latency.P90.String(),
			summary.Latency.P95.String(), summary.Latency.P99.String())
		fmt.Printf("└─────────────────────────────────────────────────────────────────────────────┘\n")
	}

	fmt.Printf("┌─ Throughput ────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│  Requests/sec: %8.2f  │  Bytes/sec: %12.0f  │\n",
		summary.RequestsPerSecond, summary.BytesPerSecond)
	fmt.Printf("└─────────────────────────────────────────────────────────────────────────────┘\n")

	// Print status codes
	if len(summary.StatusCodes) > 0 {
		fmt.Printf("┌─ Status Codes ─────────────────────────────────────────────────────────────┐\n")
		statusLine := "│  "
		count := 0
		for code, num := range summary.StatusCodes {
			if count > 0 {
				statusLine += "  │  "
			}
			statusLine += fmt.Sprintf("%d: %d", code, num)
			count++
			if count >= 6 { // Limit to 6 status codes per line
				break
			}
		}
		statusLine += strings.Repeat(" ", 60-len(statusLine)) + "│"
		fmt.Printf("%s\n", statusLine)
		fmt.Printf("└─────────────────────────────────────────────────────────────────────────────┘\n")
	}

	// Print errors if any
	if len(summary.Errors) > 0 {
		fmt.Printf("┌─ Errors ───────────────────────────────────────────────────────────────────┐\n")
		errorCount := 0
		for errorType, count := range summary.Errors {
			if errorCount >= 3 { // Limit to 3 errors
				fmt.Printf("│  ... and %d more error types\n", len(summary.Errors)-3)
				break
			}
			fmt.Printf("│  %s: %d\n", errorType, count)
			errorCount++
		}
		fmt.Printf("└─────────────────────────────────────────────────────────────────────────────┘\n")
	}

	fmt.Println()
	fmt.Printf("Press Ctrl+C to stop...")
}

// printFinalSummary prints the final summary when stopping
func (r *LiveReporter) printFinalSummary() {
	r.clearScreen()
	summary := r.collector.GetSummary()

	fmt.Println("┌─────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                        GoTsunami Test Complete                              │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────┤")

	fmt.Printf("│  Total Requests: %d\n", summary.TotalRequests)
	fmt.Printf("│  Successful: %d (%.2f%%)\n", summary.SuccessfulRequests, summary.SuccessRate)
	fmt.Printf("│  Failed: %d\n", summary.FailedRequests)
	fmt.Printf("│  Requests/sec: %.2f\n", summary.RequestsPerSecond)

	if summary.Latency != nil {
		fmt.Printf("│  Avg Latency: %s\n", summary.Latency.Mean.String())
		fmt.Printf("│  P95 Latency: %s\n", summary.Latency.P95.String())
	}

	fmt.Println("└─────────────────────────────────────────────────────────────────────────────┘")
}

// PrintProgressBar prints a simple progress bar
func PrintProgressBar(current, total int64, width int) {
	if total == 0 {
		return
	}

	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Printf("\r[%s] %.1f%% (%d/%d)", bar, percentage*100, current, total)
}

// PrintSimpleStats prints simple statistics to stdout
func PrintSimpleStats(summary *metrics.Summary) {
	fmt.Printf("Requests: %d | Success: %.2f%% | RPS: %.2f",
		summary.TotalRequests, summary.SuccessRate, summary.RequestsPerSecond)

	if summary.Latency != nil {
		fmt.Printf(" | Latency: %s", summary.Latency.Mean.String())
	}

	fmt.Println()
}
