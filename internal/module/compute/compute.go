package compute

import (
	"fmt"
	"github.com/motain/compass-compute/internal/services"
	"github.com/motain/compass-compute/internal/services/compassservice"
	"github.com/motain/compass-compute/internal/services/github"
	"os"
)

func Process(componentName string, verbose bool) error {
	compass := compassservice.NewCompassService()

	// Get component with associated metrics from Compass API
	component, err := compass.GetComponent(componentName)
	if err != nil {
		return fmt.Errorf("failed to get component ID for '%s': %w", componentName, err)
	}

	if verbose {
		fmt.Printf("Found component '%s' (ID: %s, Type: %s) with %d associated metrics\n",
			component.ComponentName, component.ComponentID, component.ComponentType, len(component.AssociatedMetrics))
	}

	// Clone repositories
	cloner := github.NewGitHubCloner(os.Getenv("GITHUB_TOKEN"))
	repos := []string{componentName, services.DefaultCatalogRepo}

	for _, repo := range repos {
		err := cloner.Clone("motain", repo, &github.CloneOptions{Destination: "./repos/"})
		if err != nil {
			return fmt.Errorf("failed to clone repository '%s': %w", repo, err)
		}
		if verbose {
			fmt.Printf("Successfully cloned repository: %s\n", repo)
		}
	}

	// Process each associated metric
	for _, associatedMetric := range component.AssociatedMetrics {
		if verbose {
			fmt.Printf("Processing metric: %s (Definition ID: %s)\n",
				associatedMetric.MetricName, associatedMetric.MetricDefinitionID)
		}

		_, err := compass.GetMetricFactsByName(associatedMetric.MetricName, component.ComponentType)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: failed to get metric facts for '%s': %v\n", associatedMetric.MetricName, err)
			}
			continue
		}

		if verbose {
			fmt.Printf("Found metric facts for '%s'\n", associatedMetric.MetricName)
		}

		// TODO: Uncomment when facts package is available
		// Evaluate the metric using the facts
		// evaluatedMetricValue, err := facts.EvaluateMetric(metricFacts, component.ComponentName)
		// if err != nil {
		//     return fmt.Errorf("failed to evaluate metric '%s' for component '%s': %w",
		//         associatedMetric.MetricName, componentName, err)
		// }

		// TODO: Remove placeholder when facts.EvaluateMetric is implemented
		evaluatedMetricValue := "placeholder_value" // Replace with actual evaluation

		if verbose {
			fmt.Printf("Evaluated metric '%s' with value: %s\n", associatedMetric.MetricName, evaluatedMetricValue)
		}

		// Submit the metric value to Compass
		err = compass.PutMetric(component.ComponentID, associatedMetric.MetricDefinitionID, evaluatedMetricValue)
		if err != nil {
			fmt.Errorf("failed to put metric '%s' for component '%s': %w",
				associatedMetric.MetricName, componentName, err)
			continue
		}

		if verbose {
			fmt.Printf("Successfully submitted metric '%s' with value '%s'\n", associatedMetric.MetricName, evaluatedMetricValue)
		}
	}

	fmt.Printf("Successfully processed %d metrics for component '%s'\n", len(component.AssociatedMetrics), componentName)
	return nil
}
