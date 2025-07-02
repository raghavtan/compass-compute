package metrics

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Metric represents the structure of a metric definition
type Metric struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Facts      []Fact   `yaml:"facts,omitempty"`
	Spec       Spec     `yaml:"spec"`
}

// Metadata contains metric metadata
type Metadata struct {
	Name          string            `yaml:"name"`
	Labels        map[string]string `yaml:"labels"`
	ComponentType []string          `yaml:"componentType"`
}

// Fact represents individual facts/checks within a metric
type Fact struct {
	ID              string      `yaml:"id"`
	Name            string      `yaml:"name"`
	Type            string      `yaml:"type"`
	Source          string      `yaml:"source,omitempty"`
	Repo            string      `yaml:"repo,omitempty"`
	FilePath        string      `yaml:"filePath,omitempty"`
	JSONPath        interface{} `yaml:"jsonPath,omitempty"`
	Rule            string      `yaml:"rule,omitempty"`
	Auth            interface{} `yaml:"auth,omitempty"`
	DependsOn       []string    `yaml:"dependsOn,omitempty"`
	Method          string      `yaml:"method,omitempty"`
	URI             string      `yaml:"uri,omitempty"`
	Pattern         string      `yaml:"pattern,omitempty"`
	SearchString    string      `yaml:"searchString,omitempty"`
	PrometheusQuery string      `yaml:"prometheusQuery,omitempty"`
}

// Spec contains metric specification
type Spec struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Format      Format `yaml:"format"`
}

// Format defines the output format for the metric
type Format struct {
	Unit string `yaml:"unit"`
}

// MetricsParser handles parsing of metric YAML files
type MetricsParser struct {
	basePath string
}

// NewMetricsParser creates a new parser instance
func NewMetricsParser(basePath string) *MetricsParser {
	return &MetricsParser{
		basePath: basePath,
	}
}

// ParseMetrics reads all YAML files from the specified path and returns an array of metrics
func (mp *MetricsParser) ParseMetrics() ([]Metric, error) {
	var metrics []Metric

	err := filepath.WalkDir(mp.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-YAML files
		if d.IsDir() || (!strings.HasSuffix(strings.ToLower(path), ".yaml") && !strings.HasSuffix(strings.ToLower(path), ".yml")) {
			return nil
		}

		// Read and parse the YAML file
		fileMetrics, err := mp.parseYAMLFile(path)
		if err != nil {
			log.Printf("Error parsing file %s: %v", path, err)
			return nil // Continue with other files even if one fails
		}

		metrics = append(metrics, fileMetrics...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return metrics, nil
}

// parseYAMLFile reads and parses a single YAML file that may contain multiple documents
func (mp *MetricsParser) parseYAMLFile(filePath string) ([]Metric, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	// Split YAML documents (separated by ---)
	documents := strings.Split(string(content), "---")
	var metrics []Metric

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var metric Metric
		if err := yaml.Unmarshal([]byte(doc), &metric); err != nil {
			log.Printf("Warning: Failed to parse YAML document in %s: %v", filePath, err)
			continue
		}

		// Skip empty metrics or non-Metric kinds
		if metric.Kind != "Metric" || metric.Metadata.Name == "" {
			continue
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}
