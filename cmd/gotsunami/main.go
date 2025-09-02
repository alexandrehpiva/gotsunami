package main

import (
	"os"

	"github.com/alexandredias/gotsunami/internal/cli"
	"github.com/sirupsen/logrus"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// Configure logging
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Create and execute CLI
	rootCmd := cli.NewRootCommand(version, buildTime)

	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Error("Command execution failed")
		os.Exit(1)
	}
}
