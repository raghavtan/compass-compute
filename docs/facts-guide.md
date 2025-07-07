# 📊 Facts & Metrics Guide

> **Create powerful metrics with the fact-based evaluation engine**

## Core Concepts

### What's a Fact?
A fact is a single data operation:
- **Extract**: Get data from a source (GitHub, API, Prometheus)
- **Validate**: Check if data meets criteria
- **Aggregate**: Combine multiple results

### What's a Metric?
A metric combines facts to produce a final score for Compass.

## Quick Example

```yaml
# deployment-frequency.yaml
apiVersion: v1
kind: Metric
metadata:
  name: deployment-frequency
  componentType: ["service"]
  facts:
    - id: get-deployments
      type: extract
      source: api
      uri: "https://api.github.com/repos/org/${Metadata.Name}/deployments"
      rule: jsonpath
      jsonPath: "length"
      
    - id: validate-frequency
      type: validate
      dependsOn: [get-deployments]
      rule: regex_match
      pattern: "^[4-9]|[1-9][0-9]+$"  # At least 4 deployments
```

## Fact Types

### 1. Extract Facts

**From GitHub:**
```yaml
- id: get-dockerfile
  type: extract
  source: github
  repo: ${Metadata.Name}
  filePath: Dockerfile
  rule: search
  searchString: "FROM node:"
```

**From APIs:**
```yaml
- id: get-health-status
  type: extract
  source: api
  uri: "https://health-api.com/status/${Metadata.Name}"
  auth:
    header: "Authorization"
    tokenVar: "API_TOKEN"
  rule: jsonpath
  jsonPath: ".status"
```

**From Prometheus:**
```yaml
- id: get-error-rate
  type: extract
  source: prometheus
  prometheusQuery: 'rate(http_errors_total{service="${Metadata.Name}"}[5m])'
  rule: instant
```

### 2. Validate Facts

**Regex Validation:**
```yaml
- id: check-version-format
  type: validate
  dependsOn: [get-version]
  rule: regex_match
  pattern: "^v[0-9]+\\.[0-9]+\\.[0-9]+$"
```

**Dependency Matching:**
```yaml
- id: versions-match
  type: validate
  dependsOn: [frontend-version, backend-version]
  rule: deps_match
```

### 3. Aggregate Facts

**Count Items:**
```yaml
- id: total-services
  type: aggregate
  dependsOn: [list-services, list-workers]
  method: count
```

**Logical Operations:**
```yaml
- id: all-checks-pass
  type: aggregate
  dependsOn: [test-pass, lint-pass, security-pass]
  method: and
```

## Processing Rules

### JSONPath Rules
```yaml
rule: jsonpath
jsonPath: ".items[0].name"           # First item
jsonPath: ".items[*].status"         # All statuses  
jsonPath: ".items | length"          # Count items
jsonPath: ".items[?(@.active)]"      # Filter active
```

### Search Rules
```yaml
rule: search
searchString: "TODO"                 # Find TODO comments
searchString: "FROM node:"           # Find Node.js usage
```

### Built-in Rules
```yaml
rule: notempty                       # Check if file exists
rule: count                          # Count files/items
rule: instant                        # Prometheus instant query
rule: range                          # Prometheus range query
```

## Data Sources

### GitHub Source
```yaml
source: github
repo: ${Metadata.Name}               # Component repository
filePath: "package.json"             # Single file
filePath: "**/*.js"                  # File pattern
rule: search                         # Search in files
searchString: "express"              # What to search for
```

### API Source
```yaml
source: api
uri: "https://api.example.com/data"
auth:                                # Optional authentication
  header: "Authorization"
  tokenVar: "API_TOKEN"
rule: jsonpath
jsonPath: ".data.count"
```

### Prometheus Source
```yaml
source: prometheus
prometheusQuery: 'up{service="${Metadata.Name}"}'
rule: instant                        # or 'range'
```

## Real Examples

### Test Coverage Metric
```yaml
apiVersion: v1
kind: Metric
metadata:
  name: test-coverage
  componentType: ["service"]
  facts:
    - id: get-coverage
      type: extract
      source: github
      repo: ${Metadata.Name}
      filePath: coverage.json
      rule: jsonpath
      jsonPath: ".total.lines.pct"
      
    - id: validate-coverage
      type: validate
      dependsOn: [get-coverage]
      rule: regex_match
      pattern: "^([8-9][0-9]|100)$"   # 80%+ coverage
```

### SLA Uptime Metric
```yaml
apiVersion: v1
kind: Metric
metadata:
  name: sla-uptime
  componentType: ["service"]
  facts:
    - id: get-uptime
      type: extract
      source: prometheus
      prometheusQuery: 'avg_over_time(up{service="${Metadata.Name}"}[30d])'
      rule: instant
      
    - id: validate-sla
      type: validate
      dependsOn: [get-uptime]
      rule: regex_match
      pattern: "^0\\.(9[5-9]|[1-9][0-9]).*|^1\\.0+$"  # 95%+ uptime
```

## Dependencies

Facts can depend on other facts:

```yaml
facts:
  - id: step1
    type: extract
    # ... get some data
    
  - id: step2
    type: extract
    dependsOn: [step1]      # Wait for step1
    # ... use step1 result
    
  - id: final
    type: aggregate
    dependsOn: [step1, step2]  # Wait for both
    method: and
```

## Tips

1. **Start simple** - Begin with a single extract fact
2. **Use verbose mode** - `--verbose` shows each step
3. **Test JSONPath** - Use online JSONPath evaluators
4. **Check data first** - Verify API responses manually
5. **Handle missing data** - Not all files/APIs will exist

## Debugging

```bash
# See what's happening
./compass-compute compute my-service --verbose

# Check metric discovery
./compass-compute compute my-service --dry-run

# Validate YAML syntax
yamllint metrics/my-metric.yaml
```

Ready to create custom data sources? See the [Extensions Guide](extensions.md).