# 🔌 Extension Guide: Supercharge compass-compute

> **Extend compass-compute without touching the core - build custom fact types, data sources, and processing rules that integrate seamlessly**

Whether you need to connect to proprietary systems, implement custom business logic, or create specialized data transformations, this guide will show you how to extend compass-compute while keeping your changes maintainable and upgrade-safe.

---

## 🎯 What You Can Extend

### 🏗️ **Custom Fact Types**
Add entirely new ways to process and evaluate data:
- **Kubernetes Integration** - Pod counts, deployment status, resource usage
- **Database Queries** - Custom SQL, NoSQL operations, schema validation
- **Security Scans** - Vulnerability assessments, compliance checks
- **Business Logic** - Complex calculations, AI/ML model inference

### 🌐 **Custom Data Sources**
Connect to any system or service:
- **Enterprise APIs** - Internal services, LDAP, Active Directory
- **Cloud Services** - AWS CloudWatch, Azure Monitor, GCP Logging
- **Databases** - PostgreSQL, MongoDB, Redis, Elasticsearch
- **File Systems** - Network shares, S3 buckets, specialized formats

### ⚙️ **Custom Processing Rules**
Transform and manipulate data in powerful ways:
- **Mathematical Operations** - Statistics, aggregations, formulas
- **Text Processing** - NLP, sentiment analysis, pattern extraction
- **Data Validation** - Custom business rules, schema validation
- **Transformations** - Format conversion, data enrichment

---

## 🚀 Quick Start: Your First Extension

Let's build a simple Kubernetes integration in 10 minutes:

### 1. Add Custom Fact Type

```go
// Add to internal/facts/evaluator.go in processCustom()
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

```go
func (fe *FactEvaluator) processKubernetes(ctx context.Context, fact *services.Fact) error {
    namespace := fact.Repo           // Creative field reuse
    resourceType := fact.FilePath    
    labelSelector := fact.SearchString
    
    switch resourceType {
    case "pods":
        count, err := fe.getPodCount(namespace, labelSelector)
        if err != nil {
            return fmt.Errorf("failed to get pod count: %w", err)
        }
        fact.Result = float64(count)
        return nil
    }
    return fmt.Errorf("unsupported k8s resource: %s", resourceType)
}
```

### 3. Use in Your Metrics

```yaml
apiVersion: v1
kind: Metric
metadata:
  name: pod-count
  facts:
    - id: count-pods
      type: kubernetes
      repo: production        # namespace
      filePath: pods         # resource type  
      searchString: app=myapp # label selector
```

**That's it!** You've created a custom Kubernetes integration.

---

## 🏗️ Custom Fact Types: Deep Dive

### Architecture Overview

Custom fact types plug into the evaluation engine through the `processCustom()` function:

```go
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Fact Evaluator │────│  processCustom() │────│ Your Extensions │
│     Engine      │    │                  │    │                 │
└─────────────────┘    │  • kubernetes    │    │ Custom Logic    │
                       │  • database      │    │ External APIs   │
                       │  • monitoring    │    │ Business Rules  │
                       └──────────────────┘    └─────────────────┘
```

### 🐳 Example: Advanced Kubernetes Integration

Let's build a comprehensive Kubernetes fact type:

```go
func (fe *FactEvaluator) processKubernetes(ctx context.Context, fact *services.Fact) error {
    // Field mapping for clarity
    namespace := fact.Repo           
    resourceType := fact.FilePath    
    labelSelector := fact.SearchString
    operation := fact.Rule
    
    switch resourceType {
    case "pods":
        return fe.handleKubernetesPods(ctx, fact, namespace, labelSelector, operation)
    case "services":
        return fe.handleKubernetesServices(ctx, fact, namespace, labelSelector, operation)
    case "deployments":
        return fe.handleKubernetesDeployments(ctx, fact, namespace, labelSelector, operation)
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
        
    case "ready":
        readyCount, totalCount, err := fe.getPodReadiness(namespace, labelSelector)
        if err != nil {
            return err
        }
        fact.Result = map[string]interface{}{
            "ready": readyCount,
            "total": totalCount,
            "percentage": float64(readyCount) / float64(totalCount) * 100,
        }
        
    case "resources":
        resources, err := fe.getPodResources(namespace, labelSelector)
        if err != nil {
            return err
        }
        fact.Result = resources
        
    default:
        return fmt.Errorf("unsupported pod operation: %s", operation)
    }
    return nil
}

