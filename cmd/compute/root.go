package compute

import (
	"fmt"
	"os"
	"strings"

	"github.com/motain/compass-compute/internal/compute"
	"github.com/spf13/cobra"
)

var (
	verbose    bool
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "compass-compute <component-name>",
	Short: "A tool to manage compass components",
	Long:  `compass-compute is a CLI tool for managing compass components.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if err := validateComponentName(name); err != nil {
			return fmt.Errorf("invalid component name: %w", err)
		}

		config := &Config{
			ComponentName: name,
			Verbose:       verbose,
		}
		return compute.Process((*compute.Config)(config))
	},
	Example: `  compass-compute my-component
  compass-compute --verbose my-component`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("compass-compute\nVersion: %s\nBuilt: %s\nCommit: %s\n",
			Version, BuildTime, CommitHash)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.AddCommand(versionCmd)
}
