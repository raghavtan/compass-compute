package compute

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "compass-compute <component-name>",
	Short: "A tool to manage compass components",
	Long: `compass-compute is a CLI tool for managing compass components.
It processes components by validating them and executing various operations.`,
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		componentName := strings.TrimSpace(args[0])
		if componentName == "" {
			return fmt.Errorf("component name cannot be empty")
		}

		if err := validateComponentName(componentName); err != nil {
			return fmt.Errorf("invalid component name: %w", err)
		}

		config = &Config{
			ComponentName: componentName,
			Verbose:       verbose,
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(config)
	},
	Example: `  compass-compute my-component
  compass-compute --verbose my-component`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

// Processing functions
func run(config *Config) error {
	if config.Verbose {
		fmt.Printf("Starting compass-compute with component: %s\n", config.ComponentName)
	}

	fmt.Printf("Processing component: %s\n", config.ComponentName)

	// GET component ID
	// GET metricName, metricDefinitionID for all metrics associated with the component
	// GET FactDefinitions for all metrics
	// Clone component repository
	// For each metric:
	// 		compute all Facts to generate metric values for each metric
	// 		Push computed metric values to the Compass API

	if config.Verbose {
		fmt.Printf("  - Validating component '%s'\n", config.ComponentName)
		fmt.Printf("  - Executing component operations\n")
		fmt.Printf("  - Finalizing component processing\n")
	}

	fmt.Printf("Component '%s' processed successfully\n", config.ComponentName)

	if config.Verbose {
		fmt.Println("Component processing completed successfully")
	}

	return nil
}
