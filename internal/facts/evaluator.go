// internal/facts/evaluator.go
package facts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/motain/compass-compute/internal/services"

	"github.com/itchyny/gojq"
)

// FactEvaluator handles the evaluation of facts
type FactEvaluator struct {
	httpClient *http.Client
	repoPath   string
}

// NewFactEvaluator creates a new fact evaluator
func NewFactEvaluator(repoPath string) *FactEvaluator {
	return &FactEvaluator{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		repoPath:   repoPath,
	}
}

// EvaluateMetric evaluates a list of facts and returns the final result
func EvaluateMetric(facts []services.Fact, componentName string) (interface{}, error) {
	if len(facts) == 0 {
		return nil, fmt.Errorf("no facts provided")
	}

	evaluator := NewFactEvaluator("./repos")
	ctx := context.Background()

	// Create a map for quick lookup
	factMap := make(map[string]*services.Fact)
	for i := range facts {
		factMap[facts[i].ID] = &facts[i]
	}

	// Replace placeholders with component name
	for i := range facts {
		facts[i] = replacePlaceholders(facts[i], componentName)
	}

	// Process facts in dependency order
	for {
		progress := false
		for i := range facts {
			fact := &facts[i]
			if fact.Done {
				continue
			}

			// Check if all dependencies are satisfied
			if !areDependenciesSatisfied(fact, factMap) {
				continue
			}

			// Process this fact
			if err := evaluator.processFact(ctx, fact, factMap); err != nil {
				return nil, fmt.Errorf("failed to process fact %s: %w", fact.ID, err)
			}

			fact.Done = true
			progress = true
		}

		// Check if all facts are done
		allDone := true
		for i := range facts {
			if !facts[i].Done {
				allDone = false
				break
			}
		}

		if allDone {
			break
		}

		if !progress {
			return nil, fmt.Errorf("circular dependency or unresolved dependencies detected")
		}
	}

	// Return the result of the last fact or aggregate result
	var finalResult interface{}
	for i := len(facts) - 1; i >= 0; i-- {
		if facts[i].Result != nil {
			finalResult = facts[i].Result
			break
		}
	}

	// Convert to float64 if possible
	return convertToFloat64(finalResult), nil
}

func (fe *FactEvaluator) processFact(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) error {
	switch strings.ToLower(fact.Type) {
	case "extract":
		return fe.processExtract(ctx, fact, factMap)
	case "validate":
		return fe.processValidate(fact, factMap)
	case "aggregate":
		return fe.processAggregate(fact, factMap)
	default:
		return fmt.Errorf("unknown fact type: %s", fact.Type)
	}
}

func (fe *FactEvaluator) processExtract(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) error {
	var data []byte
	var err error

	switch strings.ToLower(fact.Source) {
	case "github":
		data, err = fe.extractFromGitHub(fact)
	case "jsonapi":
		data, err = fe.extractFromAPI(ctx, fact)
	case "prometheus":
		data, err = fe.extractFromPrometheus(fact)
	default:
		return fmt.Errorf("unsupported source: %s", fact.Source)
	}

	if err != nil {
		return err
	}

	// Apply rule-based processing
	result, err := fe.applyRule(fact, data)
	if err != nil {
		return err
	}

	fact.Result = result
	return nil
}

func (fe *FactEvaluator) extractFromGitHub(fact *services.Fact) ([]byte, error) {
	if fact.Rule == "search" {
		// Search for string in repository files
		return fe.searchInRepo(fact.Repo, fact.SearchString)
	}

	// Read file content
	repoPath := filepath.Join(fe.repoPath, fact.Repo)
	filePath := filepath.Join(repoPath, fact.FilePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist
		}
		return nil, err
	}

	// Convert TOML to JSON if needed
	if strings.HasSuffix(fact.FilePath, ".toml") {
		return fe.convertTOMLToJSON(data)
	}

	return data, nil
}

func (fe *FactEvaluator) extractFromAPI(ctx context.Context, fact *services.Fact) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fact.URI, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication if provided
	if fact.Auth != nil {
		if authMap, ok := fact.Auth.(map[string]interface{}); ok {
			if header, exists := authMap["header"].(string); exists {
				if tokenVar, exists := authMap["tokenVar"].(string); exists {
					token := os.Getenv(tokenVar)
					req.Header.Set(header, token)
				}
			}
		}
	}

	resp, err := fe.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	return io.ReadAll(resp.Body)
}

func (fe *FactEvaluator) extractFromPrometheus(fact *services.Fact) ([]byte, error) {
	// This would integrate with your existing Prometheus service
	// For now, return a placeholder
	return json.Marshal(0.0)
}

func (fe *FactEvaluator) applyRule(fact *services.Fact, data []byte) (interface{}, error) {
	switch strings.ToLower(fact.Rule) {
	case "jsonpath":
		return fe.applyJSONPath(fact.JSONPath, data)
	case "notempty":
		return len(data) > 0, nil
	case "search":
		return fact.Result, nil // Already processed in extractFromGitHub
	default:
		return string(data), nil
	}
}

