# 📖 compass-compute Documentation

> **Everything you need to know about compass-compute**

## Quick Navigation

| 🎯 What do you want to do? | 📚 Guide | ⏱️ Time |
|----------------------------|----------|--------|
| **Get it running locally** | [Setup Guide](setup.md) | 5 min |
| **Create my first metric** | [Facts & Metrics](facts-guide.md) | 10 min |
| **Add custom data sources** | [Extensions Guide](extensions.md) | 15 min |
| **Fix broken things**      | [Debugging Guide](debugging.md) | 8 min |
| **Architecture**           | [architecture.md](architecture.md) | 8 min |


## Understanding compass-compute

### What It Does
compass-compute evaluates component metrics using a **fact-based engine**:

1. **Extracts** data from sources (GitHub, APIs, Prometheus)
2. **Validates** data meets criteria
3. **Aggregates** results into final scores
4. **Submits** to Atlassian Compass

### Key Concepts
- **Facts**: Individual data operations (extract, validate, aggregate)
- **Metrics**: Collections of facts that produce scores
- **Sources**: Where data comes from (github, api, prometheus)
- **Rules**: How data is processed (jsonpath, regex, count)

### Architecture Overview
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│      CLI        │────│  Facts Engine    │────│  Data Sources   │
│   (compute)     │    │                  │    │                 │
└─────────────────┘    │  • Extract       │    │  • GitHub       │
                       │  • Validate      │    │  • Prometheus   │
                       │  • Aggregate     │    │  • APIs         │
                       └──────────────────┘    │  • Custom...    │
                                │               └─────────────────┘
                       ┌──────────────────┐
                       │ Atlassian Compass │
                       │   (Metrics API)   │
                       └──────────────────┘
```

## Code Structure

```
compass-compute/
├── cmd/                    # CLI commands
│   ├── main.go            # Entry point
│   └── compute.go         # Main compute command
├── internal/
│   ├── services/          # External integrations
│   │   ├── compass.go     # Compass API client
│   │   ├── prometheus_service.go  # Prometheus client
│   │   ├── github.go      # GitHub operations  
│   │   ├── metrics.go     # YAML metric parser
│   │   └── models.go      # Data structures
│   ├── facts/             # Facts evaluation engine
│   │   ├── evaluator.go   # Main evaluation logic
│   │   ├── extractor.go   # Data extraction (sources)
│   │   ├── processor.go   # Fact processing (types)
│   │   ├── appliers.go    # Rule application
│   │   └── helpers.go     # Utilities
│   └── compute/           # Business logic orchestration
│       └── compute.go     # Main workflow
└── docs/                  # Documentation
```

## Extension Points

### 🔧 Where to Add Custom Logic

| What | Where | Function |
|------|-------|----------|
| **Custom fact types** | `internal/facts/evaluator.go` | `processCustom()` |
| **Custom data sources** | `internal/facts/extractor.go` | `extractCustom()` |
| **Custom processing rules** | `internal/facts/appliers.go` | `applyCustomRule()` |
| **New integrations** | `internal/services/` | New service files |

### 🎯 Common Extensions

- **Kubernetes**: Pod counts, resource usage
- **Databases**: Query results, health checks
- **Security**: Vulnerability scans, compliance
- **AI/ML**: Code analysis, predictions
- **Business**: Custom calculations, KPIs

## Quick Reference

### Environment Variables
```bash
# Required
COMPASS_API_TOKEN="your-compass-token"
COMPASS_CLOUD_ID="your-cloud-id"
GITHUB_TOKEN="your-github-token"

# For Prometheus
AWS_REGION="us-east-1"
PROMETHEUS_WORKSPACE_URL="https://aps-workspaces..."

# Optional
METRIC_DIR="/custom/metrics/path"
```

### Common Commands
```bash
# Basic usage
./compass-compute compute my-service

# With debugging
./compass-compute compute my-service --verbose

# Multiple services
./compass-compute compute service-a,service-b

# Docker
docker run --env-file .env compass-compute:latest compute my-service
```

### Metric YAML Template
```yaml
apiVersion: v1
kind: Metric
metadata:
  name: my-metric
  componentType: ["service"]
  facts:
    - id: extract-data
      type: extract
      source: github
      repo: ${Metadata.Name}
      filePath: package.json
      rule: jsonpath
      jsonPath: ".version"
      
    - id: validate-version
      type: validate
      dependsOn: [extract-data]
      rule: regex_match
      pattern: "^v[0-9]+\\.[0-9]+\\.[0-9]+$"
```

## Troubleshooting Quick Fixes

| Problem | Quick Fix |
|---------|-----------|
| "Component not found" | Check component name exists in Compass UI |
| "Repository not found" | Verify GitHub token has repo access |
| "No facts found" | Check componentType matches your service |
| "AWS credentials" | Configure AWS CLI or check region |
| "JSONPath failed" | Test expression at jsonpath.com |

## Contributing

### Development Setup
```bash
git clone <repo-url>
cd compass-compute
make setup    # Install tools
make test     # Run tests
make build    # Build binary
```

### Adding Features
1. **Small changes**: Edit relevant files directly
2. **New fact types**: Follow extension patterns
3. **New integrations**: Add to services layer
4. **Bug fixes**: Add tests first

### Testing
```bash
make test                    # Unit tests
make integration-test        # Integration tests
./compass-compute compute test-service --verbose  # Manual testing
```

## Support

### Getting Help
1. **Check the guides** - Most issues are covered
2. **Use verbose mode** - `--verbose` shows what's happening
3. **Test components** - Verify each piece works
4. **Check examples** - Look at existing metric definitions

### Reporting Issues
Include:
- Full verbose output
- Environment setup (redacted)
- Component name and type
- Metric definitions used
- Expected vs actual behavior

---

**Ready to get started?** → [Setup Guide](setup.md)

**Need to create metrics?** → [Facts & Metrics Guide](facts-guide.md)

**Want to extend functionality?** → [Extensions Guide](extensions.md)

**Something broken?** → [Debugging Guide](debugging.md)

**Want to understand the architecture?** → [Architecture Overview](architecture.md)