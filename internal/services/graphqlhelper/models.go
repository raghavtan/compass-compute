package graphqlhelper

// ComponentReference represents a component lookup result
type ComponentReference struct {
	ComponentType     string             `json:"componentType"`
	ComponentID       string             `json:"componentId"`
	AssociatedMetrics []AssociatedMetric `json:"associatedMetrics"`
}

// AssociatedMetric represents a metric associated with a component
type AssociatedMetric struct {
	MetricName         string `json:"metricName"`
	MetricDefinitionID string `json:"metricDefinitionId"`
	MetricSourceID     string `json:"metricSourceId"`
}

// MetricDefinition represents a metric definition
type MetricDefinition struct {
	Name               string `json:"name"`
	MetricDefinitionID string `json:"metricDefinitionId"`
}

// GraphQL request/response structures
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Updated response structure to match the actual API response
type ComponentByReferenceResponse struct {
	Data struct {
		Compass struct {
			ComponentByReference struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				Type          string `json:"type"`
				MetricSources struct {
					Nodes []struct {
						ID               string `json:"id"`
						MetricDefinition struct {
							Name string `json:"name"`
							ID   string `json:"id"`
						} `json:"metricDefinition"`
					} `json:"nodes"`
				} `json:"metricSources"`
			} `json:"componentByReference"`
		} `json:"compass"`
	} `json:"data"`
}
