package services

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type MetricsParser struct {
	basePath string
}

func NewMetricsParser(basePath string) *MetricsParser {
	return &MetricsParser{basePath: basePath}
}

func (mp *MetricsParser) ParseMetrics() ([]MetricDefinition, error) {
	var metrics []MetricDefinition

	err := filepath.WalkDir(mp.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if !strings.HasSuffix(strings.ToLower(path), ".yaml") &&
			!strings.HasSuffix(strings.ToLower(path), ".yml") {
			return nil
		}

		fileMetrics, err := mp.parseYAMLFile(path)
		if err != nil {
			return nil // Continue on error
		}

		metrics = append(metrics, fileMetrics...)
		return nil
	})

	return metrics, err
}

func (mp *MetricsParser) parseYAMLFile(yamlFilePath string) ([]MetricDefinition, error) {
	content, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return nil, err
	}

	documents := strings.Split(string(content), "---")
	var metrics []MetricDefinition

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var metric MetricDefinition
		if err := yaml.Unmarshal([]byte(doc), &metric); err != nil {
			continue
		}

		if metric.Kind == "Metric" && metric.Metadata.Name != "" {
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}
