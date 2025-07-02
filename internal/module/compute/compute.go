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

	component, err := compass.GetComponent(componentName)
	if err != nil {
		return fmt.Errorf("failed to get component ID for '%s': %w", componentName, err)
	}

	// Clone repositories
	cloner := github.NewGitHubCloner(os.Getenv("GITHUB_TOKEN"))
	repos := []string{componentName, services.DefaultCatalogRepo}

	for _, repo := range repos {
		err := cloner.Clone("motain", repo, &github.CloneOptions{Destination: "./repos/"})
		if err != nil {
			return fmt.Errorf("failed to clone repository '%s': %w", repo, err)
		}
	}

	associatedMetrics, err := compass.GetAssociatedMetrics(component.ComponentType)
	if err != nil {
		return fmt.Errorf("failed to get associated metrics for component '%s': %w", componentName, err)
	}
	fmt.Printf("Associated metrics for component '%s': %v\n", componentName, associatedMetrics)

	// Process each metric
	//for _, metric := range component.AssociatedMetrics {
	//	evaluatedMetricValue, err := facts.EvaluateMetric(associatedMetrics[metric.MetricName], component.ComponentName)
	//	if err != nil {
	//		return fmt.Errorf("failed to evaluate metric '%s' for component '%s': %w", metric.MetricName, componentName, err)
	//	}
	//
	//	err = compass.PutMetric(component.ComponentID, metric.MetricDefinitionID, evaluatedMetricValue)
	//	if err != nil {
	//		return fmt.Errorf("failed to put metric '%s' for component '%s': %w", metric.MetricName, componentName, err)
	//	}
	//}

	return nil
}
