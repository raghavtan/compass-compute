package facts

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/motain/compass-compute/internal/services"
)

func (fe *FactEvaluator) processExtract(ctx context.Context, fact *services.Fact) error {
	if fact.Source == "" {
		return fmt.Errorf("source is required for extract fact type")
	}

	data, err := fe.extractFromSource(ctx, fact)
	if err != nil {
		return fmt.Errorf("extraction failed from source '%s': %w", fact.Source, err)
	}

	if fact.Rule != "" {
		result, err := fe.applyRule(fact, data)
		if err != nil {
			return fmt.Errorf("rule application failed for rule '%s': %w", fact.Rule, err)
		}
		fact.Result = result
	} else {
		if data == nil {
			fact.Result = nil
		} else {
			fact.Result = string(data)
		}
	}

	return nil
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

// Validation methods

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
