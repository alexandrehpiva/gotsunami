package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version command
func NewVersionCommand(version, buildTime string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display version information including build details and runtime environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			showVersion(version, buildTime)
		},
	}

	return cmd
}

// showVersion displays version information
func showVersion(version, buildTime string) {
	fmt.Printf("GoTsunami %s\n", version)
	fmt.Printf("Build Time: %s\n", buildTime)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Compiler: %s\n", runtime.Compiler)
}
