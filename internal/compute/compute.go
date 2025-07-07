package compute

import (
	"fmt"
	"os"

	"github.com/motain/compass-compute/internal/facts"
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

	component, err := compass.GetComponent(config.ComponentName)
	if err != nil {
		return fmt.Errorf("failed to get component '%s': %w", config.ComponentName, err)
	}

	if config.Verbose {
		fmt.Printf("Found component '%s' (ID: %s, Type: %s) with %d metrics\n",
			component.Name, component.ID, component.Type, len(component.Metrics))
	}

	cloner := services.NewGitHubCloner(os.Getenv("GITHUB_TOKEN"))

	skipCatalogRepo, err := cloner.SetupMetricDirectory(config.Verbose)
	if err != nil {
		return fmt.Errorf("failed to setup metric directory: %w", err)
	}

	repos := []string{config.ComponentName}
	if !skipCatalogRepo {
		repos = append(repos, services.CatalogRepo)
	}

	for _, repo := range repos {
		if err := cloner.Clone(services.GitHubOrg, repo, services.LocalBasePath); err != nil {
			return fmt.Errorf("failed to clone repository '%s': %w", repo, err)
		}
		if config.Verbose {
			fmt.Printf("Successfully cloned repository: %s\n", repo)
		}
	}

	metricPath := services.GetMetricLocalPath()
	if _, err := os.Stat(metricPath); os.IsNotExist(err) {
		return fmt.Errorf("metric directory not found at: %s", metricPath)
	}

	if config.Verbose {
		fmt.Printf("Using metric directory: %s\n", metricPath)
	}

	// Process metrics
	processed := 0
	for _, metric := range component.Metrics {
		if config.Verbose {
			fmt.Printf("Processing metric: %s\n", metric.Name)
		}

		metricFacts, err := compass.GetMetricFacts(metric.Name, component.Type)
		if err != nil {
			if config.Verbose {
				fmt.Printf("Warning: failed to get metric facts for '%s': %v\n", metric.Name, err)
			}
			continue
		}

		evaluatedResult, err := facts.EvaluateMetric(metricFacts, component.Name)
		if err != nil {
			if config.Verbose {
				fmt.Printf("Warning: failed to evaluate metric '%s': %v\n", metric.Name, err)
			}
			continue
		}

		value := fmt.Sprintf("%v", evaluatedResult)

		if config.Verbose {
			fmt.Printf("Evaluated metric '%s' with value: %s\n", metric.Name, value)
		}

		if err := compass.PutMetric(component.ID, metric.DefinitionID, value); err != nil {
			fmt.Printf("Error submitting metric '%s': %v\n", metric.Name, err)
			continue
		}

		processed++
	}

	fmt.Printf("Successfully processed %d metrics for component '%s'\n", processed, config.ComponentName)
	return nil
}
