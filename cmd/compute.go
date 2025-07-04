package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/motain/compass-compute/internal/compute"
	"github.com/spf13/cobra"
)

var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Manage compass component metrics",
	Long:  `The compute command allows you to manage metrics for a specific compass component.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		names := strings.Split(args[0], ",")
		if err := validateComponentName(names); err != nil {
			return fmt.Errorf("invalid component name: %w", err)
		}
		return AllCompute(names)
	},
}

var componentNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,100}$`)

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
		return fmt.Errorf("the 'all' option is not Implemented yet, please specify component names")
	}
	for _, component := range componentList {
		config := &compute.Config{
			ComponentName: component,
			Verbose:       verbose,
		}
		err := compute.Process(config)
		if err != nil {
			_ = fmt.Errorf("failed to process component '%s': %w", component, err)
			continue
		}
	}
	return nil
}
