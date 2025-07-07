package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type CompassService struct {
	token   string
	cloudID string
	client  *http.Client
}

func NewCompassService() *CompassService {
	return &CompassService{
		token:   os.Getenv("COMPASS_API_TOKEN"),
		cloudID: os.Getenv("COMPASS_CLOUD_ID"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (cs *CompassService) GetComponent(name string) (*Component, error) {

	variables := map[string]interface{}{
		"cloudId": cs.cloudID,
		"slug":    ServiceSlugPrefix + name,
	}

	respData, err := cs.graphqlRequest(getComponentQuery, variables)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(respData, &getComponentResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	comp := getComponentResponse.Data.Compass.ComponentByReference
	if comp.ID == "" {
		return nil, fmt.Errorf("component not found: %s", name)
	}

	var metrics []Metric
	for _, node := range comp.MetricSources.Nodes {
		if !strings.Contains(node.MetricDefinition.ID, "builtin") {
			metrics = append(metrics, Metric{
				Name:         node.MetricDefinition.Name,
				DefinitionID: node.MetricDefinition.ID,
				SourceID:     node.ID,
			})
		}
	}

	return &Component{
		Name:    name,
		ID:      comp.ID,
		Type:    comp.Type,
		Metrics: metrics,
	}, nil
}

func (cs *CompassService) PutMetric(componentID, metricDefinitionID, value string) error {
	payload := map[string]string{
		"metricDefinitionId": metricDefinitionID,
		"value":              value,
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
		"componentId":        componentID,
	}

	_, err := cs.httpRequest("POST", MetricsEndpoint, payload)
	return err
}

func (cs *CompassService) GetMetricFacts(metricName, componentType string) ([]Fact, error) {
	parser := NewMetricsParser(GetMetricLocalPath())
	metrics, err := parser.ParseMetrics()
	if err != nil {
		return nil, err
	}

	for _, metric := range metrics {
		if metric.Metadata.Name == metricName {
			for _, ct := range metric.Metadata.ComponentType {
				if strings.EqualFold(ct, componentType) {
					return metric.Metadata.Facts, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no facts found for metric '%s' and type '%s'", metricName, componentType)
}

func (cs *CompassService) graphqlRequest(query string, variables map[string]interface{}) ([]byte, error) {
	reqBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	respData, err := cs.httpRequest("POST", GraphQLEndpoint, reqBody)
	if err != nil {
		return nil, err
	}

	var response struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respData, &response); err == nil && len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	return respData, nil
}

func (cs *CompassService) httpRequest(method, url string, payload interface{}) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+cs.token)

	resp, err := cs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}(resp.Body)

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respData))
	}

	return respData, nil
}

func (cs *CompassService) GetAllComponentList() ([]string, error) {

	variables := map[string]interface{}{
		"cloudId": cs.cloudID,
		"query": map[string]interface{}{
			"first": 200,
			"fieldFilters": map[string]interface{}{
				"name": "state",
				"filter": map[string]interface{}{
					"neq": "PENDING",
				},
			},
		},
	}

	respData, err := cs.graphqlRequest(getAllComponentQuery, variables)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(respData, &getAllComponentResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	comp := getAllComponentResponse.Data.Compass.SearchComponents

	var componentList []string
	for _, node := range comp.Nodes {
		componentList = append(componentList, node.Component.Name)
	}

	return componentList, nil
}