// Helper functions using kubectl or client-go
func (fe *FactEvaluator) getPodCount(namespace, labelSelector string) (int, error) {
    cmd := exec.CommandContext(context.Background(), "kubectl", "get", "pods", "-n", namespace)
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

func (fe *FactEvaluator) getPodReadiness(namespace, labelSelector string) (ready, total int, err error) {
    cmd := exec.CommandContext(context.Background(), "kubectl", "get", "pods", "-n", namespace)
    if labelSelector != "" {
        cmd.Args = append(cmd.Args, "-l", labelSelector)
    }
    cmd.Args = append(cmd.Args, "-o", "jsonpath={range .items[*]}{.status.conditions[?(@.type=='Ready')].status}{\"\n\"}{end}")
    
    output, err := cmd.Output()
    if err != nil {
        return 0, 0, fmt.Errorf("kubectl readiness check failed: %w", err)
    }
    
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    total = len(lines)
    
    for _, line := range lines {
        if strings.TrimSpace(line) == "True" {
            ready++
        }
    }
    
    return ready, total, nil
}
```

### 💾 Example: Database Integration

Build a powerful database fact type for SQL and NoSQL operations:

```go
func (fe *FactEvaluator) processDatabase(ctx context.Context, fact *services.Fact) error {
    connectionString := fact.URI
    query := fact.PrometheusQuery  // Reuse for SQL query
    operation := fact.Rule
    dbType := fact.Pattern         // Database type (postgres, mysql, mongodb)
    
    switch strings.ToLower(dbType) {
    case "postgres", "postgresql":
        return fe.handlePostgreSQL(ctx, fact, connectionString, query, operation)
    case "mysql":
        return fe.handleMySQL(ctx, fact, connectionString, query, operation)
    case "mongodb", "mongo":
        return fe.handleMongoDB(ctx, fact, connectionString, query, operation)
    case "redis":
        return fe.handleRedis(ctx, fact, connectionString, query, operation)
    default:
        return fmt.Errorf("unsupported database type: %s", dbType)
    }
}

func (fe *FactEvaluator) handlePostgreSQL(ctx context.Context, fact *services.Fact, connectionString, query, operation string) error {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
    }
    defer db.Close()
    
    // Test connection
    if err := db.PingContext(ctx); err != nil {
        return fmt.Errorf("failed to ping PostgreSQL: %w", err)
    }
    
    switch operation {
    case "count":
        return fe.executeSQLCount(ctx, db, fact, query)
    case "value":
        return fe.executeSQLValue(ctx, db, fact, query)
    case "exists":
        return fe.executeSQLExists(ctx, db, fact, query)
    case "health":
        fact.Result = true  // If we got here, connection is healthy
        return nil
    default:
        return fmt.Errorf("unsupported PostgreSQL operation: %s", operation)
    }
}

func (fe *FactEvaluator) executeSQLCount(ctx context.Context, db *sql.DB, fact *services.Fact, query string) error {
    var count int64
    err := db.QueryRowContext(ctx, query).Scan(&count)
    if err != nil {
        return fmt.Errorf("SQL count query failed: %w", err)
    }
    fact.Result = float64(count)
    return nil
}

func (fe *FactEvaluator) executeSQLValue(ctx context.Context, db *sql.DB, fact *services.Fact, query string) error {
    var value interface{}
    err := db.QueryRowContext(ctx, query).Scan(&value)
    if err != nil {
        return fmt.Errorf("SQL value query failed: %w", err)
    }
    
    // Convert to appropriate type
    if val := convertToFloat64(value); val != nil {
        fact.Result = *val
    } else {
        fact.Result = fmt.Sprintf("%v", value)
    }
    return nil
}
```

---

## 🌐 Custom Data Sources: Connect Everything

### Adding New Data Sources

Data sources are implemented in the `extractCustom()` function:

```go
// Add to internal/facts/extractor.go
func (fe *FactEvaluator) extractCustom(ctx context.Context, fact *services.Fact) ([]byte, error) {
    switch strings.ToLower(fact.Source) {
    case "elasticsearch":
        return fe.extractFromElasticsearch(ctx, fact)
    case "s3":
        return fe.extractFromS3(ctx, fact)
    case "vault":
        return fe.extractFromVault(ctx, fact)
    default:
        return nil, fmt.Errorf("unsupported source: %s", fact.Source)
    }
}
```

### 🔍 Example: Elasticsearch Integration

```go
func (fe *FactEvaluator) extractFromElasticsearch(ctx context.Context, fact *services.Fact) ([]byte, error) {
    // Field mapping
    esURL := fact.URI               // Elasticsearch URL
    index := fact.Repo              // Index name
    query := fact.PrometheusQuery   // ES query JSON
    operation := fact.Rule          // search, count, aggregate
    
    client, err := elasticsearch.NewClient(elasticsearch.Config{
        Addresses: []string{esURL},
        Username:  os.Getenv("ES_USERNAME"),
        Password:  os.Getenv("ES_PASSWORD"),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
    }
    
    switch operation {
    case "search":
        return fe.elasticsearchSearch(ctx, client, index, query)
    case "count":
        return fe.elasticsearchCount(ctx, client, index, query)
    case "aggregate":
        return fe.elasticsearchAggregate(ctx, client, index, query)
    default:
        return nil, fmt.Errorf("unsupported ES operation: %s", operation)
    }
}

func (fe *FactEvaluator) elasticsearchSearch(ctx context.Context, client *elasticsearch.Client, index, query string) ([]byte, error) {
    res, err := client.Search(
        client.Search.WithContext(ctx),
        client.Search.WithIndex(index),
        client.Search.WithBody(strings.NewReader(query)),
        client.Search.WithSize(1000),
    )
    if err != nil {
        return nil, fmt.Errorf("Elasticsearch search failed: %w", err)
    }
    defer res.Body.Close()
    
    if res.IsError() {
        return nil, fmt.Errorf("Elasticsearch error: %s", res.Status())
    }
    
    return io.ReadAll(res.Body)
}
```

### ☁️ Example: AWS S3 Integration

```go
func (fe *FactEvaluator) extractFromS3(ctx context.Context, fact *services.Fact) ([]byte, error) {
    // Field mapping
    bucket := fact.Repo             // S3 bucket name
    key := fact.FilePath           // Object key
    region := fact.Pattern         // AWS region
    operation := fact.Rule         // get, list, metadata
    
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    
    client := s3.NewFromConfig(cfg)
    
    switch operation {
    case "get":
        return fe.s3GetObject(ctx, client, bucket, key)
    case "list":
        return fe.s3ListObjects(ctx, client, bucket, key) // key as prefix
    case "metadata":
        return fe.s3GetMetadata(ctx, client, bucket, key)
    default:
        return nil, fmt.Errorf("unsupported S3 operation: %s", operation)
    }
}

func (fe *FactEvaluator) s3GetObject(ctx context.Context, client *s3.Client, bucket, key string) ([]byte, error) {
    result, err := client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        return nil, fmt.Errorf("S3 GetObject failed: %w", err)
    }
    defer result.Body.Close()
    
    return io.ReadAll(result.Body)
}

