package services

var query = `
		query getComponent($cloudId: ID!, $slug: String!) {
			compass {
				componentByReference(reference: {slug: {slug: $slug, cloudId: $cloudId}}) {
					... on CompassComponent {
						id name type
						metricSources {
							... on CompassComponentMetricSourcesConnection {
								nodes {
									id
									metricDefinition { name id }
								}
							}
						}
					}
				}
			}
		}`

var response struct {
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
