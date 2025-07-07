# 🐛 Debugging Guide

> **Troubleshoot compass-compute issues quickly and effectively**

## Quick Debugging Checklist

```bash
# 1. Run with verbose output
./compass-compute compute my-service --verbose

# 2. Check environment variables
env | grep -E "(COMPASS|GITHUB|AWS|PROMETHEUS)"

# 3. Test component exists in Compass
# Visit Compass UI and verify component name

# 4. Validate metric YAML files
yamllint metrics/*.yaml

# 5. Test connectivity
curl -H "Authorization: Basic $COMPASS_API_TOKEN" \
  "https://onefootball.atlassian.net/gateway/api/graphql"
```

## Common Issues

### "Component not found"

**Error:** `failed to get component 'my-service': component not found`

**Causes & Fixes:**
```bash
# 1. Wrong component name
./compass-compute compute my-service --verbose
# Check: Does component exist in Compass UI?

# 2. Component name prefix issue
# Compass expects "svc-" prefix for services
# Code automatically adds it, so use: "my-service" not "svc-my-service"

# 3. Wrong cloud ID
echo $COMPASS_CLOUD_ID
# Verify this matches your Compass instance
```

### "Failed to clone repository"

**Error:** `git clone failed: repository not found`

**Causes & Fixes:**
```bash
# 1. GitHub token permissions
# Token needs 'repo' access for private repos

# 2. Repository doesn't exist
# Check: https://github.com/motain/my-service exists

# 3. Wrong organization
# Default org is "motain" - change in services/config.go if needed
```

### "No facts found for metric"

**Error:** `no facts found for metric 'deployment-frequency' and type 'service'`

**Causes & Fixes:**
```bash
# 1. Metric definition missing
ls -la ./repos/of-catalog/config/grading-system/
# Should contain .yaml files

# 2. Component type mismatch
# Check metric YAML:
# metadata:
#   componentType: ["service"]  # Must match your component

# 3. Custom metric directory
export METRIC_DIR="/path/to/your/metrics"
./compass-compute compute my-service --verbose
```

### "AWS credentials not found"

**Error:** `failed to get AWS credentials`

**Causes & Fixes:**
```bash
# 1. AWS credentials not configured
aws configure list
# Should show access key, secret key, region

# 2. AWS role not accessible
aws sts get-caller-identity
# Should return your identity

# 3. Prometheus workspace URL wrong
echo $PROMETHEUS_WORKSPACE_URL
# Should be: https://aps-workspaces.region.amazonaws.com/workspaces/ws-xxxxx/
```

## Debugging by Component

### CLI Layer (`cmd/`)

**Issue:** Command not recognized or flags not working

```bash
# Check command structure
./compass-compute --help
./compass-compute compute --help

# Verify binary is built correctly
make build
file ./compass-compute  # Should show executable
```

### Services Layer (`internal/services/`)

**Issue:** API calls failing

```bash
# Test Compass API manually
curl -X POST \
  -H "Authorization: Basic $COMPASS_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __typename }", "variables": {}}' \
  "https://onefootball.atlassian.net/gateway/api/graphql"

# Test GitHub API
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
  "https://api.github.com/repos/motain/my-service"

# Test Prometheus
curl -H "Authorization: AWS4-HMAC-SHA256 ..." \
  "$PROMETHEUS_WORKSPACE_URL/api/v1/query?query=up"
```

### Facts Engine (`internal/facts/`)

**Issue:** Facts not evaluating correctly

```bash
# Enable debug logging in evaluator.go
# Add: fmt.Printf("Processing fact: %+v\n", fact)

# Check fact dependencies
# Facts with dependsOn must wait for dependencies to complete

# Validate JSONPath expressions
# Use online JSONPath evaluator: https://jsonpath.com/
```

### Data Processing Issues

**Issue:** JSONPath not working

```yaml
# Test JSONPath step by step
- id: debug-json
  type: extract
  source: api
  uri: "https://api.github.com/repos/owner/repo"
  rule: jsonpath
  jsonPath: "."  # Start with full response
  
# Then narrow down:
# jsonPath: ".stargazers_count"
# jsonPath: ".owner.login"
```

**Issue:** Regex not matching

```yaml
# Test regex patterns
- id: debug-regex
  type: validate
  dependsOn: [get-data]
  rule: regex_match
  pattern: ".*"  # Start with match-all pattern
  
# Then narrow down:
# pattern: "^v[0-9]+\\.[0-9]+\\.[0-9]+$"
```

## Debugging Tools

### 1. Verbose Mode

```bash
# See everything that's happening
./compass-compute compute my-service --verbose

# Example output shows:
# - Component lookup
# - Repository cloning
# - Metric discovery
# - Fact evaluation
# - API submissions
```

### 2. Environment Debugging

```bash
# Check all environment variables
env | grep -E "(COMPASS|GITHUB|AWS|PROMETHEUS|METRIC)"

# Test specific variables
echo "Compass Token: ${COMPASS_API_TOKEN:0:10}..."
echo "GitHub Token: ${GITHUB_TOKEN:0:10}..."
echo "AWS Region: $AWS_REGION"
```

### 3. Network Debugging

```bash
# Test Compass connectivity
curl -v "https://onefootball.atlassian.net/gateway/api/graphql"

# Test GitHub connectivity  
curl -v "https://api.github.com"

# Test Prometheus connectivity
curl -v "$PROMETHEUS_WORKSPACE_URL"
```

### 4. File System Debugging