func (fe *FactEvaluator) s3ListObjects(ctx context.Context, client *s3.Client, bucket, prefix string) ([]byte, error) {
    result, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
        Bucket: aws.String(bucket),
        Prefix: aws.String(prefix),
    })
    if err != nil {
        return nil, fmt.Errorf("S3 ListObjects failed: %w", err)
    }
    
    var objects []map[string]interface{}
    for _, obj := range result.Contents {
        objects = append(objects, map[string]interface{}{
            "key":          *obj.Key,
            "size":         *obj.Size,
            "lastModified": obj.LastModified.Format(time.RFC3339),
        })
    }
    
    return json.Marshal(objects)
}
```

---

## ⚙️ Custom Processing Rules: Transform Data

### Rule Architecture

Processing rules are implemented in the `applyCustomRule()` function:

```go
// Add to internal/facts/appliers.go
func (fe *FactEvaluator) applyCustomRule(fact *services.Fact, data []byte) (interface{}, error) {
    switch strings.ToLower(fact.Rule) {
    case "math", "calculate":
        return fe.applyMath(fact, data)
    case "nlp", "sentiment":
        return fe.applyNLP(fact, data)
    case "crypto":
        return fe.applyCrypto(fact, data)
    case "ai", "ml":
        return fe.applyML(fact, data)
    default:
        return string(data), nil // Default: return as string
    }
}
```

### 🧮 Example: Advanced Math Operations

```go
func (fe *FactEvaluator) applyMath(fact *services.Fact, data []byte) (interface{}, error) {
    var inputData interface{}
    if err := json.Unmarshal(data, &inputData); err != nil {
        return nil, fmt.Errorf("failed to unmarshal data for math operation: %w", err)
    }
    
    operation := strings.ToLower(fact.Pattern)
    parameter := fact.SearchString
    
    switch operation {
    case "sum", "total":
        return fe.mathSum(inputData)
    case "average", "avg", "mean":
        return fe.mathAverage(inputData)
    case "median":
        return fe.mathMedian(inputData)
    case "percentile":
        percentile, err := strconv.ParseFloat(parameter, 64)
        if err != nil {
            return nil, fmt.Errorf("invalid percentile value: %s", parameter)
        }
        return fe.mathPercentile(inputData, percentile)
    case "stddev", "stdev":
        return fe.mathStandardDeviation(inputData)
    case "normalize":
        return fe.mathNormalize(inputData)
    case "formula":
        return fe.mathFormula(inputData, parameter) // Custom formula evaluation
    default:
        return nil, fmt.Errorf("unsupported math operation: %s", operation)
    }
}

func (fe *FactEvaluator) mathPercentile(data interface{}, percentile float64) (float64, error) {
    values, err := fe.extractNumbers(data)
    if err != nil {
        return 0, err
    }
    
    if len(values) == 0 {
        return 0, fmt.Errorf("no numeric values found")
    }
    
    // Sort values
    sort.Float64s(values)
    
    // Calculate percentile
    index := (percentile / 100.0) * float64(len(values)-1)
    lower := int(math.Floor(index))
    upper := int(math.Ceil(index))
    
    if lower == upper {
        return values[lower], nil
    }
    
    // Linear interpolation
    weight := index - float64(lower)
    return values[lower]*(1-weight) + values[upper]*weight, nil
}

func (fe *FactEvaluator) mathStandardDeviation(data interface{}) (float64, error) {
    values, err := fe.extractNumbers(data)
    if err != nil {
        return 0, err
    }
    
    if len(values) < 2 {
        return 0, fmt.Errorf("need at least 2 values for standard deviation")
    }
    
    // Calculate mean
    sum := 0.0
    for _, v := range values {
        sum += v
    }
    mean := sum / float64(len(values))
    
    // Calculate variance
    variance := 0.0
    for _, v := range values {
        variance += math.Pow(v-mean, 2)
    }
    variance /= float64(len(values) - 1) // Sample standard deviation
    
    return math.Sqrt(variance), nil
}

