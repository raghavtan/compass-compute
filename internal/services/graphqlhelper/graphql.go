package graphqlhelper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ComponentByReference finds a component by name or ID and returns type, ID, and associated metrics
func ComponentByReference(name, token, cloudId, baseUrl string) (*ComponentReference, error) {
	component, err := searchComponentByName(name, token, cloudId, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("component not found: %w", err)
	}

	return component, nil
}

// searchComponentByName searches for a component by name and returns component with associated metrics
func searchComponentByName(name, token, cloudId, baseUrl string) (*ComponentReference, error) {
	query := `
		query getComponentBySlug($cloudId: ID!, $slug: String!) {
			compass {
				componentByReference(reference: {slug: {slug: $slug, cloudId: $cloudId}}) {
					... on CompassComponent {
						id
						name,
						type,
						metricSources {
							... on CompassComponentMetricSourcesConnection {
								nodes {
									id,
									metricDefinition {
										name,
										id
									}
								}
							}
						}
					}
				}
			}
		}`

	variables := map[string]interface{}{
		"cloudId": cloudId,
		"slug":    name,
	}

	response, err := executeGraphQLQuery(query, variables, token, cloudId, baseUrl)
	if err != nil {
		return nil, err
	}

	var componentResponse ComponentByReferenceResponse
	if err := json.Unmarshal(response, &componentResponse); err != nil {
		return nil, fmt.Errorf("failed to parse component response: %w", err)
	}

	component := componentResponse.Data.Compass.ComponentByReference

	// Check if component was found
	if component.ID == "" {
		return nil, fmt.Errorf("no component found with name: %s", name)
	}

	// Extract associated metrics from metricSources
	var associatedMetrics []AssociatedMetric
	for _, metricSource := range component.MetricSources.Nodes {
		if !strings.Contains(metricSource.MetricDefinition.ID, "builtin") {
			associatedMetrics = append(associatedMetrics, AssociatedMetric{
				MetricName:         metricSource.MetricDefinition.Name,
				MetricDefinitionID: metricSource.MetricDefinition.ID,
				MetricSourceID:     metricSource.ID,
			})
		}
	}

	return &ComponentReference{
		ComponentType:     component.Type,
		ComponentID:       component.ID,
		AssociatedMetrics: associatedMetrics,
	}, nil
}

// executeGraphQLQuery executes a GraphQL query against the Compass API
func executeGraphQLQuery(query string, variables map[string]interface{}, token, cloudId, baseUrl string) ([]byte, error) {
	if cloudId == "" {
		return nil, fmt.Errorf("cloudId is required")
	}

	if token == "" {
		return nil, fmt.Errorf("API token is required")
	}

	// Build GraphQL request
	graphqlReq := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(graphqlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	req, err := http.NewRequest("POST", baseUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", token))
	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, responseBody.String())
	}

	// Check for GraphQL errors
	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(responseBody.Bytes(), &graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphqlResp.Errors[0].Message)
	}

	return responseBody.Bytes(), nil
}
