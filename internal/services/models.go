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
	ID              string      `json:"id" yaml:"id"`
	Name            string      `json:"name ,omitempty" yaml:"name,omitempty"`
	Type            string      `json:"type ,omitempty" yaml:"type,omitempty"`
	Source          string      `json:"source,omitempty" yaml:"source,omitempty"`
	Repo            string      `json:"repo,omitempty" yaml:"repo,omitempty"`
	FilePath        string      `json:"filePath,omitempty" yaml:"filePath,omitempty"`
	JSONPath        interface{} `json:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`
	Rule            string      `json:"rule,omitempty" yaml:"rule,omitempty"`
	Auth            interface{} `json:"auth,omitempty" yaml:"auth,omitempty"`
	DependsOn       []string    `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	Method          string      `json:"method,omitempty" yaml:"method,omitempty"`
	URI             string      `json:"uri,omitempty" yaml:"uri,omitempty"`
	Pattern         string      `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	SearchString    string      `json:"searchString,omitempty" yaml:"searchString,omitempty"`
	PrometheusQuery string      `json:"prometheusQuery,omitempty" yaml:"prometheusQuery,omitempty"`

	// Runtime fields
	Result interface{} `json:"-"`
	Done   bool        `json:"-"`
}

type MetricDefinition struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind" json:"kind"`
	Metadata   struct {
		Name          string            `yaml:"name" json:"name,omitempty"`
		Labels        map[string]string `yaml:"labels" json:"labels,omitempty"`
		ComponentType []string          `yaml:"componentType" json:"componentType,omitempty"`
		Facts         []Fact            `yaml:"facts,omitempty" json:"facts,omitempty"`
	} `yaml:"metadata"`
	Spec struct {
		Name        string `yaml:"name" json:"name,omitempty"`
		Description string `yaml:"description" json:"description,omitempty"`
		Format      struct {
			Unit string `yaml:"unit" json:"unit,omitempty"`
		} `yaml:"format" json:"format,omitempty"`
	} `yaml:"spec" json:"spec,omitempty"`
}