func (fe *FactEvaluator) mathFormula(data interface{}, formula string) (interface{}, error) {
    // Use a formula parser like govaluate or implement custom logic
    evaluator, err := govaluate.NewEvaluableExpression(formula)
    if err != nil {
        return nil, fmt.Errorf("invalid formula: %w", err)
    }
    
    // Convert data to parameters
    params, err := fe.dataToParameters(data)
    if err != nil {
        return nil, err
    }
    
    result, err := evaluator.Evaluate(params)
    if err != nil {
        return nil, fmt.Errorf("formula evaluation failed: %w", err)
    }
    
    return result, nil
}
```

### 🧠 Example: NLP and Sentiment Analysis

```go
func (fe *FactEvaluator) applyNLP(fact *services.Fact, data []byte) (interface{}, error) {
    operation := strings.ToLower(fact.Pattern)
    text := string(data)
    
    switch operation {
    case "sentiment":
        return fe.analyzeSentiment(text)
    case "keywords":
        return fe.extractKeywords(text)
    case "language":
        return fe.detectLanguage(text)
    case "readability":
        return fe.calculateReadability(text)
    case "entities":
        return fe.extractEntities(text)
    default:
        return nil, fmt.Errorf("unsupported NLP operation: %s", operation)
    }
}

func (fe *FactEvaluator) analyzeSentiment(text string) (map[string]interface{}, error) {
    // Simple rule-based sentiment analysis (you can integrate with cloud APIs)
    positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "love", "like"}
    negativeWords := []string{"bad", "terrible", "awful", "hate", "dislike", "horrible", "worst"}
    
    text = strings.ToLower(text)
    words := strings.Fields(text)
    
    positiveCount := 0
    negativeCount := 0
    
    for _, word := range words {
        for _, pos := range positiveWords {
            if strings.Contains(word, pos) {
                positiveCount++
            }
        }
        for _, neg := range negativeWords {
            if strings.Contains(word, neg) {
                negativeCount++
            }
        }
    }
    
    score := float64(positiveCount-negativeCount) / float64(len(words))
    sentiment := "neutral"
    
    if score > 0.1 {
        sentiment = "positive"
    } else if score < -0.1 {
        sentiment = "negative"
    }
    
    return map[string]interface{}{
        "sentiment":       sentiment,
        "score":          score,
        "positiveWords":  positiveCount,
        "negativeWords":  negativeCount,
        "totalWords":     len(words),
    }, nil
}

func (fe *FactEvaluator) extractKeywords(text string) ([]string, error) {
    // Simple keyword extraction based on frequency and length
    words := strings.Fields(strings.ToLower(text))
    wordCount := make(map[string]int)
    
    // Count word frequencies (excluding common stop words)
    stopWords := map[string]bool{
        "the": true, "and": true, "or": true, "but": true, "in": true,
        "on": true, "at": true, "to": true, "for": true, "of": true,
        "with": true, "by": true, "is": true, "are": true, "was": true,
        "were": true, "be": true, "been": true, "have": true, "has": true,
        "had": true, "do": true, "does": true, "did": true, "will": true,
        "would": true, "could": true, "should": true, "may": true, "might": true,
        "must": true, "can": true, "cannot": true, "a": true, "an": true,
    }
    
    for _, word := range words {
        // Clean word
        word = strings.Trim(word, ".,!?;:()")
        if len(word) > 3 && !stopWords[word] {
            wordCount[word]++
        }
    }
    
    // Sort by frequency
    type wordFreq struct {
        word  string
        count int
    }
    
    var wordFreqs []wordFreq
    for word, count := range wordCount {
        wordFreqs = append(wordFreqs, wordFreq{word, count})
    }
    
    sort.Slice(wordFreqs, func(i, j int) bool {
        return wordFreqs[i].count > wordFreqs[j].count
    })
    
    // Return top 10 keywords
    var keywords []string
    for i, wf := range wordFreqs {
        if i >= 10 {
            break
        }
        keywords = append(keywords, wf.word)
    }
    
    return keywords, nil
}
```

---

## 🎯 Real-World Extension Examples

### 🔐 Security Compliance Extension

```go
func (fe *FactEvaluator) processSecurity(ctx context.Context, fact *services.Fact) error {
    scanType := fact.Pattern       // vulnerability, compliance, secrets
    target := fact.URI            // What to scan
    operation := fact.Rule        // scan, report, remediate
    
    switch scanType {
    case "vulnerability":
        return fe.securityVulnerabilityScan(ctx, fact, target, operation)
    case "compliance":
        return fe.securityComplianceScan(ctx, fact, target, operation)
    case "secrets":
        return fe.securitySecretsCheck(ctx, fact, target, operation)
    default:
        return fmt.Errorf("unsupported security scan: %s", scanType)
    }
}

