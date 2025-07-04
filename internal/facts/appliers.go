package facts

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/motain/compass-compute/internal/services"
)

func (fe *FactEvaluator) applyRule(fact *services.Fact, data []byte) (interface{}, error) {
	switch strings.ToLower(fact.Rule) {
	case "jsonpath":
		return fe.applyJSONPath(fact.JSONPath, data)
	case "notempty":
		return len(data) > 0, nil
	case "search":
		var result bool
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal search result: %w", err)
		}
		return result, nil
	default:
		return fe.applyCustomRule(fact, data)
	}
}

func (fe *FactEvaluator) applyCustomRule(fact *services.Fact, data []byte) (interface{}, error) {
	return string(data), nil
}

func (fe *FactEvaluator) applyJSONPath(jsonPath interface{}, data []byte) (interface{}, error) {
	if len(data) == 0 {
		return []interface{}{}, nil // Return empty array instead of nil
	}

	var jsonPathStr string
	switch v := jsonPath.(type) {
	case string:
		jsonPathStr = v
	case nil:
		return nil, fmt.Errorf("jsonPath is required but not provided")
	default:
		return nil, fmt.Errorf("jsonPath must be a string, got %T", jsonPath)
	}

	if jsonPathStr == "" {
		return nil, fmt.Errorf("jsonPath cannot be empty")
	}

	query, err := gojq.Parse(jsonPathStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse jsonPath '%s': %w", jsonPathStr, err)
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}
	if jsonData == nil {
		return []interface{}{}, nil
	}

	iter := query.Run(jsonData)
	results := make([]interface{}, 0)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("jsonPath query error: %w", err)
		}
		results = append(results, v)
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}
