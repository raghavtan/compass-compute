package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/motain/compass-compute/internal/compute"
	"github.com/spf13/cobra"
)

var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Manage compass component metrics",
	Long: `The compute command allows you to manage metrics for a specific compass component.

REQUIRED ENVIRONMENT VARIABLES:
  GITHUB_TOKEN       GitHub personal access token for repository access
  COMPASS_API_TOKEN  Compass API authentication token
  COMPASS_CLOUD_ID   Compass cloud instance identifier
  AWS_REGION         AWS region for cloud resources (e.g., us-east-1)
  AWS_ROLE           AWS IAM role ARN for authentication`,
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validateEnvironmentVariables()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		names := strings.Split(args[0], ",")
		if err := validateComponentName(names); err != nil {
			return fmt.Errorf("invalid component name: %w", err)
		}
		return AllCompute(names)
	},
}

var componentNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,100}$`)

func validateEnvironmentVariables() error {
	requiredEnvVars := []string{
		"GITHUB_TOKEN",
		"COMPASS_API_TOKEN",
		"COMPASS_CLOUD_ID",
		"AWS_REGION",
		"AWS_ROLE",
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

func validateComponentName(names []string) error {
	if len(names) != 0 {
		for _, name := range names {
			if !componentNameRegex.MatchString(name) {
				return fmt.Errorf("invalid name: must be 1-100 chars, alphanumeric with .-_ only")
			}
		}
		return nil
	}
	return fmt.Errorf("component name cannot be empty")
}

func AllCompute(componentList []string) error {

	if strings.Contains(strings.Join(componentList, ","), "all") {
		// Fetch all components from the catalog
		// Update componentList to include all components
		return fmt.Errorf("the 'all' option is not implemented yet, please specify component names")
	}
	for _, component := range componentList {
		config := &compute.Config{
			ComponentName: component,
			Verbose:       verbose,
		}
		err := compute.Process(config)
		if err != nil {
			if verbose {
				fmt.Printf("Error processing component '%s': %v\n", component, err)
			}
			continue
		}
	}
	return nil
}
