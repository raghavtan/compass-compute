# 🔌 Extensions Guide

> **Add custom data sources, fact types, and processing rules**

## Quick Start: Your First Extension

Let's add a Kubernetes integration in 10 minutes:

### 1. Add to Facts Evaluator

Edit `internal/facts/evaluator.go`, in the `processCustom()` function:

```go
func (fe *FactEvaluator) processCustom(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) error {
    switch strings.ToLower(fact.Type) {
    case "kubernetes", "k8s":
        return fe.processKubernetes(ctx, fact)
    default:
        return fmt.Errorf("unknown fact type: %s", fact.Type)
    }
}
```

### 2. Implement the Processor

Add to `internal/facts/evaluator.go`:

```go
func (fe *FactEvaluator) processKubernetes(ctx context.Context, fact *services.Fact) error {
    namespace := fact.Repo           // Reuse field creatively
    resourceType := fact.FilePath    
    labelSelector := fact.SearchString
    operation := fact.Rule
    
    switch resourceType {
    case "pods":
        return fe.handleKubernetesPods(ctx, fact, namespace, labelSelector, operation)
    default:
        return fmt.Errorf("unsupported k8s resource: %s", resourceType)
    }
}

func (fe *FactEvaluator) handleKubernetesPods(ctx context.Context, fact *services.Fact, namespace, labelSelector, operation string) error {
    switch operation {
    case "count":
        count, err := fe.getPodCount(namespace, labelSelector)
        if err != nil {
            return err
        }
        fact.Result = float64(count)
        return nil
    default:
        return fmt.Errorf("unsupported pod operation: %s", operation)
    }
}

func (fe *FactEvaluator) getPodCount(namespace, labelSelector string) (int, error) {
    cmd := exec.Command("kubectl", "get", "pods", "-n", namespace)
    if labelSelector != "" {
        cmd.Args = append(cmd.Args, "-l", labelSelector)
    }
    cmd.Args = append(cmd.Args, "--no-headers")
    
    output, err := cmd.Output()
    if err != nil {
        return 0, fmt.Errorf("kubectl command failed: %w", err)
    }
    
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    if len(lines) == 1 && lines[0] == "" {
        return 0, nil
    }
    
    return len(lines), nil
}
```

### 3. Use Your Extension

```yaml
# pod-count-metric.yaml
apiVersion: v1
kind: Metric
metadata:
  name: pod-count
  componentType: ["service"]
  facts:
    - id: count-pods
      type: kubernetes
      repo: production        # namespace
      filePath: pods         # resource type
      searchString: app=myapp # label selector
      rule: count            # operation
```

That's it! Your Kubernetes extension is ready.

## Extension Points

### 1. Custom Fact Types

Add new fact types in `internal/facts/evaluator.go`:

```go
// In processCustom()
case "database":
    return fe.processDatabase(ctx, fact)
case "security":
    return fe.processSecurity(ctx, fact)
case "ai":
    return fe.processAI(ctx, fact)
```

### 2. Custom Data Sources

Add new data sources in `internal/facts/extractor.go`:

```go
// In extractCustom()
case "elasticsearch":
    return fe.extractFromElasticsearch(ctx, fact)
case "s3":
    return fe.extractFromS3(ctx, fact)
case "redis":
    return fe.extractFromRedis(ctx, fact)
```

### 3. Custom Processing Rules

Add new rules in `internal/facts/appliers.go`:

```go
// In applyCustomRule()
case "math":
    return fe.applyMath(fact, data)
case "nlp":
    return fe.applyNLP(fact, data)
case "crypto":
    return fe.applyCrypto(fact, data)
```

## Real Extension Examples

### Database Extension

```go
// In internal/facts/evaluator.go
func (fe *FactEvaluator) processDatabase(ctx context.Context, fact *services.Fact) error {
    connectionString := fact.URI
    query := fact.PrometheusQuery  // Reuse for SQL query
    dbType := fact.Pattern         // Database type
    
    switch strings.ToLower(dbType) {
    case "postgres":
        return fe.handlePostgreSQL(ctx, fact, connectionString, query)
    case "mysql":
        return fe.handleMySQL(ctx, fact, connectionString, query)
    default:
        return fmt.Errorf("unsupported database: %s", dbType)
    }
}

func (fe *FactEvaluator) handlePostgreSQL(ctx context.Context, fact *services.Fact, connectionString, query string) error {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    defer db.Close()
    
    var count int64
    err = db.QueryRowContext(ctx, query).Scan(&count)
    if err != nil {
        return fmt.Errorf("query failed: %w", err)
    }
    
    fact.Result = float64(count)
    return nil
}
```

**Usage:**
```yaml
- id: user-count
  type: database
  pattern: postgres
  uri: "postgres://user:pass@localhost/db"
  prometheusQuery: "SELECT COUNT(*) FROM users"
```

### Elasticsearch Extension