```bash
# Check cloned repositories
ls -la ./repos/
ls -la ./repos/my-service/
ls -la ./repos/of-catalog/config/grading-system/

# Check metric files
find ./repos -name "*.yaml" -type f
cat ./repos/of-catalog/config/grading-system/deployment-frequency.yaml
```

## Debugging Specific Components

### Prometheus Integration

```bash
# Check AWS credentials
aws sts get-caller-identity

# Test Prometheus query manually
# Use AWS CLI to get temporary credentials
aws sts assume-role --role-arn $AWS_ROLE --role-session-name debug-session

# Check workspace URL format
echo $PROMETHEUS_WORKSPACE_URL
# Should end with /workspaces/ws-xxxxx/ (note trailing slash)
```

### GitHub Integration

```bash
# Test token permissions
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
  "https://api.github.com/user"

# Check repository access
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
  "https://api.github.com/repos/motain/my-service"

# Test cloning manually
git clone https://$GITHUB_TOKEN@github.com/motain/my-service.git
```

### Compass Integration

```bash
# Decode your API token (it's base64 encoded)
echo $COMPASS_API_TOKEN | base64 -d

# Test GraphQL query
curl -X POST \
  -H "Authorization: Basic $COMPASS_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query getComponent($cloudId: ID!, $slug: String!) { compass { componentByReference(reference: {slug: {slug: $slug, cloudId: $cloudId}}) { ... on CompassComponent { id name type } } } }",
    "variables": {"cloudId": "'$COMPASS_CLOUD_ID'", "slug": "svc-my-service"}
  }' \
  "https://onefootball.atlassian.net/gateway/api/graphql"
```

## Common Code Issues

### Memory Leaks

```bash
# Monitor memory usage
go build -race ./cmd
./compass-compute compute my-service

# Use pprof for profiling
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Concurrent Access Issues

```bash
# Build with race detector
go build -race ./cmd
./compass-compute compute my-service

# Look for race condition warnings
```

### JSON Parsing Issues

```go
// Add debug logging in appliers.go
func (fe *FactEvaluator) applyJSONPath(jsonPath interface{}, data []byte) (interface{}, error) {
    fmt.Printf("JSONPath input data: %s\n", string(data))
    fmt.Printf("JSONPath expression: %v\n", jsonPath)
    
    // ... rest of function
}
```

## Performance Debugging

### Slow Execution

```bash
# Time individual operations
time ./compass-compute compute my-service

# Profile CPU usage
go tool pprof ./compass-compute profile.prof
```

### Network Timeouts

```go
// Increase timeouts in compass.go
client: &http.Client{Timeout: 60 * time.Second},  // Was 30

// Add retry logic
for i := 0; i < 3; i++ {
    resp, err := client.Do(req)
    if err == nil {
        return resp, nil
    }
    time.Sleep(time.Duration(i+1) * time.Second)
}
```

## Advanced Debugging

### Add Debug Logging

```go
// In internal/facts/evaluator.go
func (fe *FactEvaluator) processFact(ctx context.Context, fact *services.Fact, factMap map[string]*services.Fact) error {
    if os.Getenv("DEBUG") == "true" {
        fmt.Printf("Processing fact: %s (type: %s)\n", fact.ID, fact.Type)
    }
    // ... rest of function
}
```

### Inspect Fact Results

```go
// In internal/facts/evaluator.go - after fact processing
func EvaluateMetric(facts []services.Fact, componentName string) (interface{}, error) {
    // ... existing code ...
    
    if os.Getenv("DEBUG") == "true" {
        for _, fact := range facts {
            fmt.Printf("Fact %s result: %+v\n", fact.ID, fact.Result)
        }
    }
    
    // ... rest of function
}
```

### Mock External Services

```go
// For testing, replace HTTP client
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// Use mock client in tests
type MockHTTPClient struct {
    Response *http.Response
    Error    error
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.Response, m.Error
}
```

## Emergency Debugging

### Complete System Check

```bash
#!/bin/bash
# save as debug.sh

echo "=== Environment Check ==="
env | grep -E "(COMPASS|GITHUB|AWS|PROMETHEUS)" | head -10

echo "=== Connectivity Check ==="
curl -s -o /dev/null -w "%{http_code}" "https://api.github.com"
curl -s -o /dev/null -w "%{http_code}" "https://onefootball.atlassian.net"

echo "=== Binary Check ==="
./compass-compute --version 2>/dev/null || echo "Binary not found or broken"

echo "=== Repository Check ==="
ls -la ./repos/ 2>/dev/null || echo "No repos directory"

echo "=== Metric Files Check ==="
find ./repos -name "*.yaml" -type f | head -5

echo "=== Test Component Lookup ==="
./compass-compute compute test-service --verbose 2>&1 | head -20
```

Run this script when everything seems broken:
```bash
chmod +x debug.sh
./debug.sh
```

## Getting Help

### Before Asking for Help

1. Run with `--verbose` and save the output
2. Check environment variables are set correctly
3. Verify component exists in Compass UI
4. Test network connectivity manually
5. Check metric YAML files exist and are valid

### Useful Information to Include

- Full verbose output
- Environment variable values (redacted)
- Component name and type
- Metric definitions being used
- Error messages (full stack traces)

### Self-Help Resources

- Use existing metric definitions as templates
- Check the codebase for similar implementations
- Test with simpler configurations first
- Use online tools for JSONPath and regex testing

Ready to contribute? Check out the main [README](../README.md) for contribution guidelines.