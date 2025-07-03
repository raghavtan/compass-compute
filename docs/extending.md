# Extension Guide: Custom Fact Types, Data Sources, and Rules

## Overview

The modular facts package provides three main extension points to add custom functionality without modifying core code:

1. **Custom Fact Types** - New ways to process facts (e.g., kubernetes, database, monitoring)
2. **Custom Data Sources** - New places to extract data from (e.g., databases, APIs, file systems)
3. **Custom Rules** - New ways to process extracted data (e.g., math operations, transformations)

## 1. Custom Fact Types

Custom fact types define new ways to process facts. They're added to the `processCustom()` function in `evaluator.go`.

### Example: Kubernetes Fact Type

```go
// Add to evaluator.go in processCustom() function
func (fe *FactEvaluator) processCustom(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) error {
    switch strings.ToLower(fact.Type) {
    case "kubernetes", "k8s":
        return fe.processKubernetes(ctx, fact)
    case "database":
        return fe.processDatabase(ctx, fact)
    case "monitoring":
        return fe.processMonitoring(ctx, fact)
    default:
        return fmt.Errorf("unknown fact type: %s", fact.Type)
    }
}

// Kubernetes processor implementation
func (fe *FactEvaluator) processKubernetes(ctx context.Context, fact *services.Fact) error {
    // Use existing fact fields creatively:
    namespace := fact.Repo           // Repository field as namespace
    resourceType := fact.FilePath    // FilePath field as resource type
    labelSelector := fact.SearchString // SearchString field as label selector
    
    switch resourceType {
    case "pods":
        count, err := fe.getPodCount(namespace, labelSelector)
        if err != nil {
            return fmt.Errorf("failed to get pod count: %w", err)
        }
        fact.Result = float64(count)
        
    case "services":
        count, err := fe.getServiceCount(namespace, labelSelector)
        if err != nil {
            return fmt.Errorf("failed to get service count: %w", err)
        }
        fact.Result = float64(count)
        
    case "deployments":
        status, err := fe.getDeploymentStatus(namespace, labelSelector)
        if err != nil {
            return fmt.Errorf("failed to get deployment status: %w", err)
        }
        fact.Result = status
        
    default:
        return fmt.Errorf("unsupported kubernetes resource type: %s", resourceType)
    }
    
    return nil
}

// Helper functions for Kubernetes operations
func (fe *FactEvaluator) getPodCount(namespace, labelSelector string) (int, error) {
    // Implementation using kubectl or Kubernetes client-go
    // Example using kubectl command:
    cmd := exec.CommandContext(context.Background(), "kubectl", "get", "pods", "-n", namespace)
    if labelSelector != "" {
        cmd.Args = append(cmd.Args, "-l", labelSelector)
    }
    cmd.Args = append(cmd.Args, "--no-headers")
    
    output, err := cmd.Output()
    if err != nil {
        return 0, err
    }
    
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    if len(lines) == 1 && lines[0] == "" {
        return 0, nil
    }
    
    return len(lines), nil
}

func (fe *FactEvaluator) getServiceCount(namespace, labelSelector string) (int, error) {
    // Similar implementation for services
    cmd := exec.CommandContext(context.Background(), "kubectl", "get", "svc", "-n", namespace)
    if labelSelector != "" {
        cmd.Args = append(cmd.Args, "-l", labelSelector)
    }
    cmd.Args = append(cmd.Args, "--no-headers")
    
    output, err := cmd.Output()
    if err != nil {
        return 0, err
    }
    
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    if len(lines) == 1 && lines[0] == "" {
        return 0, nil
    }
    
    return len(lines), nil
}

func (fe *FactEvaluator) getDeploymentStatus(namespace, labelSelector string) (string, error) {
    // Get deployment status
    cmd := exec.CommandContext(context.Background(), "kubectl", "get", "deployment", "-n", namespace)
    if labelSelector != "" {
        cmd.Args = append(cmd.Args, "-l", labelSelector)
    }
    cmd.Args = append(cmd.Args, "-o", "jsonpath={.items[0].status.conditions[?(@.type=='Available')].status}")
    
    output, err := cmd.Output()
    if err != nil {
        return "Unknown", err
    }
    
    status := strings.TrimSpace(string(output))
    if status == "True" {
        return "Available", nil
    }
    return "NotAvailable", nil
}
```

