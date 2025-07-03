package services

type Component struct {
	Name    string   `json:"name"`
	ID      string   `json:"id"`
	Type    string   `json:"type"`
	Metrics []Metric `json:"metrics"`
}

type Metric struct {
	Name         string `json:"name"`
	DefinitionID string `json:"definitionId"`
	SourceID     string `json:"sourceId"`
}

type Fact struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Type            string      `json:"type"`
	Source          string      `json:"source,omitempty"`
	Repo            string      `json:"repo,omitempty"`
	FilePath        string      `json:"filePath,omitempty"`
	JSONPath        interface{} `json:"jsonPath,omitempty"`
	Rule            string      `json:"rule,omitempty"`
	Auth            interface{} `json:"auth,omitempty"`
	DependsOn       []string    `json:"dependsOn,omitempty"`
	Method          string      `json:"method,omitempty"`
	URI             string      `json:"uri,omitempty"`
	Pattern         string      `json:"pattern,omitempty"`
	SearchString    string      `json:"searchString,omitempty"`
	PrometheusQuery string      `json:"prometheusQuery,omitempty"`

	// Runtime fields
	Result interface{} `json:"-"`
	Done   bool        `json:"-"`
}

type MetricDefinition struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name          string            `yaml:"name"`
		Labels        map[string]string `yaml:"labels"`
		ComponentType []string          `yaml:"componentType"`
		Facts         []Fact            `yaml:"facts,omitempty"`
	} `yaml:"metadata"`
	Spec struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Format      struct {
			Unit string `yaml:"unit"`
		} `yaml:"format"`
	} `yaml:"spec"`
}
