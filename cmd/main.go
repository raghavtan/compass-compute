package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.AddCommand(computeCmd)
}

var rootCmd = &cobra.Command{
	Use:   "compass-compute <component-name>",
	Short: "A tool to compute compass component metrics",
	Long: `compass-compute is a CLI tool for computing compass component metrics.

ENVIRONMENT VARIABLES:
  GITHUB_TOKEN       GitHub personal access token for repository access
  COMPASS_API_TOKEN  Compass API authentication token
  COMPASS_CLOUD_ID   Compass cloud instance identifier
  AWS_REGION         AWS region for cloud resources (e.g., us-east-1)
  AWS_ROLE           AWS IAM role ARN for authentication

All environment variables are required for proper operation.`,
	Example: `  # Compute metrics for a single component
  compass-compute compute my-component
  
  # Compute metrics for multiple components
  compass-compute compute my-component1,my-component2
  
  # Enable verbose output
  compass-compute compute my-component --verbose
  
  # Compute metrics for all components (when implemented)
  compass-compute compute all
  compass-compute compute all --verbose
  
  # Set required environment variables
  export GITHUB_TOKEN="your-github-token"
  export COMPASS_API_TOKEN="your-compass-token"
  export COMPASS_CLOUD_ID="your-cloud-id"
  export AWS_REGION="us-east-1"
  export AWS_ROLE="arn:aws:iam::123456789012:role/CompassRole"`,
}
