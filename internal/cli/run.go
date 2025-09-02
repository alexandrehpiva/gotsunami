package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/alexandredias/gotsunami/internal/config"
	"github.com/alexandredias/gotsunami/internal/engine"
	"github.com/alexandredias/gotsunami/internal/reporting"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRunCommand creates the run command
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <scenario.json>",
		Short: "Run a load test scenario",
		Long: `Run a load test scenario defined in a JSON configuration file.
The scenario file contains all the necessary configuration for the test including
the target URL, request parameters, validation rules, and load patterns.`,
		Args: cobra.ExactArgs(1),
		RunE: runLoadTest,
	}

	// Load test configuration flags
	cmd.Flags().IntP("vus", "u", 10, "number of virtual users (threads)")
	cmd.Flags().DurationP("duration", "d", 30*time.Second, "test duration")
	cmd.Flags().Duration("ramp-up", 10*time.Second, "ramp-up duration")
	cmd.Flags().Duration("ramp-down", 5*time.Second, "ramp-down duration")
	cmd.Flags().Duration("delay", 0, "delay between requests per user")
	cmd.Flags().Int("max-requests", 0, "maximum requests per user (0 = unlimited)")
	cmd.Flags().Duration("timeout", 30*time.Second, "global timeout for requests")

	// Load patterns
	cmd.Flags().String("pattern", "steady", "load pattern (spike, steady, ramp-up, stress)")

	// Output configuration
	cmd.Flags().Bool("live", false, "show real-time metrics in terminal")
	cmd.Flags().String("report-format", "json", "report format (json, yaml, csv)")
	cmd.Flags().String("outfile", "", "output file for report")
	cmd.Flags().Bool("stdout", false, "force output to stdout (for CI/CD)")

	// Validation flags
	cmd.Flags().IntSlice("expect-status", []int{200}, "expected status codes")
	cmd.Flags().String("expect-body", "", "content that should be in response body")
	cmd.Flags().String("expect-body-not", "", "content that should NOT be in response body")
	cmd.Flags().Duration("expect-response-time", 0, "maximum expected response time")

	// Advanced configuration
	cmd.Flags().Int("workers", 0, "number of workers (0 = CPU cores)")
	cmd.Flags().Int("connections", 100, "HTTP connection pool size")
	cmd.Flags().Bool("keep-alive", true, "keep HTTP connections alive")
	cmd.Flags().Bool("disable-keep-alive", false, "disable HTTP keep-alive")
	cmd.Flags().Bool("tls-skip-verify", false, "skip TLS verification (testing only)")
	cmd.Flags().String("proxy", "", "HTTP/HTTPS proxy")
	cmd.Flags().String("user-agent", "GoTsunami/1.0", "custom user agent")

	// Bind flags to viper
	viper.BindPFlag("run.vus", cmd.Flags().Lookup("vus"))
	viper.BindPFlag("run.duration", cmd.Flags().Lookup("duration"))
	viper.BindPFlag("run.ramp_up", cmd.Flags().Lookup("ramp-up"))
	viper.BindPFlag("run.ramp_down", cmd.Flags().Lookup("ramp-down"))
	viper.BindPFlag("run.delay", cmd.Flags().Lookup("delay"))
	viper.BindPFlag("run.max_requests", cmd.Flags().Lookup("max-requests"))
	viper.BindPFlag("run.timeout", cmd.Flags().Lookup("timeout"))
	viper.BindPFlag("run.pattern", cmd.Flags().Lookup("pattern"))
	viper.BindPFlag("run.live", cmd.Flags().Lookup("live"))
	viper.BindPFlag("run.report_format", cmd.Flags().Lookup("report-format"))
	viper.BindPFlag("run.outfile", cmd.Flags().Lookup("outfile"))
	viper.BindPFlag("run.stdout", cmd.Flags().Lookup("stdout"))
	viper.BindPFlag("run.expect_status", cmd.Flags().Lookup("expect-status"))
	viper.BindPFlag("run.expect_body", cmd.Flags().Lookup("expect-body"))
	viper.BindPFlag("run.expect_body_not", cmd.Flags().Lookup("expect-body-not"))
	viper.BindPFlag("run.expect_response_time", cmd.Flags().Lookup("expect-response-time"))
	viper.BindPFlag("run.workers", cmd.Flags().Lookup("workers"))
	viper.BindPFlag("run.connections", cmd.Flags().Lookup("connections"))
	viper.BindPFlag("run.keep_alive", cmd.Flags().Lookup("keep-alive"))
	viper.BindPFlag("run.disable_keep_alive", cmd.Flags().Lookup("disable-keep-alive"))
	viper.BindPFlag("run.tls_skip_verify", cmd.Flags().Lookup("tls-skip-verify"))
	viper.BindPFlag("run.proxy", cmd.Flags().Lookup("proxy"))
	viper.BindPFlag("run.user_agent", cmd.Flags().Lookup("user-agent"))

	return cmd
}

// runLoadTest executes the load test
func runLoadTest(cmd *cobra.Command, args []string) error {
	scenarioFile := args[0]

	// Check if scenario file exists
	if _, err := os.Stat(scenarioFile); os.IsNotExist(err) {
		return fmt.Errorf("scenario file not found: %s", scenarioFile)
	}

	// Load scenario configuration
	scenario, err := config.LoadScenarioFromFile(scenarioFile)
	if err != nil {
		return fmt.Errorf("failed to load scenario: %w", err)
	}

	// Create load test configuration
	loadConfig := &config.LoadTestConfig{
		Scenario:      scenario,
		VirtualUsers:  viper.GetInt("run.vus"),
		Duration:      viper.GetDuration("run.duration"),
		RampUp:        viper.GetDuration("run.ramp_up"),
		RampDown:      viper.GetDuration("run.ramp_down"),
		Delay:         viper.GetDuration("run.delay"),
		MaxRequests:   viper.GetInt("run.max_requests"),
		Timeout:       viper.GetDuration("run.timeout"),
		Pattern:       viper.GetString("run.pattern"),
		Live:          viper.GetBool("run.live"),
		ReportFormat:  viper.GetString("run.report_format"),
		Outfile:       viper.GetString("run.outfile"),
		Stdout:        viper.GetBool("run.stdout"),
		Workers:       viper.GetInt("run.workers"),
		Connections:   viper.GetInt("run.connections"),
		KeepAlive:     viper.GetBool("run.keep_alive"),
		TLSSkipVerify: viper.GetBool("run.tls_skip_verify"),
		Proxy:         viper.GetString("run.proxy"),
		UserAgent:     viper.GetString("run.user_agent"),
	}

	// Create and run load engine
	engine, err := engine.NewLoadEngine(loadConfig, scenario)
	if err != nil {
		return fmt.Errorf("failed to create load engine: %w", err)
	}

	// Start live reporting if enabled
	var liveReporter *reporting.LiveReporter
	if loadConfig.Live {
		liveReporter = reporting.NewLiveReporter(engine.GetCollector(), 1*time.Second)
		liveReporter.Start()
		defer liveReporter.Stop()
	}

	// Run the load test
	summary, err := engine.Run()
	if err != nil {
		return fmt.Errorf("load test failed: %w", err)
	}

	// Generate and write report
	reporter := reporting.NewJSONReporter(loadConfig)
	report, err := reporter.GenerateReport(summary, scenario)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Write report
	outfile := loadConfig.Outfile
	if loadConfig.Stdout {
		outfile = ""
	}

	if err := reporter.WriteReport(report, outfile); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Exit with appropriate code based on results
	if summary.SuccessRate < 95.0 {
		os.Exit(2) // Validation failed
	}

	return nil
}
