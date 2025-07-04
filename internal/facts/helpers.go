package facts

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/motain/compass-compute/internal/services"
	"github.com/pelletier/go-toml/v2"
)

func replacePlaceholders(fact services.Fact, componentName string) services.Fact {
	pattern := regexp.MustCompile(`\$\{Metadata\.Name\}`)

	fact.Repo = pattern.ReplaceAllString(fact.Repo, componentName)
	fact.FilePath = pattern.ReplaceAllString(fact.FilePath, componentName)
	fact.URI = pattern.ReplaceAllString(fact.URI, componentName)
	fact.PrometheusQuery = pattern.ReplaceAllString(fact.PrometheusQuery, componentName)

	if jsonPath, ok := fact.JSONPath.(string); ok {
		fact.JSONPath = pattern.ReplaceAllString(jsonPath, componentName)
	}

	return fact
}

func areDependenciesSatisfied(fact *services.Fact, factMap map[string]*services.Fact) bool {
	for _, depID := range fact.DependsOn {
		if dep, exists := factMap[depID]; !exists || !dep.Done {
			return false
		}
	}
	return true
}

func getDependencyResults(fact *services.Fact, factMap map[string]*services.Fact) []interface{} {
	var results []interface{}
	for _, depID := range fact.DependsOn {
		if dep, exists := factMap[depID]; exists && dep.Done {
			results = append(results, dep.Result)
		}
	}
	return results
}

func convertTOMLToJSON(tomlData []byte) ([]byte, error) {
	var config map[string]interface{}

	// Use a proper TOML parser that preserves structure
	if err := toml.Unmarshal(tomlData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Convert to JSON while preserving nested structure
	return json.Marshal(config)
}