func (fe *FactEvaluator) securityVulnerabilityScan(ctx context.Context, fact *services.Fact, target, operation string) error {
    switch operation {
    case "scan":
        // Run vulnerability scanner (e.g., Trivy, Snyk)
        cmd := exec.CommandContext(ctx, "trivy", "image", target, "--format", "json")
        output, err := cmd.Output()
        if err != nil {
            return fmt.Errorf("vulnerability scan failed: %w", err)
        }
        
        var results map[string]interface{}
        if err := json.Unmarshal(output, &results); err != nil {
            return fmt.Errorf("failed to parse scan results: %w", err)
        }
        
        fact.Result = results
        return nil
        
    case "report":
        // Generate vulnerability report
        fact.Result = fe.generateVulnerabilityReport(target)
        return nil
    }
    
    return fmt.Errorf("unsupported vulnerability operation: %s", operation)
}
```

### 📊 Business Intelligence Extension

```go
func (fe *FactEvaluator) processBI(ctx context.Context, fact *services.Fact) error {
    analysisType := fact.Pattern   // trend, forecast, anomaly
    dataSource := fact.URI        // Data source URL/connection
    operation := fact.Rule        // analyze, predict, detect
    
    switch analysisType {
    case "trend":
        return fe.biTrendAnalysis(ctx, fact, dataSource, operation)
    case "forecast":
        return fe.biForecast(ctx, fact, dataSource, operation)
    case "anomaly":
        return fe.biAnomalyDetection(ctx, fact, dataSource, operation)
    default:
        return fmt.Errorf("unsupported BI analysis: %s", analysisType)
    }
}

func (fe *FactEvaluator) biTrendAnalysis(ctx context.Context, fact *services.Fact, dataSource, operation string) error {
    // Fetch time series data
    data, err := fe.fetchTimeSeriesData(ctx, dataSource)
    if err != nil {
        return err
    }
    
    // Analyze trends using statistical methods
    trend := fe.calculateTrend(data)
    
    fact.Result = map[string]interface{}{
        "direction": trend.Direction, // "up", "down", "stable"
        "slope":     trend.Slope,
        "confidence": trend.Confidence,
        "period":    trend.Period,
    }
    
    return nil
}
```

---

## 🔧 Integration Patterns

### 🔌 Plugin Architecture

Create reusable extension modules:

```go
// Define extension interface
type FactExtension interface {
    Name() string
    SupportedTypes() []string
    Process(ctx context.Context, fact *services.Fact) error
}

// Register extensions
var extensions = map[string]FactExtension{
    "kubernetes": &KubernetesExtension{},
    "security":   &SecurityExtension{},
    "bi":         &BusinessIntelligenceExtension{},
}

// Modified processCustom to use extensions
func (fe *FactEvaluator) processCustom(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) error {
    ext, exists := extensions[strings.ToLower(fact.Type)]
    if !exists {
        return fmt.Errorf("unknown fact type: %s", fact.Type)
    }
    
    return ext.Process(ctx, fact)
}

// Example extension implementation
type KubernetesExtension struct {
    client kubernetes.Interface
}

func (k *KubernetesExtension) Name() string {
    return "kubernetes"
}

func (k *KubernetesExtension) SupportedTypes() []string {
    return []string{"kubernetes", "k8s"}
}

func (k *KubernetesExtension) Process(ctx context.Context, fact *services.Fact) error {
    // Implementation here
    return k.processKubernetesResource(ctx, fact)
}
```

### 🏗️ Configuration-Driven Extensions

Create extensions that can be configured without code changes:

```yaml
# extensions.yaml
extensions:
  kubernetes:
    enabled: true
    config:
      defaultNamespace: "default"
      timeout: "30s"
      kubeconfig: "${HOME}/.kube/config"
  
  database:
    enabled: true
    config:
      maxConnections: 10
      timeout: "15s"
      poolSize: 5
  
  ai:
    enabled: false
    config:
      provider: "openai"
      model: "gpt-4"
      apiKey: "${OPENAI_API_KEY}"
```

### 🌍 Environment-Specific Extensions

Create extensions that work differently in different environments:

```go
type EnvironmentAwareExtension struct {
    env string
}

func (e *EnvironmentAwareExtension) Process(ctx context.Context, fact *services.Fact) error {
    switch e.env {
    case "production":
        return e.processProduction(ctx, fact)
    case "staging":
        return e.processStaging(ctx, fact)
    case "development":
        return e.processDevelopment(ctx, fact)
    default:
        return e.processDefault(ctx, fact)
    }
}
```

---

## 📚 Advanced Usage Examples

### 🤖 AI/ML Integration Extension

```yaml
# AI-powered code quality analysis
apiVersion: v1
kind: Metric
metadata:
  name: ai-code-quality
  facts:
    - id: extract-code
      type: extract
      source: github
      repo: ${Metadata.Name}
      filePath: "**/*.go"
      
    - id: analyze-quality
      type: ai
      dependsOn: [extract-code]
      pattern: code-analysis
      searchString: "quality,maintainability,complexity"
      uri: "https://api.openai.com/v1/chat/completions"
      auth:
        header: "Authorization"
        tokenVar: "OPENAI_API_KEY"
