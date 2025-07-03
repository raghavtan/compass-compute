package facts

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/motain/compass-compute/internal/services"
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
	var result map[string]interface{}

	lines := strings.Split(string(tomlData), "\n")
	result = make(map[string]interface{})

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				value = strings.Trim(value, `"`)

				if num, err := strconv.ParseFloat(value, 64); err == nil {
					result[key] = num
				} else if value == "true" || value == "false" {
					result[key] = value == "true"
				} else {
					result[key] = value
				}
			}
		}
	}

	return json.Marshal(result)
}