### Example: Database Fact Type

```go
// Add to processCustom() function
case "database":
    return fe.processDatabase(ctx, fact)

// Database processor implementation
func (fe *FactEvaluator) processDatabase(ctx context.Context, fact *services.Fact) error {
    // Use existing fields:
    connectionString := fact.URI          // URI field as connection string
    query := fact.PrometheusQuery        // PrometheusQuery field as SQL query
    operation := fact.Rule               // Rule field as operation type
    
    switch strings.ToLower(operation) {
    case "count":
        result, err := fe.executeDatabaseCount(connectionString, query)
        if err != nil {
            return fmt.Errorf("database count failed: %w", err)
        }
        fact.Result = float64(result)
        
    case "exists":
        result, err := fe.executeDatabaseExists(connectionString, query)
        if err != nil {
            return fmt.Errorf("database exists check failed: %w", err)
        }
        fact.Result = result
        
    case "value":
        result, err := fe.executeDatabaseValue(connectionString, query)
        if err != nil {
            return fmt.Errorf("database value query failed: %w", err)
        }
        fact.Result = result
        
    default:
        return fmt.Errorf("unsupported database operation: %s", operation)
    }
    
    return nil
}

func (fe *FactEvaluator) executeDatabaseCount(connectionString, query string) (int, error) {
    // Connect to database and execute count query
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return 0, err
    }
    defer db.Close()
    
    var count int
    err = db.QueryRow(query).Scan(&count)
    return count, err
}

func (fe *FactEvaluator) executeDatabaseExists(connectionString, query string) (bool, error) {
    // Execute existence check query
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return false, err
    }
    defer db.Close()
    
    var exists bool
    err = db.QueryRow(query).Scan(&exists)
    return exists, err
}

func (fe *FactEvaluator) executeDatabaseValue(connectionString, query string) (interface{}, error) {
    // Execute value query
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }
    defer db.Close()
    
    var value interface{}
    err = db.QueryRow(query).Scan(&value)
    return value, err
}
```

## 2. Custom Data Sources

Custom data sources define new places to extract data from. They're added to the `extractCustom()` function in `extractors.go`.

### Example: Database Data Source

```go
// Add to extractors.go in extractCustom() function
func (fe *FactEvaluator) extractCustom(ctx context.Context, fact *services.Fact) ([]byte, error) {
    switch strings.ToLower(fact.Source) {
    case "database", "postgres", "mysql":
        return fe.extractFromDatabase(ctx, fact)
    case "redis":
        return fe.extractFromRedis(ctx, fact)
    case "elasticsearch":
        return fe.extractFromElasticsearch(ctx, fact)
    case "filesystem":
        return fe.extractFromFilesystem(fact)
    default:
        return nil, fmt.Errorf("unsupported source: %s", fact.Source)
    }
}

// Database extractor implementation
func (fe *FactEvaluator) extractFromDatabase(ctx context.Context, fact *services.Fact) ([]byte, error) {
    // Use existing fields:
    connectionString := fact.URI          // URI as connection string
    query := fact.PrometheusQuery        // PrometheusQuery as SQL query
    
    // Connect to database
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    defer db.Close()
    
    // Execute query
    rows, err := db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to execute query: %w", err)
    }
    defer rows.Close()
    
    // Get column names
    columns, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("failed to get columns: %w", err)
    }
    
    // Prepare result
    var results []map[string]interface{}
    
    for rows.Next() {
        // Create a slice of interface{} to hold the values
        values := make([]interface{}, len(columns))
        valuePointers := make([]interface{}, len(columns))
        
        for i := range values {
            valuePointers[i] = &values[i]
        }
        
        // Scan the row
        if err := rows.Scan(valuePointers...); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }
        
        // Convert to map
        row := make(map[string]interface{})
        for i, col := range columns {
            row[col] = values[i]
        }
        results = append(results, row)
    }
    
    return json.Marshal(results)
}
```

