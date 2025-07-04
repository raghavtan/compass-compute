package facts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/motain/compass-compute/internal/services"
)

func (fe *FactEvaluator) extractFromSource(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) ([]byte, error) {

	switch strings.ToLower(fact.Source) {

	case "github":
		return fe.extractFromGitHub(fact)
	case "jsonapi", "api":
		return fe.extractFromAPI(ctx, fact, factMap)
	case "prometheus":
		return fe.extractFromPrometheus(fact)
	default:
		// Hook for custom extractors
		return fe.extractCustom(ctx, fact)
	}
}

func (fe *FactEvaluator) extractCustom(ctx context.Context, fact *services.Fact) ([]byte, error) {
	return nil, fmt.Errorf("unsupported source: %s", fact.Source)
}

func (fe *FactEvaluator) extractFromGitHub(fact *services.Fact) ([]byte, error) {

	if fact.Rule == "search" {
		search, err := fe.searchInRepo(fact.Repo, fact.SearchString)
		if err != nil {
			return nil, fmt.Errorf("failed to search in repository '%s': %w", fact.Repo, err)
		}
		fmt.Printf("Search result for '%s' in repository '%s': %s\n", fact.SearchString, fact.Rule, string(search))
		return search, nil
	}

	if fact.FilePath == "" {
		return nil, fmt.Errorf("filePath is required for GitHub source")
	}

	repoPath := filepath.Join(fe.repoPath, fact.Repo)
	filePath := filepath.Join(repoPath, fact.FilePath)

	if filePath == repoPath {
		return nil, fmt.Errorf("invalid file path: %s", fact.FilePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	if strings.HasSuffix(fact.FilePath, ".toml") {
		return convertTOMLToJSON(data)
	}

	return data, nil
}

func (fe *FactEvaluator) extractFromAPI(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	uri := fe.substituteDependencyValues(fact, factMap)

	// If URI is empty (no dependencies to process), return appropriate empty result
	if uri == "" {
		// For recipients endpoint, return object with empty recipients array
		if strings.Contains(fact.URI, "recipients") || strings.Contains(fact.URI, ":alert_id") {
			emptyResponse := map[string]interface{}{
				"recipients": []interface{}{},
			}
			return json.Marshal(emptyResponse)
		}
		// For other endpoints, return empty array
		return json.Marshal([]interface{}{})
	}

	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
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

	resp, err := client.Do(req)
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

	if fe.prometheusService == nil {
		return nil, fmt.Errorf("prometheus service not configured")
	}
	query := fact.PrometheusQuery
	if query == "" {
		return nil, fmt.Errorf("no query specified for prometheus source (use URI field)")
	}

	switch fact.Rule {
	case "range":
		start := time.Now().Add(-1 * time.Hour) // Default to 1 hour ago
		end := time.Now()
		step := 15 * time.Second // Default 15-second step
		result, err := fe.prometheusService.RangeQuery(query, start, end, step)
		if err != nil {
			return nil, fmt.Errorf("prometheus range query failed: %w", err)
		}
		return json.Marshal(result)
	case "instant", "":
		// Default to instant query
		result, err := fe.prometheusService.InstantQuery(query)
		if err != nil {
			return nil, fmt.Errorf("prometheus instant query failed: %w", err)
		}
		response := map[string]interface{}{
			"value":     result,
			"timestamp": time.Now().Unix(),
			"query":     query,
		}

		return json.Marshal(response["value"])

	default:
		return nil, fmt.Errorf("unsupported prometheus rule: %s (use 'instant' or 'range')", fact.Rule)
	}
}

func (fe *FactEvaluator) searchInRepo(repo, searchString string) ([]byte, error) {
	repoPath := filepath.Join(fe.repoPath, repo)
	found := false

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".git" || ext == ".bin" || ext == ".exe" {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		if strings.Contains(string(data), searchString) {
			found = true
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return json.Marshal(found)
}

func (fe *FactEvaluator) substituteDependencyValues(fact *services.Fact, factMap map[string]*services.Fact) string {
	uri := fact.URI
	deps := getDependencyResults(fact, factMap)

	// If no dependencies are defined, return original URI
	if len(fact.DependsOn) == 0 {
		return uri
	}

	// If dependencies are defined but no results available, skip API call
	if len(deps) == 0 {
		return ""
	}

	// Check each dependency result
	for _, dep := range deps {
		// Handle string results (SLO IDs, Alert IDs)
		if sloID, ok := dep.(string); ok {
			uri = strings.ReplaceAll(uri, ":slo_id", sloID)
			uri = strings.ReplaceAll(uri, ":alert_id", sloID)
			continue
		}

		// Handle array results
		if arr, ok := dep.([]interface{}); ok {
			// If array is empty, skip API call
			if len(arr) == 0 {
				return ""
			}
			// Use first element of array for substitution
			if firstItem, ok := arr[0].(string); ok {
				uri = strings.ReplaceAll(uri, ":slo_id", firstItem)
				uri = strings.ReplaceAll(uri, ":alert_id", firstItem)
				continue
			}
		}
	}

	return uri
}
