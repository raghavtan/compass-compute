package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/motain/compass-compute/internal/compute"
	"github.com/spf13/cobra"
)

var computeCmd = &cobra.Command{
	Use:   "compute [component-name]",
	Short: "Manage compass component metrics",
	Long: `The compute command allows you to manage metrics for a specific compass component or all components.

REQUIRED ENVIRONMENT VARIABLES:
  GITHUB_TOKEN       GitHub personal access token for repository access
  COMPASS_API_TOKEN  Compass API authentication token
  COMPASS_CLOUD_ID   Compass cloud instance identifier
  AWS_REGION         AWS region for cloud resources (e.g., us-east-1)
  AWS_ROLE           AWS IAM role ARN for authentication

OPTIONAL ENVIRONMENT VARIABLES:
  METRIC_DIR         Override metric directory source:
                     - Local path: /path/to/local/metrics
                     - Git repo: https://github.com/owner/repo.git/path/to/metrics
                     - Git SSH: git@github.com:owner/repo.git/path/to/metrics
                     - GitHub tree: https://github.com/owner/repo/tree/branch/path/to/metrics`,
	Args: func(cmd *cobra.Command, args []string) error {
		if allComponents {
			if len(args) > 0 {
				return fmt.Errorf("cannot specify component names when using -A/--all flag")
			}
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("requires exactly one component name argument when not using -A/--all flag")
		}
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validateEnvironmentVariables()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if allComponents {
			return compute.ProcessAll(nil, verbose, allComponents)
		}
		return compute.ProcessAll(strings.Split(args[0], ","), verbose, false)
	},
}

func validateEnvironmentVariables() error {
	requiredEnvVars := []string{
		"GITHUB_TOKEN",
		"COMPASS_API_TOKEN",
		"COMPASS_CLOUD_ID",
		"AWS_REGION",
	}

	var missing []string
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}
