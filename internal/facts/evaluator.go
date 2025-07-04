package facts

import (
	"context"
	"fmt"
	"strings"

	"github.com/motain/compass-compute/internal/services"
)

type FactEvaluator struct {
	repoPath          string
	prometheusService *services.PrometheusService
}

func NewFactEvaluator(repoPath string) *FactEvaluator {

	client := services.NewPrometheusClient()
	prometheusService := services.NewPrometheusService(client)

	return &FactEvaluator{
		repoPath:          repoPath,
		prometheusService: prometheusService,
	}
}

func EvaluateMetric(facts []services.Fact, componentName string) (interface{}, error) {
	if len(facts) == 0 {
		return nil, fmt.Errorf("no facts provided")
	}

	evaluator := NewFactEvaluator("./repos")
	ctx := context.Background()

	factMap := make(map[string]*services.Fact)
	for i := range facts {
		factMap[facts[i].ID] = &facts[i]
	}

	for i := range facts {
		facts[i] = replacePlaceholders(facts[i], componentName)
	}

	for {
		progress := false
		for i := range facts {

			fact := &facts[i]
			if fact.Done {
				continue
			}

			if !areDependenciesSatisfied(fact, factMap) {
				continue
			}

			if err := evaluator.processFact(ctx, fact, factMap); err != nil {
				return nil, fmt.Errorf("failed to process fact %s: %w", fact.ID, err)
			}

			fact.Done = true
			progress = true
		}

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

	var finalResult interface{}
	for i := len(facts) - 1; i >= 0; i-- {
		if facts[i].Result != nil {
			finalResult = facts[i].Result
			break
		}
	}

	if convertedFloat := convertToFloat64(finalResult); convertedFloat != nil {
		return *convertedFloat, nil
	}

	return finalResult, nil
}

func (fe *FactEvaluator) processFact(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) error {
	if fact.Type == "" {
		return fmt.Errorf("fact type is empty for fact ID: %s", fact.ID)
	}

	switch strings.ToLower(fact.Type) {
	case "extract":
		return fe.processExtract(ctx, fact, factMap)
	case "validate":
		return fe.processValidate(fact, factMap)
	case "aggregate":
		return fe.processAggregate(fact, factMap)
	default:
		return fe.processCustom(ctx, fact, factMap)
	}
}

func (fe *FactEvaluator) processCustom(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) error {

	return fmt.Errorf("unknown fact type: %s", fact.Type)
}
