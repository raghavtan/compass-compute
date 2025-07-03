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

func (fe *FactEvaluator) extractFromSource(ctx context.Context, fact *services.Fact) ([]byte, error) {

	switch strings.ToLower(fact.Source) {

	case "github":
		return fe.extractFromGitHub(fact)
	case "jsonapi", "api":
		return fe.extractFromAPI(ctx, fact)
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
		return fe.searchInRepo(fact.Repo, fact.SearchString)
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

func (fe *FactEvaluator) extractFromAPI(ctx context.Context, fact *services.Fact) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

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
	return json.Marshal(0.0)
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