### Example: Redis Data Source

```go
// Redis extractor implementation
func (fe *FactEvaluator) extractFromRedis(ctx context.Context, fact *services.Fact) ([]byte, error) {
    // Use existing fields:
    addr := fact.URI                    // URI as Redis address
    key := fact.FilePath               // FilePath as Redis key
    operation := fact.Rule             // Rule as Redis operation
    
    // Connect to Redis
    client := redis.NewClient(&redis.Options{
        Addr: addr,
    })
    defer client.Close()
    
    switch strings.ToLower(operation) {
    case "get":
        val, err := client.Get(ctx, key).Result()
        if err != nil {
            return nil, fmt.Errorf("redis GET failed: %w", err)
        }
        return []byte(val), nil
        
    case "exists":
        exists, err := client.Exists(ctx, key).Result()
        if err != nil {
            return nil, fmt.Errorf("redis EXISTS failed: %w", err)
        }
        return json.Marshal(exists > 0)
        
    case "hgetall":
        hash, err := client.HGetAll(ctx, key).Result()
        if err != nil {
            return nil, fmt.Errorf("redis HGETALL failed: %w", err)
        }
        return json.Marshal(hash)
        
    default:
        return nil, fmt.Errorf("unsupported Redis operation: %s", operation)
    }
}
```

### Example: Filesystem Data Source

```go
// Filesystem extractor implementation
func (fe *FactEvaluator) extractFromFilesystem(fact *services.Fact) ([]byte, error) {
    // Use existing fields:
    basePath := fact.Repo              // Repo as base path
    pattern := fact.FilePath           // FilePath as glob pattern
    operation := fact.Rule             // Rule as operation type
    
    fullPattern := filepath.Join(basePath, pattern)
    
    switch strings.ToLower(operation) {
    case "list":
        files, err := filepath.Glob(fullPattern)
        if err != nil {
            return nil, fmt.Errorf("failed to glob files: %w", err)
        }
        return json.Marshal(files)
        
    case "count":
        files, err := filepath.Glob(fullPattern)
        if err != nil {
            return nil, fmt.Errorf("failed to glob files: %w", err)
        }
        return json.Marshal(len(files))
        
    case "sizes":
        files, err := filepath.Glob(fullPattern)
        if err != nil {
            return nil, fmt.Errorf("failed to glob files: %w", err)
        }
        
        var sizes []map[string]interface{}
        for _, file := range files {
            info, err := os.Stat(file)
            if err != nil {
                continue
            }
            sizes = append(sizes, map[string]interface{}{
                "name": file,
                "size": info.Size(),
            })
        }
        return json.Marshal(sizes)
        
    default:
        return nil, fmt.Errorf("unsupported filesystem operation: %s", operation)
    }
}
```

## 3. Custom Rules

Custom rules define new ways to process extracted data. They're added to the `applyCustomRule()` function in `appliers.go`.

### Example: Math Operations Rule

