package services

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PrometheusClientInterface interface {
	Query(query string, timestamp time.Time) (float64, error)
	QueryRange(query string, r v1.Range) (model.Value, error)
}

type PrometheusClient struct {
	api    v1.API // Underlying Prometheus API client
	region string // AWS region for the Prometheus workspace
}

func NewPrometheusClient() PrometheusClientInterface {
	ctx := context.Background()
	region := os.Getenv("AWS_REGION")
	workspaceURL := os.Getenv("PROMETHEUS_WORKSPACE_URL")
	awsRole := os.Getenv("AWS_ROLE")

	if workspaceURL == "" {
		panic("Prometheus workspace URL not configured")
	}

	// Initialize AWS configuration
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		panic(fmt.Errorf("failed to load AWS config: %w", err))
	}

	// Set up credentials provider
	credProvider := getCredentialsProvider(ctx, awsCfg, awsRole)

	// Create authenticated HTTP client
	httpClient := &http.Client{
		Transport: &SigV4RoundTripper{
			Transport:   http.DefaultTransport,
			Region:      region,
			Service:     "aps",
			Credentials: credProvider,
		},
	}

	// Initialize Prometheus client
	promClient, err := api.NewClient(api.Config{
		Address: workspaceURL,
		Client:  httpClient,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create Prometheus client: %w", err))
	}

	return &PrometheusClient{
		api:    v1.NewAPI(promClient),
		region: region,
	}
}

func getCredentialsProvider(ctx context.Context, awsCfg aws.Config, roleARN string) aws.CredentialsProvider {
	if roleARN == "" {
		stsClient := sts.NewFromConfig(awsCfg)
		_, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			panic(fmt.Errorf("failed to get caller identity: %w", err))
		}
		return awsCfg.Credentials
	}
	return stscreds.NewAssumeRoleProvider(sts.NewFromConfig(awsCfg), roleARN)
}

func (pc *PrometheusClient) Query(query string, timestamp time.Time) (float64, error) {
	result, _, err := pc.api.Query(context.Background(), query, timestamp)
	if err != nil {
		fmt.Printf("Query: %s, Timestamp: %s\n", query, timestamp)
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	response := 0.0
	if vector, ok := result.(model.Vector); ok {
		for _, sample := range vector {
			response = float64(sample.Value)
		}
	} else {
		fmt.Println("Result is not a vector for query:", query)
	}

	return response, nil
}

func (pc *PrometheusClient) QueryRange(query string, r v1.Range) (model.Value, error) {
	result, _, err := pc.api.QueryRange(context.Background(), query, r)
	if err != nil {
		return nil, fmt.Errorf("failed to execute range query: %w", err)
	}
	return result, nil
}

type PrometheusServiceInterface interface {
	InstantQuery(queryString string) (float64, error)
	RangeQuery(queryString string, start, end time.Time, step time.Duration) (model.Value, error)
}

type PrometheusService struct {
	client PrometheusClientInterface
}

func NewPrometheusService(client PrometheusClientInterface) *PrometheusService {
	return &PrometheusService{client: client}
}

func (ps *PrometheusService) InstantQuery(queryString string) (float64, error) {
	return ps.client.Query(queryString, time.Now())
}
func (ps *PrometheusService) RangeQuery(queryString string, start, end time.Time, step time.Duration) (model.Value, error) {
	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}
	return ps.client.QueryRange(queryString, r)
}
