# 📊 Facts and Metrics: The Complete Guide

> **Master the art of metric definition and fact evaluation**

This guide will transform you from a metrics novice to a compass-compute expert. Whether you're creating your first metric or building complex multi-source evaluations, we've got you covered.

---

## 🎯 What You'll Learn

- **[Quick Start](#-quick-start)** - Get your first metric running in 5 minutes
- **[Core Concepts](#-core-concepts)** - Understanding facts, metrics, and how they work together
- **[Fact Types Deep Dive](#-fact-types-deep-dive)** - Master extract, validate, and aggregate operations
- **[Data Sources](#-data-sources)** - Connect to GitHub, Prometheus, APIs, and more
- **[Advanced Patterns](#-advanced-patterns)** - Dependencies, transformations, and complex workflows
- **[Real-World Examples](#-real-world-examples)** - Production-ready metric configurations
- **[Troubleshooting](#-troubleshooting)** - Debug and fix common issues

---

## 🚀 Quick Start

### Your First Metric in 5 Minutes

Let's create a simple metric that counts the number of files in your repository:

```yaml
apiVersion: v1
kind: Metric
metadata:
  name: file-count
  componentType: ["service"]
  facts:
    - id: count-files
      type: extract
      source: github
      repo: ${Metadata.Name}
      filePath: "."
      rule: count
```

**That's it!** This metric will:
1. Clone your component's repository
2. Count all files in the root directory
3. Submit the count to Compass

### Running Your Metric

```bash
# Test with your service
./compass-compute compute my-service --verbose
```

---

## 🧠 Core Concepts

### What is a Fact?

A **fact** is a single piece of data or computation in your metric evaluation. Think of facts as building blocks:

```yaml
facts:
  - id: get-package-json        # ← Fact 1: Extract data
    type: extract
    source: github
    filePath: package.json
    
  - id: validate-node-version   # ← Fact 2: Validate data
    type: validate
    dependsOn: [get-package-json]
    rule: regex_match
    pattern: '"node": ">=18"'
```

### What is a Metric?

A **metric** is a complete evaluation workflow made up of one or more facts:

```yaml
apiVersion: v1
kind: Metric
metadata:
  name: node-version-compliance
  componentType: ["service", "library"]
  facts:
    # Multiple facts work together to create the final metric value
```

### The Magic of Dependencies

Facts can depend on each other, creating powerful data pipelines:

```yaml
facts:
  - id: step1
    type: extract
    # ... extract some data

  - id: step2
    type: extract
    dependsOn: [step1]  # ← This fact waits for step1 to complete
    # ... use data from step1

  - id: final-result
    type: aggregate
    dependsOn: [step1, step2]  # ← This fact combines results
    method: and
```

---

## 🔍 Fact Types Deep Dive

### 📥 Extract Facts: Getting Data

Extract facts pull data from various sources. They're your data collection workhorses.

#### Basic Structure
```yaml
- id: my-extract-fact
  type: extract
  source: github        # Where to get data from
  repo: my-service      # Repository name
  filePath: README.md   # What file to read
  rule: jsonpath        # How to process the data (optional)
  jsonPath: ".title"    # Processing parameters (optional)
```

#### Common Extract Patterns

**Reading Configuration Files:**
```yaml
- id: get-docker-config
  type: extract
  source: github
  repo: ${Metadata.Name}
  filePath: Dockerfile
  rule: search
  searchString: "FROM node:"
```

**API Data Collection:**
```yaml
- id: get-service-health
  type: extract
  source: api
  uri: https://api.example.com/health/${Metadata.Name}
  rule: jsonpath
  jsonPath: ".status"
  auth:
    header: "Authorization"
    tokenVar: "API_TOKEN"
```

**Prometheus Metrics:**
```yaml
- id: get-error-rate
  type: extract
  source: prometheus
  prometheusQuery: 'rate(http_requests_total{service="${Metadata.Name}",status=~"5.."}[5m])'
  rule: instant
```

### ✅ Validate Facts: Checking Data Quality

Validate facts ensure your data meets specific criteria.

#### Regex Validation
```yaml
- id: check-node-version
  type: validate
  dependsOn: [get-package-json]
  rule: regex_match
  pattern: '"node": ">=18"'
```

#### Dependency Matching
```yaml
- id: ensure-versions-match
  type: validate
  dependsOn: [frontend-version, backend-version]
  rule: deps_match
```

#### Uniqueness Validation
```yaml
- id: check-unique-ports
  type: validate
  dependsOn: [get-all-ports]
  rule: unique
```

### 📊 Aggregate Facts: Combining Results

Aggregate facts combine multiple results into a single value.

#### Counting
```yaml
- id: total-services
  type: aggregate
  dependsOn: [list-services, list-libraries]
  method: count
```

#### Mathematical Operations
```yaml
- id: total-coverage
  type: aggregate
  dependsOn: [frontend-coverage, backend-coverage]
  method: sum
```

#### Logical Operations
```yaml
- id: all-checks-pass
  type: aggregate
  dependsOn: [test-pass, build-pass, deploy-pass]
  method: and

- id: any-alert-active
  type: aggregate
  dependsOn: [critical-alerts, warning-alerts]
  method: or
```

---

## 🌐 Data Sources

### 🐙 GitHub Source

Perfect for code-related metrics.

**Basic File Reading:**
```yaml
source: github
repo: ${Metadata.Name}
filePath: package.json
```

**Repository Search:**
```yaml
source: github
repo: ${Metadata.Name}
rule: search
searchString: "TODO"
```

**Multi-File Processing:**
```yaml
source: github
repo: ${Metadata.Name}
filePath: "**/*.js"
rule: count
```

### 📈 Prometheus Source

For runtime and operational metrics.

**Instant Queries:**
```yaml
source: prometheus
prometheusQuery: 'up{service="${Metadata.Name}"}'
rule: instant
```

**Range Queries:**
```yaml
source: prometheus
prometheusQuery: 'rate(http_requests_total{service="${Metadata.Name}"}[5m])'
rule: range
```

**Complex Queries:**
```yaml
source: prometheus
prometheusQuery: |
  (
    sum(rate(http_requests_total{service="${Metadata.Name}",status=~"2.."}[5m])) /
    sum(rate(http_requests_total{service="${Metadata.Name}"}[5m]))
  ) * 100
rule: instant
```

### 🌍 API Source

For external data and integrations.

**Simple GET Requests:**
```yaml
source: api
uri: https://api.github.com/repos/owner/${Metadata.Name}
rule: jsonpath
jsonPath: ".stargazers_count"
```

**Authenticated Requests:**
```yaml
source: api
uri: https://api.example.com/metrics/${Metadata.Name}
auth:
  header: "Authorization"
  tokenVar: "SERVICE_TOKEN"
rule: jsonpath
jsonPath: ".metrics.availability"
```

**Dynamic URLs with Dependencies:**
```yaml
source: api
uri: https://api.example.com/alerts/:alert_id/recipients
dependsOn: [get-alert-id]
rule: jsonpath
jsonPath: ".recipients | length"
```

---

## 🎨 Advanced Patterns

### 🔗 Dependency Chains

Create complex data processing pipelines:

```yaml
facts:
  # Step 1: Get all service configurations
  - id: get-services
    type: extract
    source: github
    repo: service-catalog
    filePath: services.yaml
    rule: jsonpath
    jsonPath: ".services[*].name"

  # Step 2: For each service, get its health status
  - id: check-health
    type: extract
    source: api
    uri: https://health-check.com/status/${Metadata.Name}
    dependsOn: [get-services]
    rule: jsonpath
    jsonPath: ".status"

  # Step 3: Aggregate all health statuses
  - id: overall-health
    type: aggregate
    dependsOn: [check-health]
    method: and
```

### 🔄 Data Transformations

Process and transform data as it flows through your pipeline:

```yaml
facts:
  # Extract raw deployment data
  - id: get-deployments
    type: extract
    source: api
    uri: https://ci-cd.com/deployments/${Metadata.Name}
    rule: jsonpath
    jsonPath: ".deployments[*].timestamp"

  # Transform timestamps to deployment frequency
  - id: calculate-frequency
    type: aggregate
    dependsOn: [get-deployments]
    method: count  # Count number of deployments

  # Validate deployment frequency meets standards
  - id: validate-frequency
    type: validate
    dependsOn: [calculate-frequency]
    rule: regex_match
    pattern: "^[1-9][0-9]*$"  # At least 1 deployment
```

### 🎯 Conditional Processing

Use validation facts to create conditional logic:

```yaml
facts:
  # Check if service uses Docker
  - id: uses-docker
    type: extract
    source: github
    repo: ${Metadata.Name}
    filePath: Dockerfile
    rule: notempty

  # Only check Docker security if service uses Docker
  - id: docker-security
    type: extract
    source: github
    repo: ${Metadata.Name}
    filePath: Dockerfile
    dependsOn: [uses-docker]
    rule: search
    searchString: "USER"

  # Final score considers Docker usage
  - id: security-score
    type: aggregate
    dependsOn: [uses-docker, docker-security]
    method: and
```

---

## 🌟 Real-World Examples

### 📦 Deployment Frequency Metric

Measures how often a service is deployed:

```yaml
apiVersion: v1
kind: Metric
metadata:
  name: deployment-frequency
  componentType: ["service"]
  facts:
    - id: get-github-deployments
      type: extract
      source: api
      uri: https://api.github.com/repos/myorg/${Metadata.Name}/deployments
      auth:
        header: "Authorization"
        tokenVar: "GITHUB_TOKEN"
      rule: jsonpath
      jsonPath: ".[?(@.created_at >= '2024-01-01')].created_at"

    - id: count-deployments
      type: aggregate
      dependsOn: [get-github-deployments]
      method: count

    - id: calculate-weekly-frequency
      type: validate
      dependsOn: [count-deployments]
      rule: regex_match
      pattern: "^[4-9]|[1-9][0-9]+$"  # At least 4 deployments per period
```

### 🧪 Test Coverage Metric

Combines frontend and backend test coverage:

```yaml
apiVersion: v1
kind: Metric
metadata:
  name: overall-test-coverage
  componentType: ["service"]
  facts:
    - id: get-frontend-coverage
      type: extract
      source: github
      repo: ${Metadata.Name}
      filePath: frontend/coverage.json
      rule: jsonpath
      jsonPath: ".total.lines.pct"

    - id: get-backend-coverage
      type: extract
      source: github
      repo: ${Metadata.Name}
      filePath: backend/coverage.xml
      rule: regex_match
      pattern: 'line-rate="([0-9.]+)"'

    - id: average-coverage
      type: aggregate
      dependsOn: [get-frontend-coverage, get-backend-coverage]
      method: sum  # Will be divided by count automatically

    - id: validate-coverage
      type: validate
      dependsOn: [average-coverage]
      rule: regex_match
      pattern: "^([8-9][0-9]|100)$"  # At least 80% coverage
```

### 🚨 SLA Compliance Metric

Checks service level agreement compliance:

```yaml
apiVersion: v1
kind: Metric
metadata:
  name: sla-compliance
  componentType: ["service"]
  facts:
    - id: get-uptime
      type: extract
      source: prometheus
      prometheusQuery: 'avg_over_time(up{service="${Metadata.Name}"}[30d])'
      rule: instant

    - id: get-error-rate
      type: extract
      source: prometheus
      prometheusQuery: |
        (
          sum(rate(http_requests_total{service="${Metadata.Name}",status=~"5.."}[30d])) /
          sum(rate(http_requests_total{service="${Metadata.Name}"}[30d]))
        ) * 100
      rule: instant

    - id: get-response-time
      type: extract
      source: prometheus
      prometheusQuery: 'histogram_quantile(0.95, http_request_duration_seconds{service="${Metadata.Name}"})'
      rule: instant

    - id: check-uptime-sla
      type: validate
      dependsOn: [get-uptime]
      rule: regex_match
      pattern: "^0\\.9[5-9]|1\\.0+$"  # 95%+ uptime

    - id: check-error-rate-sla
      type: validate
      dependsOn: [get-error-rate]
      rule: regex_match
      pattern: "^[0-4]\\..*|^5\\.0+$"  # <5% error rate

    - id: check-response-time-sla
      type: validate
      dependsOn: [get-response-time]
      rule: regex_match
      pattern: "^0\\.[0-4].*"  # <500ms response time

    - id: overall-sla-compliance
      type: aggregate
      dependsOn: [check-uptime-sla, check-error-rate-sla, check-response-time-sla]
      method: and
```

---

## 🔧 Advanced Configuration

### 🎛️ Processing Rules Reference

#### JSONPath Rules
```yaml
rule: jsonpath
jsonPath: ".items[0].name"          # Get first item name
jsonPath: ".items[*].status"        # Get all statuses
jsonPath: ".items | length"         # Count items
jsonPath: ".items[?(@.active)]"     # Filter active items
```

#### Search Rules
```yaml
rule: search
searchString: "TODO"                # Search for TODO comments
searchString: "FROM node:"          # Search for Node.js base image
```

#### Validation Rules
```yaml
rule: regex_match
pattern: "^v[0-9]+\\.[0-9]+\\.[0-9]+$"  # Semantic version
pattern: "^[1-9][0-9]*$"                # Positive number
pattern: "^(true|false)$"               # Boolean value
```

### 🔗 Placeholder Substitution

Use dynamic values in your configurations:

```yaml
# Component name substitution
repo: ${Metadata.Name}
uri: https://api.example.com/services/${Metadata.Name}
prometheusQuery: 'up{service="${Metadata.Name}"}'

# Environment-based substitution (custom implementation)
uri: ${ENV.API_BASE_URL}/services/${Metadata.Name}
```

### 🏷️ Component Type Targeting

Target specific component types:

```yaml
metadata:
  name: docker-security
  componentType: ["service"]        # Only for services

metadata:
  name: npm-audit
  componentType: ["service", "library"]  # For services and libraries

metadata:
  name: database-metrics
  componentType: ["database"]       # Only for databases
```

---

## 🔧 Troubleshooting

### 🐛 Common Issues and Solutions

#### "No facts found for metric"
**Problem:** Metric not found for your component type.
**Solution:**
1. Check component type matches metric definition
2. Verify metric files are in the correct path
3. Check YAML syntax

```bash
# Debug metric discovery
./compass-compute compute my-service --verbose
```

#### "Failed to process fact: extraction failed"
**Problem:** Data source is unreachable or returns unexpected data.
**Solution:**
1. Verify source URLs and authentication
2. Check network connectivity
3. Test API endpoints manually

```bash
# Test API endpoint
curl -H "Authorization: Bearer $TOKEN" "https://api.example.com/health/my-service"
```

#### "Circular dependency detected"
**Problem:** Facts depend on each other in a loop.
**Solution:**
1. Review dependency chains
2. Ensure facts have proper order
3. Remove circular references

```yaml
# ❌ Circular dependency
facts:
  - id: fact-a
    dependsOn: [fact-b]
  - id: fact-b
    dependsOn: [fact-a]

# ✅ Proper dependency chain
facts:
  - id: fact-a
  - id: fact-b
    dependsOn: [fact-a]
```

### 📊 Debugging Techniques

#### Use Verbose Mode
```bash
./compass-compute compute my-service --verbose
```

#### Test Individual Components
```bash
# Test metric discovery
./compass-compute compute my-service --dry-run

# Test specific metric
./compass-compute compute my-service --metric deployment-frequency
```

#### Validate YAML Syntax
```bash
# Check YAML files
yamllint metrics/

# Validate specific file
yaml-validator metrics/deployment-frequency.yaml
```

---

## 📈 Best Practices

### 🎯 Metric Design

1. **Keep metrics focused** - One metric should measure one thing
2. **Use meaningful names** - `deployment-frequency` not `metric-1`
3. **Document complex logic** - Add comments explaining complex JSONPath or regex
4. **Test with real data** - Always test with actual component data

### 🔧 Fact Organization

1. **Logical ordering** - Organize facts in dependency order
2. **Descriptive IDs** - Use clear, descriptive fact IDs
3. **Error handling** - Consider what happens when data is missing
4. **Performance** - Minimize API calls and heavy operations

### 🛡️ Security

1. **Environment variables** - Store sensitive data in environment variables
2. **Minimal permissions** - Use least-privilege access for tokens
3. **Audit trails** - Log metric evaluations for compliance
4. **Data privacy** - Be mindful of sensitive data in metrics

---