```go
// Add to appliers.go in applyCustomRule() function
func (fe *FactEvaluator) applyCustomRule(fact *services.Fact, data []byte) (interface{}, error) {
    switch strings.ToLower(fact.Rule) {
    case "math", "calculate":
        return fe.applyMath(fact, data)
    case "regex":
        return fe.applyRegex(fact, data)
    case "transform":
        return fe.applyTransform(fact, data)
    case "filter":
        return fe.applyFilter(fact, data)
    default:
        return string(data), nil
    }
}

// Math operations implementation
func (fe *FactEvaluator) applyMath(fact *services.Fact, data []byte) (interface{}, error) {
    var inputData interface{}
    if err := json.Unmarshal(data, &inputData); err != nil {
        return nil, fmt.Errorf("failed to unmarshal data for math operation: %w", err)
    }
    
    operation := strings.ToLower(fact.Pattern)
    
    switch operation {
    case "sum":
        return fe.mathSum(inputData)
    case "average", "avg":
        return fe.mathAverage(inputData)
    case "min":
        return fe.mathMin(inputData)
    case "max":
        return fe.mathMax(inputData)
    case "multiply":
        multiplier, _ := strconv.ParseFloat(fact.SearchString, 64)
        return fe.mathMultiply(inputData, multiplier)
    case "percentage":
        total, _ := strconv.ParseFloat(fact.SearchString, 64)
        return fe.mathPercentage(inputData, total)
    default:
        return nil, fmt.Errorf("unsupported math operation: %s", operation)
    }
}

func (fe *FactEvaluator) mathSum(data interface{}) (float64, error) {
    switch v := data.(type) {
    case []interface{}:
        sum := 0.0
        for _, item := range v {
            if val := convertToFloat64(item); val != nil {
                sum += *val
            }
        }
        return sum, nil
    case map[string]interface{}:
        sum := 0.0
        for _, value := range v {
            if val := convertToFloat64(value); val != nil {
                sum += *val
            }
        }
        return sum, nil
    default:
        if val := convertToFloat64(data); val != nil {
            return *val, nil
        }
        return 0, fmt.Errorf("cannot sum data of type %T", data)
    }
}

func (fe *FactEvaluator) mathAverage(data interface{}) (float64, error) {
    switch v := data.(type) {
    case []interface{}:
        if len(v) == 0 {
            return 0, nil
        }
        sum := 0.0
        count := 0
        for _, item := range v {
            if val := convertToFloat64(item); val != nil {
                sum += *val
                count++
            }
        }
        if count == 0 {
            return 0, fmt.Errorf("no numeric values found")
        }
        return sum / float64(count), nil
    default:
        return 0, fmt.Errorf("cannot average data of type %T", data)
    }
}

func (fe *FactEvaluator) mathMultiply(data interface{}, multiplier float64) (float64, error) {
    if val := convertToFloat64(data); val != nil {
        return *val * multiplier, nil
    }
    return 0, fmt.Errorf("cannot multiply data of type %T", data)
}

func (fe *FactEvaluator) mathPercentage(data interface{}, total float64) (float64, error) {
    if val := convertToFloat64(data); val != nil {
        if total == 0 {
            return 0, fmt.Errorf("cannot calculate percentage with total of 0")
        }
        return (*val / total) * 100, nil
    }
    return 0, fmt.Errorf("cannot calculate percentage for data of type %T", data)
}
```

### Example: Regex Rule

```go
// Regex operations implementation
func (fe *FactEvaluator) applyRegex(fact *services.Fact, data []byte) (interface{}, error) {
    pattern := fact.Pattern
    operation := strings.ToLower(fact.SearchString)
    
    regex, err := regexp.Compile(pattern)
    if err != nil {
        return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
    }
    
    dataStr := string(data)
    
    switch operation {
    case "match":
        return regex.MatchString(dataStr), nil
    case "find":
        return regex.FindString(dataStr), nil
    case "findall":
        return regex.FindAllString(dataStr, -1), nil
    case "replace":
        replacement := fact.PrometheusQuery // Reuse field for replacement string
        return regex.ReplaceAllString(dataStr, replacement), nil
    case "count":
        matches := regex.FindAllString(dataStr, -1)
        return len(matches), nil
    default:
        return nil, fmt.Errorf("unsupported regex operation: %s", operation)
    }
}
```

### Example: Transform Rule

