package compute

import (
	"fmt"
	"os"

	"github.com/motain/compass-compute/internal/facts"
	"github.com/motain/compass-compute/internal/services"
)

func Process(componentName string, verbose bool, compass *services.CompassService) error {
	if verbose {
		fmt.Printf("Starting compass-compute with component: %s\n", componentName)
	}

	component, err := compass.GetComponent(componentName)
	if err != nil {
		return fmt.Errorf("failed to get component '%s': %w", componentName, err)
	}

	if verbose {
		fmt.Printf("Found component '%s' (ID: %s, Type: %s) with %d metrics\n",
			component.Name, component.ID, component.Type, len(component.Metrics))
	}

	cloner := services.NewGitHubCloner(os.Getenv("GITHUB_TOKEN"))

	skipCatalogRepo, err := cloner.SetupMetricDirectory(verbose)
	if err != nil {
		return fmt.Errorf("failed to setup metric directory: %w", err)
	}

	repos := []string{componentName}
	if !skipCatalogRepo {
		repos = append(repos, services.CatalogRepo)
	}

	for _, repo := range repos {
		if err := cloner.Clone(services.GitHubOrg, repo, services.LocalBasePath); err != nil {
			return fmt.Errorf("failed to clone repository '%s': %w", repo, err)
		}
		if verbose {
			fmt.Printf("Successfully cloned repository: %s\n", repo)
		}
	}

	metricPath := services.GetMetricLocalPath()
	if _, err := os.Stat(metricPath); os.IsNotExist(err) {
		return fmt.Errorf("metric directory not found at: %s", metricPath)
	}

	if verbose {
		fmt.Printf("Using metric directory: %s\n", metricPath)
	}

	// Process metrics
	processed := 0
	for _, metric := range component.Metrics {
		if verbose {
			fmt.Printf("Processing metric: %s\n", metric.Name)
		}

		metricFacts, err := compass.GetMetricFacts(metric.Name, component.Type)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: failed to get metric facts for '%s': %v\n", metric.Name, err)
			}
			continue
		}

		evaluatedResult, err := facts.EvaluateMetric(metricFacts, component.Name)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: failed to evaluate metric '%s': %v\n", metric.Name, err)
			}
			continue
		}

		value := fmt.Sprintf("%v", evaluatedResult)

		if verbose {
			fmt.Printf("Evaluated metric '%s' with value: %s\n", metric.Name, value)
		}

		if err := compass.PutMetric(component.ID, metric.DefinitionID, value); err != nil {
			fmt.Printf("Error submitting metric '%s': %v\n", metric.Name, err)
			continue
		}

		processed++
	}

	fmt.Printf("Successfully processed %d metrics for component '%s'\n", processed, componentName)
	return nil
}

func ProcessAll(componentList []string, verbose bool, allComponents bool) error {
	compass := services.NewCompassService()
	if allComponents {
		if verbose {
			list, err := compass.GetAllComponentList()
			if err != nil {
				return err
			}
			fmt.Printf("Processing all components: %v\n", list)
			return nil
		}
	}

	for _, componentName := range componentList {
		if err := Process(componentName, verbose, compass); err != nil {
			return fmt.Errorf("failed to process component '%s': %w", componentName, err)
		}
	}

	return nil

}
