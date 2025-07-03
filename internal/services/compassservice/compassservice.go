package compassservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/motain/compass-compute/internal/services"
	"github.com/motain/compass-compute/internal/services/graphqlhelper"
	"github.com/motain/compass-compute/internal/services/metrics"
)

type Service struct {
	token           string // API token for authentication
	cloudId         string // Compass Cloud ID for the current instance
	baseUrl         string // Base URL for Compass API endpoints
	metricsEndpoint string // Endpoint for metrics operations
}

func NewCompassService() *Service {
	return &Service{
		baseUrl:         services.DefaultCompassAPIBaseURL,
		token:           os.Getenv("COMPASS_API_TOKEN"),
		cloudId:         os.Getenv("COMPASS_CLOUD_ID"),
		metricsEndpoint: services.DefaultCompassMetricsEndpoint,
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
	MetricSourceID     string // Unique identifier for the metric source
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

	// Convert graphqlhelper.AssociatedMetric to compassservice.AssociatedMetric
	var associatedMetrics []AssociatedMetric
	for _, metric := range componentReference.AssociatedMetrics {
		associatedMetrics = append(associatedMetrics, AssociatedMetric{
			MetricName:         metric.MetricName,
			MetricDefinitionID: metric.MetricDefinitionID,
			MetricSourceID:     metric.MetricSourceID,
		})
	}

	return &Component{
		ComponentName:     name,
		ComponentID:       componentReference.ComponentID,
		ComponentType:     componentReference.ComponentType,
		AssociatedMetrics: associatedMetrics,
	}, nil
}

func (cs *Service) PutMetric(componentID string, metricDefinitionID string, metricValue string) error {
	if componentID == "" || metricDefinitionID == "" || metricValue == "" {
		return errors.New("componentID, metricDefinitionID, and metricValue cannot be empty")
	}

	requestBody := map[string]string{
		"metricDefinitionId": metricDefinitionID,
		"value":              metricValue,
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
		"componentId":        componentID,
	}

	_, err := cs.makeAPIRequest("PUT", cs.metricsEndpoint, requestBody)
	return err
}

// GetMetricFactsByName returns the metric facts for a given metric name and component type
func (cs *Service) GetMetricFactsByName(metricName string, componentType string) (*[]metrics.Fact, error) {
	metricStore, err := metrics.NewMetricsParser(services.DefaultMetricLocalPath).ParseMetrics()
	if err != nil {
		log.Printf("Error parsing metrics: %v", err)
		return nil, err
	}

	for _, metric := range metricStore {
		if metric.Metadata.Name == metricName {
			for _, ct := range metric.Metadata.ComponentType {
				if strings.EqualFold(ct, componentType) {
					return &metric.Metadata.Facts, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no metric facts found for metric '%s' and component type '%s'", metricName, componentType)
}

// makeAPIRequest is a private helper method for making HTTP requests to Compass API
func (cs *Service) makeAPIRequest(method string, endpoint string, payload interface{}) ([]byte, error) {
	// Prepare request body
	var requestBody []byte
	var err error

	if payload != nil {
		requestBody, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+cs.token)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}(resp.Body)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
