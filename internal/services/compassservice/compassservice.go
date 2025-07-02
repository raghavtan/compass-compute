package compassservice

import (
	"errors"
	"fmt"
	"github.com/motain/compass-compute/internal/services"
	"github.com/motain/compass-compute/internal/services/graphqlhelper"
	"github.com/motain/compass-compute/internal/services/metrics"
	"log"
	"os"
)

type Service struct {
	token   string // API token for authentication
	cloudId string // Compass Cloud ID for the current instance
	baseUrl string // Base URL for Compass API endpoints
}

func NewCompassService() *Service {
	return &Service{
		baseUrl: os.Getenv("COMPASS_API_BASE_URL"),
		token:   os.Getenv("COMPASS_API_TOKEN"),
		cloudId: os.Getenv("COMPASS_CLOUD_ID"),
	}
}

type Component struct {
	ComponentName     string // Name of the component
	ComponentID       string // Unique identifier for the component
	ComponentType     string // Type of the component (e.g., service, application)
	AssociatedMetrics []AssociatedMetric
}

type AssociatedMetric struct {
	MetricName         string // Name of the metric
	MetricDefinitionID string // Unique identifier for the metric definition
}

func (cs *Service) GetComponent(name string) (*Component, error) {

	if name == "" {
		return nil, errors.New("component name cannot be empty")
	}
	componentReference, err := graphqlhelper.ComponentByReference("svc-"+name, cs.token, cs.cloudId, cs.baseUrl)
	if err != nil {
		log.Printf("Error retrieving component by reference: %v", err)
		return nil, fmt.Errorf("failed to get component '%s': %w", name, err)
	}

	return &Component{
		ComponentName: name,
		ComponentID:   componentReference.ComponentID,
		ComponentType: componentReference.ComponentType,
	}, nil
}

func (cs *Service) PutMetric(componentID string, metricDefinitionID string, metricValue string) error {
	// TODO: Implement the actual API call to associate a metric with a component
	// This method should:
	// 1. Validate the input parameters (componentID and metricDefinitionID)
	// 2. Prepare the request payload for the PUT operation
	//    Example payload: {"metricDefinitionId": metricDefinitionID}
	// 3. Make an HTTP PUT request to the Compass API
	//    Example endpoint: PUT /compass/v1/components/{componentID}/metrics
	// 4. Handle the response and any errors
	// 5. Return nil on success or an appropriate error on failure

	if componentID == "" {
		return errors.New("component ID cannot be empty")
	}

	if metricDefinitionID == "" {
		return errors.New("metric definition ID cannot be empty")
	}

	// TODO: Add additional validation if needed
	// if err := cs.validateIDs(componentID, metricDefinitionID); err != nil {
	//     return err
	// }

	// Placeholder implementation
	return fmt.Errorf("PutMetric not implemented yet - TODO: associate metric definition '%s' with component '%s'",
		metricDefinitionID, componentID)
}

// Private helper methods for the CompassService class

func (cs *Service) GetAssociatedMetrics(componentType string) (*[]metrics.Metric, error) {
	metricStore, err := metrics.NewMetricsParser(services.DefaultMetricLocalPath).ParseMetrics()
	if err != nil {
		log.Fatalf("Error parsing metrics: %v", err)
		return nil, err
	}
	var associatedMetrics *[]metrics.Metric

	for _, metric := range metricStore {
		for _, ct := range metric.Metadata.ComponentType {
			if ct == componentType {
				if associatedMetrics == nil {
					associatedMetrics = &[]metrics.Metric{}
				}
				*associatedMetrics = append(*associatedMetrics, metric)
				fmt.Printf("Found associated metric '%s' for component type '%s'\n", metric.Metadata.Name, componentType)
			}
		}
	}

	return associatedMetrics, nil
}

// makeAPIRequest is a private helper method for making HTTP requests to Compass API
func (cs *Service) makeAPIRequest(method, endpoint string, payload interface{}) ([]byte, error) {
	// TODO: Implement generic API request handling
	// This method should:
	// 1. Prepare the HTTP request with proper headers and authentication
	// 2. Handle request serialization (JSON encoding)
	// 3. Execute the HTTP request
	// 4. Handle response parsing and error checking
	// 5. Return response body or error
	return nil, errors.New("makeAPIRequest not implemented yet")
}