```go
// In internal/facts/extractor.go
func (fe *FactEvaluator) extractFromElasticsearch(ctx context.Context, fact *services.Fact) ([]byte, error) {
    esURL := fact.URI
    index := fact.Repo
    query := fact.PrometheusQuery
    
    client, err := elasticsearch.NewClient(elasticsearch.Config{
        Addresses: []string{esURL},
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create ES client: %w", err)
    }
    
    res, err := client.Search(
        client.Search.WithContext(ctx),
        client.Search.WithIndex(index),
        client.Search.WithBody(strings.NewReader(query)),
    )
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }
    defer res.Body.Close()
    
    return io.ReadAll(res.Body)
}
```

**Usage:**
```yaml
- id: log-errors
  type: extract
  source: elasticsearch
  uri: "http://localhost:9200"
  repo: "logs-2024"
  prometheusQuery: '{"query": {"term": {"level": "error"}}}'
  rule: jsonpath
  jsonPath: ".hits.total.value"
```

### Math Processing Rule

```go
// In internal/facts/appliers.go
func (fe *FactEvaluator) applyMath(fact *services.Fact, data []byte) (interface{}, error) {
    var numbers []float64
    if err := json.Unmarshal(data, &numbers); err != nil {
        return nil, fmt.Errorf("failed to parse numbers: %w", err)
    }
    
    operation := fact.Pattern
    
    switch operation {
    case "sum":
        sum := 0.0
        for _, num := range numbers {
            sum += num
        }
        return sum, nil
        
    case "average":
        if len(numbers) == 0 {
            return 0.0, nil
        }
        sum := 0.0
        for _, num := range numbers {
            sum += num
        }
        return sum / float64(len(numbers)), nil
        
    case "max":
        if len(numbers) == 0 {
            return 0.0, nil
        }
        max := numbers[0]
        for _, num := range numbers {
            if num > max {
                max = num
            }
        }
        return max, nil
        
    default:
        return nil, fmt.Errorf("unsupported math operation: %s", operation)
    }
}
```

**Usage:**
```yaml
- id: calculate-average
  type: extract
  source: api
  uri: "https://api.example.com/metrics"
  rule: math
  pattern: average
```

## Field Mapping Strategy

Since fact structs have limited fields, reuse them creatively:

| Field | Primary Use | Alternative Uses |
|-------|-------------|------------------|
| `Type` | Fact type | - |
| `Source` | Data source | Protocol type |
| `Repo` | Repository | Namespace, Database, Index |
| `FilePath` | File path | Resource type, Table name |
| `URI` | API endpoint | Connection string |
| `Rule` | Processing rule | Operation type |
| `Pattern` | Regex pattern | Sub-operation, Format |
| `SearchString` | Search term | Labels, Filters |
| `PrometheusQuery` | Prometheus query | SQL query, ES query |
| `JSONPath` | JSONPath expression | - |
| `Auth` | Authentication | Configuration object |

## Testing Extensions

### Unit Tests

```go
func TestKubernetesExtension(t *testing.T) {
    tests := []struct {
        name     string
        fact     *services.Fact
        expected float64
        wantErr  bool
    }{
        {
            name: "count pods",
            fact: &services.Fact{
                Type:         "kubernetes",
                Repo:         "default",
                FilePath:     "pods",
                Rule:         "count",
            },
            expected: 3,
            wantErr:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fe := &FactEvaluator{}
            err := fe.processKubernetes(context.Background(), tt.fact)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, tt.fact.Result)
            }
        })
    }
}
```

### Integration Tests

```bash
# Test with real service
./compass-compute compute test-service --verbose

# Test specific extension
METRIC_DIR=./test-metrics ./compass-compute compute test-service
```

## Best Practices

### Security
- Always validate inputs
- Use environment variables for secrets
- Sanitize data in logs
- Use secure HTTP clients with timeouts

### Performance
- Use connection pools for databases
- Implement caching for expensive operations
- Use context for cancellation
- Limit concurrent operations

### Error Handling
- Return descriptive errors
- Handle missing data gracefully
- Log errors with context
- Don't crash on single fact failures

### Example with Best Practices

```go
func (fe *FactEvaluator) processSecureAPI(ctx context.Context, fact *services.Fact) error {
    // Input validation
    if fact.URI == "" {
        return fmt.Errorf("URI is required for API fact")
    }
    
    // Secure HTTP client
    client := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
        },
    }
    
    // Create request with context
    req, err := http.NewRequestWithContext(ctx, "GET", fact.URI, nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    
    // Add authentication
    if token := os.Getenv("API_TOKEN"); token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }
    
    // Make request
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Handle response
    if resp.StatusCode >= 400 {
        return fmt.Errorf("API error: %d", resp.StatusCode)
    }
    
    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read response: %w", err)
    }
    
    fact.Result = string(data)
    return nil
}
```

## Need Help?

- Check [Debugging Guide](debugging.md) for troubleshooting
- Look at existing extensions in the codebase
- Test with simple cases first
- Use `--verbose` to see what's happening

Ready to debug issues? See the [Debugging Guide](debugging.md).