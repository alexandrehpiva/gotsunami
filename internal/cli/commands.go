package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCommand creates the root command for GoTsunami CLI
func NewRootCommand(version, buildTime string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gotsunami",
		Short: "GoTsunami - Enterprise-grade load testing tool for REST APIs",
		Long: `GoTsunami is a professional CLI tool for load testing REST APIs.
It provides comprehensive testing capabilities with real-time metrics,
advanced validation, and detailed reporting for production environments.`,
		Version: fmt.Sprintf("%s (built %s)", version, buildTime),
	}

	// Add subcommands
	rootCmd.AddCommand(NewRunCommand())
	rootCmd.AddCommand(NewValidateCommand())
	rootCmd.AddCommand(NewVersionCommand(version, buildTime))

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.gotsunami.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet mode (only errors)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")

	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))

	// Initialize configuration
	cobra.OnInitialize(initConfig)

	return rootCmd
}

// initConfig initializes the configuration
func initConfig() {
	// Set config file if provided
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".gotsunami" (without extension)
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gotsunami")
	}

	// Environment variables
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