```

```go
func (fe *FactEvaluator) processAI(ctx context.Context, fact *services.Fact) error {
    analysisType := fact.Pattern
    prompt := fact.SearchString
    apiEndpoint := fact.URI
    
    switch analysisType {
    case "code-analysis":
        return fe.aiCodeAnalysis(ctx, fact, prompt, apiEndpoint)
    case "documentation":
        return fe.aiDocumentationAnalysis(ctx, fact, prompt, apiEndpoint)
    case "security":
        return fe.aiSecurityAnalysis(ctx, fact, prompt, apiEndpoint)
    default:
        return fmt.Errorf("unsupported AI analysis: %s", analysisType)
    }
}

func (fe *FactEvaluator) aiCodeAnalysis(ctx context.Context, fact *services.Fact, prompt, apiEndpoint string) error {
    // Get code from dependencies
    deps := getDependencyResults(fact, factMap)
    if len(deps) == 0 {
        return fmt.Errorf("no code data available for AI analysis")
    }
    
    codeContent := fmt.Sprintf("%v", deps[0])
    
    // Prepare AI request
    request := map[string]interface{}{
        "model": "gpt-4",
        "messages": []map[string]string{
            {
                "role": "system",
                "content": "You are a code quality expert. Analyze the provided code and return a JSON response with quality metrics.",
            },
            {
                "role": "user",
                "content": fmt.Sprintf("Analyze this code for %s:\n\n%s", prompt, codeContent),
            },
        },
        "response_format": map[string]string{"type": "json_object"},
    }
    
    // Make API call
    response, err := fe.makeAIRequest(ctx, apiEndpoint, request)
    if err != nil {
        return err
    }
    
    fact.Result = response
    return nil
}
```

### 🔄 Multi-Cloud Integration

```yaml
# Multi-cloud resource monitoring
apiVersion: v1
kind: Metric
metadata:
  name: multi-cloud-resources
  facts:
    - id: aws-resources
      type: cloud
      pattern: aws
      rule: count
      searchString: "ec2,rds,s3"
      
    - id: azure-resources
      type: cloud
      pattern: azure
      rule: count
      searchString: "vm,sql,storage"
      
    - id: gcp-resources
      type: cloud
      pattern: gcp
      rule: count
      searchString: "compute,sql,storage"
      
    - id: total-resources
      type: aggregate
      dependsOn: [aws-resources, azure-resources, gcp-resources]
      method: sum
```

```go
func (fe *FactEvaluator) processCloud(ctx context.Context, fact *services.Fact) error {
    provider := fact.Pattern
    operation := fact.Rule
    resources := strings.Split(fact.SearchString, ",")
    
    switch provider {
    case "aws":
        return fe.processAWS(ctx, fact, operation, resources)
    case "azure":
        return fe.processAzure(ctx, fact, operation, resources)
    case "gcp":
        return fe.processGCP(ctx, fact, operation, resources)
    default:
        return fmt.Errorf("unsupported cloud provider: %s", provider)
    }
}

func (fe *FactEvaluator) processAWS(ctx context.Context, fact *services.Fact, operation string, resources []string) error {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return fmt.Errorf("failed to load AWS config: %w", err)
    }
    
    totalCount := 0
    
    for _, resource := range resources {
        switch strings.TrimSpace(resource) {
        case "ec2":
            ec2Client := ec2.NewFromConfig(cfg)
            result, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
            if err != nil {
                continue
            }
            for _, reservation := range result.Reservations {
                totalCount += len(reservation.Instances)
            }
            
        case "rds":
            rdsClient := rds.NewFromConfig(cfg)
            result, err := rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
            if err != nil {
                continue
            }
            totalCount += len(result.DBInstances)
            
        case "s3":
            s3Client := s3.NewFromConfig(cfg)
            result, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
            if err != nil {
                continue
            }
            totalCount += len(result.Buckets)
        }
    }
    
    fact.Result = float64(totalCount)
    return nil
}
```

### 🏭 Manufacturing/IoT Extension

```yaml
# IoT device monitoring
apiVersion: v1
kind: Metric
metadata:
  name: iot-device-health
  facts:
    - id: device-count
      type: iot
      pattern: mqtt
      uri: "mqtt://iot-broker.company.com:1883"
      rule: count
      searchString: "devices/+/status"
      
    - id: device-health
      type: iot
      pattern: mqtt
      uri: "mqtt://iot-broker.company.com:1883"
      rule: health
      searchString: "devices/+/heartbeat"
      
    - id: alert-count
      type: iot
      pattern: mqtt
      uri: "mqtt://iot-broker.company.com:1883"
      rule: count
      searchString: "alerts/+/critical"
```

```go
func (fe *FactEvaluator) processIoT(ctx context.Context, fact *services.Fact) error {
    protocol := fact.Pattern    // mqtt, http, coap
    brokerURL := fact.URI      // Broker/endpoint URL
    operation := fact.Rule     // count, health, data
    topic := fact.SearchString // MQTT topic pattern
    
    switch protocol {
    case "mqtt":
        return fe.processMQTT(ctx, fact, brokerURL, operation, topic)
    case "http":
        return fe.processHTTPIoT(ctx, fact, brokerURL, operation, topic)
    default:
        return fmt.Errorf("unsupported IoT protocol: %s", protocol)
    }
}