func (fe *FactEvaluator) applyJSONPath(jsonPath interface{}, data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	jsonPathStr, ok := jsonPath.(string)
	if !ok {
		return nil, fmt.Errorf("jsonPath must be a string")
	}

	query, err := gojq.Parse(jsonPathStr)
	if err != nil {
		return nil, err
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	iter := query.Run(jsonData)
	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, err
		}
		results = append(results, v)
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

func (fe *FactEvaluator) processValidate(fact *services.Fact, factMap map[string]*services.Fact) error {
	deps := getDependencyResults(fact, factMap)
	if len(deps) == 0 {
		return fmt.Errorf("validation requires dependencies")
	}

	switch strings.ToLower(fact.Rule) {
	case "regex_match":
		return fe.validateRegex(fact, deps)
	case "deps_match":
		return fe.validateDepsMatch(fact, deps)
	case "unique":
		return fe.validateUnique(fact, deps)
	default:
		return fmt.Errorf("unknown validation rule: %s", fact.Rule)
	}
}

func (fe *FactEvaluator) validateRegex(fact *services.Fact, deps []interface{}) error {
	regex, err := regexp.Compile(fact.Pattern)
	if err != nil {
		return err
	}

	if len(deps) == 1 {
		value := fmt.Sprintf("%v", deps[0])
		fact.Result = regex.MatchString(value)
		return nil
	}

	// Validate each item in list
	results := make([]bool, len(deps))
	for i, dep := range deps {
		value := fmt.Sprintf("%v", dep)
		results[i] = regex.MatchString(value)
	}
	fact.Result = results
	return nil
}

func (fe *FactEvaluator) validateDepsMatch(fact *services.Fact, deps []interface{}) error {
	if len(deps) < 2 {
		fact.Result = true
		return nil
	}

	first := deps[0]
	for _, dep := range deps[1:] {
		if fmt.Sprintf("%v", first) != fmt.Sprintf("%v", dep) {
			fact.Result = false
			return nil
		}
	}
	fact.Result = true
	return nil
}

func (fe *FactEvaluator) validateUnique(fact *services.Fact, deps []interface{}) error {
	seen := make(map[string]bool)
	for _, dep := range deps {
		key := fmt.Sprintf("%v", dep)
		if seen[key] {
			fact.Result = false
			return nil
		}
		seen[key] = true
	}
	fact.Result = true
	return nil
}

func (fe *FactEvaluator) processAggregate(fact *services.Fact, factMap map[string]*services.Fact) error {
	deps := getDependencyResults(fact, factMap)
	if len(deps) == 0 {
		return fmt.Errorf("aggregation requires dependencies")
	}

	switch strings.ToLower(fact.Method) {
	case "count":
		fact.Result = float64(len(deps))
	case "sum":
		sum := 0.0
		for _, dep := range deps {
			if val := convertToFloat64(dep); val != nil {
				sum += *val
			}
		}
		fact.Result = sum
	case "and":
		result := true
		for _, dep := range deps {
			if val := convertToBool(dep); val != nil {
				result = result && *val
			}
		}
		fact.Result = result
	case "or":
		result := false
		for _, dep := range deps {
			if val := convertToBool(dep); val != nil {
				result = result || *val
			}
		}
		fact.Result = result
	default:
		return fmt.Errorf("unknown aggregation method: %s", fact.Method)
	}

	return nil
}

// Helper functions

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

func (fe *FactEvaluator) searchInRepo(repo, searchString string) ([]byte, error) {
	repoPath := filepath.Join(fe.repoPath, repo)
	found := false

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Skip binary files and common non-text files
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".git" || ext == ".bin" || ext == ".exe" {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil // Skip files we can't read
		}

		if strings.Contains(string(data), searchString) {
			found = true
			return filepath.SkipDir // Stop searching once found
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return json.Marshal(found)
}

func (fe *FactEvaluator) convertTOMLToJSON(tomlData []byte) ([]byte, error) {
	// Simple TOML to JSON conversion
	// This is a basic implementation - you might want to use a proper TOML library
	var result map[string]interface{}

	// For simplicity, assuming key=value format
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

				// Remove quotes
				value = strings.Trim(value, `"`)

				// Try to parse as number
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

func convertToFloat64(value interface{}) *float64 {
	switch v := value.(type) {
	case float64:
		return &v
	case float32:
		f := float64(v)
		return &f
	case int:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return &f
		}
	case bool:
		if v {
			f := 1.0
			return &f
		} else {
			f := 0.0
			return &f
		}
	}
	return nil
}

func convertToBool(value interface{}) *bool {
	switch v := value.(type) {
	case bool:
		return &v
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return &b
		}
	case float64:
		b := v != 0
		return &b
	case int:
		b := v != 0
		return &b
	}
	return nil
}
