package graphqlhelper

// ComponentReference represents a component lookup result
type ComponentReference struct {
	ComponentType string `json:"componentType"`
	ComponentID   string `json:"componentId"`
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

type ComponentsSearchResponse struct {
	Data struct {
		Compass struct {
			ComponentByReference struct {
				Nodes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
				} `json:"nodes"`
			} `json:"componentByReference"`
		} `json:"compass"`
	} `json:"data"`
}