func (fe *FactEvaluator) processMQTT(ctx context.Context, fact *services.Fact, brokerURL, operation, topic string) error {
    opts := mqtt.NewClientOptions()
    opts.AddBroker(brokerURL)
    opts.SetClientID("compass-compute-" + uuid.New().String())
    
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        return fmt.Errorf("MQTT connection failed: %w", token.Error())
    }
    defer client.Disconnect(250)
    
    switch operation {
    case "count":
        return fe.mqttCount(ctx, client, fact, topic)
    case "health":
        return fe.mqttHealth(ctx, client, fact, topic)
    case "data":
        return fe.mqttData(ctx, client, fact, topic)
    default:
        return fmt.Errorf("unsupported MQTT operation: %s", operation)
    }
}

func (fe *FactEvaluator) mqttCount(ctx context.Context, client mqtt.Client, fact *services.Fact, topic string) error {
    messageCount := 0
    done := make(chan bool)
    
    // Subscribe and count messages for a short period
    token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
        messageCount++
    })
    
    if token.Wait() && token.Error() != nil {
        return fmt.Errorf("MQTT subscribe failed: %w", token.Error())
    }
    
    // Wait for messages for 10 seconds
    go func() {
        time.Sleep(10 * time.Second)
        done <- true
    }()
    
    select {
    case <-done:
        fact.Result = float64(messageCount)
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

---

## 🛡️ Best Practices for Extensions

### 🔒 Security Considerations

```go
// 1. Always validate inputs
func validateExtensionInput(fact *services.Fact) error {
    if fact.URI != "" {
        if _, err := url.Parse(fact.URI); err != nil {
            return fmt.Errorf("invalid URI: %w", err)
        }
    }
    
    // Validate other fields
    return nil
}

// 2. Use secure defaults
func getSecureHTTPClient() *http.Client {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
        DisableKeepAlives: true,
        MaxIdleConns:      10,
        IdleConnTimeout:   30 * time.Second,
    }
    
    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }
}

// 3. Sanitize sensitive data in logs
func sanitizeForLogging(data map[string]interface{}) map[string]interface{} {
    sensitiveKeys := []string{"password", "token", "secret", "key", "auth"}
    cleaned := make(map[string]interface{})
    
    for k, v := range data {
        key := strings.ToLower(k)
        isSensitive := false
        
        for _, sensitive := range sensitiveKeys {
            if strings.Contains(key, sensitive) {
                isSensitive = true
                break
            }
        }
        
        if isSensitive {
            cleaned[k] = "[REDACTED]"
        } else {
            cleaned[k] = v
        }
    }
    
    return cleaned
}
```

### 🚀 Performance Optimization

```go
// 1. Use connection pooling
type ExtensionManager struct {
    dbPool    *sql.DB
    httpPool  *http.Client
    redisPool *redis.Pool
}

// 2. Implement caching
type CachedExtension struct {
    cache map[string]CacheEntry
    mutex sync.RWMutex
    ttl   time.Duration
}

type CacheEntry struct {
    Value     interface{}
    ExpiresAt time.Time
}

func (ce *CachedExtension) GetCached(key string) (interface{}, bool) {
    ce.mutex.RLock()
    defer ce.mutex.RUnlock()
    
    entry, exists := ce.cache[key]
    if !exists || time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    
    return entry.Value, true
}

// 3. Use worker pools for concurrent operations
func (fe *FactEvaluator) processMultipleResources(ctx context.Context, resources []string) error {
    const workerCount = 5
    jobs := make(chan string, len(resources))
    results := make(chan error, len(resources))
    
    // Start workers
    for i := 0; i < workerCount; i++ {
        go func() {
            for resource := range jobs {
                results <- fe.processResource(ctx, resource)
            }
        }()
    }
    
    // Send jobs
    for _, resource := range resources {
        jobs <- resource
    }
    close(jobs)
    
    // Collect results
    var errors []error
    for i := 0; i < len(resources); i++ {
        if err := <-results; err != nil {
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("some resources failed: %v", errors)
    }
    
    return nil
}
```

### 🧪 Testing Extensions

```go
// 1. Create testable interfaces
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type Extension struct {
    httpClient HTTPClient
}

// 2. Write comprehensive tests
func TestKubernetesExtension(t *testing.T) {
    tests := []struct {
        name        string
        fact        *services.Fact
        mockCommand func() *exec.Cmd
        expected    interface{}
        wantErr     bool
    }{
        {
            name: "count pods success",
            fact: &services.Fact{
                Type:         "kubernetes",
                Repo:         "default",
                FilePath:     "pods",
                SearchString: "app=test",
                Rule:         "count",
            },
            mockCommand: func() *exec.Cmd {
                return exec.Command("echo", "pod1\npod2\npod3")
            },
            expected: float64(3),
            wantErr:  false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ext := &KubernetesExtension{
                commandRunner: tt.mockCommand,
            }
            
            err := ext.Process(context.Background(), tt.fact)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, tt.fact.Result)
            }
        })
    }
}

// 3. Integration tests
func TestExtensionIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Test with real services in test environment
}
```

---

## 📖 Extension Documentation Template

When creating extensions, document them thoroughly:

```markdown
# MyExtension Documentation

## Overview
Brief description of what the extension does and why it's useful.

## Configuration
```yaml
facts:
  - id: my-fact
    type: myextension
    pattern: operation-type
    uri: connection-string
    rule: processing-rule
```

## Field Mapping
- `pattern`: Operation type (e.g., "scan", "analyze", "monitor")
- `uri`: Connection string or endpoint URL
- `rule`: Processing rule (e.g., "count", "health", "data")
- `repo`: Context identifier (e.g., namespace, database name)

## Operations
### Operation 1: scan
Description and examples...

### Operation 2: analyze
Description and examples...

## Examples
Real-world usage examples...

## Error Handling
Common errors and troubleshooting...

## Dependencies
Required tools, services, or permissions...
```

---

## 🚀 Deployment and Distribution

### 📦 Packaging Extensions

```go
// Create extension packages
package myextension

import (
    "context"
    "github.com/motain/compass-compute/internal/services"
)

// Extension interface implementation
type MyExtension struct {
    config Config
}

type Config struct {
    APIKey    string
    BaseURL   string
    Timeout   time.Duration
}

func New(config Config) *MyExtension {
    return &MyExtension{config: config}
}

func (e *MyExtension) Name() string {
    return "myextension"
}

func (e *MyExtension) Process(ctx context.Context, fact *services.Fact) error {
    // Implementation
    return nil
}

// Register function for easy integration
func Register() {
    extensions.Register("myextension", New(getDefaultConfig()))
}

func getDefaultConfig() Config {
    return Config{
        APIKey:  os.Getenv("MY_EXTENSION_API_KEY"),
        BaseURL: os.Getenv("MY_EXTENSION_BASE_URL"),
        Timeout: 30 * time.Second,
    }
}
```

### 🔧 Installation Guide

```bash
# 1. Add extension to your project
go get github.com/myorg/compass-compute-myextension

# 2. Import and register in main.go
import _ "github.com/myorg/compass-compute-myextension"

# 3. Configure environment variables
export MY_EXTENSION_API_KEY="your-api-key"
export MY_EXTENSION_BASE_URL="https://api.example.com"

# 4. Use in your metrics
cat > metrics/my-metric.yaml << EOF
apiVersion: v1
kind: Metric
metadata:
  name: my-metric
  facts:
    - id: my-fact
      type: myextension
      pattern: scan
      uri: target-url
EOF
```

---

## 🤝 Contributing Extensions

### 📋 Extension Checklist

Before submitting an extension:

- [ ] **Documentation**: Complete README with examples
- [ ] **Tests**: Unit tests with >80% coverage
- [ ] **Error handling**: Graceful error handling and logging
- [ ] **Security**: Input validation and secure defaults
- [ ] **Performance**: Efficient resource usage
- [ ] **Compatibility**: Works with current compass-compute version
- [ ] **Examples**: Real-world usage examples
- [ ] **License**: Compatible license (MIT preferred)

### 🏗️ Extension Template

Use our extension template to get started quickly:

```bash
# Clone the extension template
git clone https://github.com/motain/compass-compute-extension-template
cd compass-compute-extension-template

# Customize for your extension
./customize.sh myextension "My Extension Description"

# Implement your logic
# Test thoroughly
# Submit pull request
```

---

## 🎓 Learning Path

### 🎯 Beginner Extensions
1. **Simple HTTP API** - Start with a basic REST API integration
2. **File Processing** - Create a custom file format processor
3. **Environment Variables** - Build a configuration validator

### 🏗️ Intermediate Extensions
1. **Database Integration** - Connect to a specific database system
2. **Cloud Service** - Integrate with a cloud provider's API
3. **Message Queue** - Process messages from RabbitMQ, Kafka, etc.

### 🧙‍♂️ Advanced Extensions
1. **AI/ML Integration** - Build intelligent data processing
2. **Multi-Protocol** - Support multiple protocols in one extension
3. **Performance Critical** - Optimize for high-throughput scenarios

---

## 📞 Getting Help

### 🆘 Troubleshooting Extensions

1. **Enable Debug Logging**:
```bash
export DEBUG=true
./compass-compute compute my-service --verbose
```

2. **Test Extension Independently**:
```go
func TestMyExtension(t *testing.T) {
    fact := &services.Fact{
        Type: "myextension",
        // ... configure fact
    }
    
    ext := &MyExtension{}
    err := ext.Process(context.Background(), fact)
    assert.NoError(t, err)
}
```


---

<div align="center">

**Ready to extend compass-compute?** 🚀

[🏠 Back to Main](../README.md) | [📊 Facts Guide](facts-and-metrics.md) | [💻 Contributing](contributing.md)

**[⭐ Star this repo](https://github.com/motain/compass-compute)** if this guide helped you build awesome extensions!

</div>