```go
// Transform operations implementation
func (fe *FactEvaluator) applyTransform(fact *services.Fact, data []byte) (interface{}, error) {
    var inputData interface{}
    if err := json.Unmarshal(data, &inputData); err != nil {
        // If it's not JSON, treat as string
        inputData = string(data)
    }
    
    transform := strings.ToLower(fact.Pattern)
    
    switch transform {
    case "uppercase":
        if str, ok := inputData.(string); ok {
            return strings.ToUpper(str), nil
        }
        return nil, fmt.Errorf("uppercase transform requires string input")
        
    case "lowercase":
        if str, ok := inputData.(string); ok {
            return strings.ToLower(str), nil
        }
        return nil, fmt.Errorf("lowercase transform requires string input")
        
    case "length":
        switch v := inputData.(type) {
        case string:
            return len(v), nil
        case []interface{}:
            return len(v), nil
        case map[string]interface{}:
            return len(v), nil
        default:
            return nil, fmt.Errorf("cannot get length of type %T", inputData)
        }
        
    case "keys":
        if m, ok := inputData.(map[string]interface{}); ok {
            keys := make([]string, 0, len(m))
            for k := range m {
                keys = append(keys, k)
            }
            return keys, nil
        }
        return nil, fmt.Errorf("keys transform requires map input")
        
    case "values":
        if m, ok := inputData.(map[string]interface{}); ok {
            values := make([]interface{}, 0, len(m))
            for _, v := range m {
                values = append(values, v)
            }
            return values, nil
        }
        return nil, fmt.Errorf("values transform requires map input")
        
    default:
        return nil, fmt.Errorf("unsupported transform operation: %s", transform)
    }
}
```

## Usage Examples

### Using Custom Fact Types

```yaml
# Kubernetes pod count
apiVersion: v1
kind: Metric
metadata:
  name: pod-count
  facts:
    - id: get-pod-count
      type: kubernetes
      repo: production        # namespace
      filePath: pods         # resource type
      searchString: app=myapp # label selector
```

```yaml
# Database user count
apiVersion: v1
kind: Metric
metadata:
  name: active-users
  facts:
    - id: count-users
      type: database
      uri: postgres://user:pass@localhost/db
      prometheusQuery: "SELECT COUNT(*) FROM users WHERE active = true"
      rule: count
```

### Using Custom Data Sources

```yaml
# Redis cache check
apiVersion: v1
kind: Metric
metadata:
  name: cache-status
  facts:
    - id: check-cache
      type: extract
      source: redis
      uri: localhost:6379
      filePath: cache:status
      rule: get
```

```yaml
# Database query with JSON processing
apiVersion: v1
kind: Metric
metadata:
  name: user-metrics
  facts:
    - id: get-user-data
      type: extract
      source: database
      uri: postgres://user:pass@localhost/db
      prometheusQuery: "SELECT count(*) as total, avg(age) as avg_age FROM users"
      rule: jsonpath
      jsonPath: ".[0].total"
```

### Using Custom Rules

```yaml
# Math operations
apiVersion: v1
kind: Metric
metadata:
  name: calculated-metric
  facts:
    - id: get-raw-data
      type: extract
      source: api
      uri: https://api.example.com/metrics
      rule: math
      pattern: sum          # operation
      searchString: "100"   # parameter
```

```yaml
# Regex matching
apiVersion: v1
kind: Metric
metadata:
  name: version-check
  facts:
    - id: check-version
      type: extract
      source: github
      repo: my-service
      filePath: package.json
      rule: regex
      pattern: '"version":\s*"([^"]+)"'
      searchString: find
```

## Best Practices

### 1. Field Reuse Strategy
Since the `Fact` struct has fixed fields, reuse them creatively:
- `Repo` → namespace, database name, base path
- `FilePath` → resource type, key, file pattern
- `URI` → connection string, API endpoint
- `PrometheusQuery` → SQL query, replacement string
- `SearchString` → label selector, operation parameter
- `Pattern` → regex pattern, operation type

### 2. Error Handling
Always provide clear error messages:
```go
if err != nil {
    return fmt.Errorf("kubernetes pod count failed for namespace '%s': %w", namespace, err)
}
```

### 3. Type Safety
Use type assertions safely:
```go
if val := convertToFloat64(data); val != nil {
    return *val, nil
}
return 0, fmt.Errorf("cannot convert %T to float64", data)
```

### 4. Resource Management
Always clean up resources:
```go
defer db.Close()
defer client.Close()
defer resp.Body.Close()
```

### 5. Context Usage
Always respect context for cancellation:
```go
func (fe *FactEvaluator) extractFromDatabase(ctx context.Context, fact *services.Fact) ([]byte, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // proceed with extraction
    }
}
```

This extension system allows you to add any custom functionality while maintaining the simple, consistent interface of the facts package!