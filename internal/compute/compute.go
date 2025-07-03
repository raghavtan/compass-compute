package compute

import (
	"fmt"
	"os"

	"github.com/motain/compass-compute/internal/services"
)

type Config struct {
	ComponentName string
	Verbose       bool
}

func Process(config *Config) error {
	if config.Verbose {
		fmt.Printf("Starting compass-compute with component: %s\n", config.ComponentName)
	}

	compass := services.NewCompassService()

	// Get component with metrics
	component, err := compass.GetComponent(config.ComponentName)
	if err != nil {
		return fmt.Errorf("failed to get component '%s': %w", config.ComponentName, err)
	}

	if config.Verbose {
		fmt.Printf("Found component '%s' (ID: %s, Type: %s) with %d metrics\n",
			component.Name, component.ID, component.Type, len(component.Metrics))
	}

	// Clone repositories
	cloner := services.NewGitHubCloner(os.Getenv("GITHUB_TOKEN"))
	repos := []string{config.ComponentName, services.CatalogRepo}

	for _, repo := range repos {
		if err := cloner.Clone("motain", repo, "./repos/"); err != nil {
			return fmt.Errorf("failed to clone repository '%s': %w", repo, err)
		}
		if config.Verbose {
			fmt.Printf("Successfully cloned repository: %s\n", repo)
		}
	}

	// Process metrics
	processed := 0
	for _, metric := range component.Metrics {
		if config.Verbose {
			fmt.Printf("Processing metric: %s (ID: %s)\n", metric.Name, metric.DefinitionID)
		}

		_, err := compass.GetMetricFacts(metric.Name, component.Type)
		if err != nil {
			if config.Verbose {
				fmt.Printf("Warning: failed to get metric facts for '%s': %v\n", metric.Name, err)
			}
			continue
		}

		// TODO: Implement facts.EvaluateMetric when available
		value := "1"

		if config.Verbose {
			fmt.Printf("Evaluated metric '%s' with value: %s\n", metric.Name, value)
		}

		if err := compass.PutMetric(component.ID, metric.DefinitionID, value); err != nil {
			fmt.Printf("Error submitting metric '%s': %v\n", metric.Name, err)
			continue
		}

		if config.Verbose {
			fmt.Printf("Successfully submitted metric '%s'\n", metric.Name)
		}
		processed++
	}

	fmt.Printf("Successfully processed %d metrics for component '%s'\n", processed, config.ComponentName)
	return nil
}
