# 🛠️ Setup Guide

> **Get compass-compute running locally in 5 minutes**

## Prerequisites

- Go 1.18+
- Git
- Docker (optional)

## Environment Setup

### Required Environment Variables

```bash
# Compass integration
export COMPASS_API_TOKEN="your-compass-api-token"
export COMPASS_CLOUD_ID="your-compass-cloud-id"

# GitHub access
export GITHUB_TOKEN="your-github-personal-access-token"

# AWS/Prometheus (if using Prometheus metrics)
export AWS_REGION="us-east-1"
export PROMETHEUS_WORKSPACE_URL="https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-xxxxx/"
export AWS_ROLE="arn:aws:iam::123456789012:role/PrometheusRole"  # Optional
```

### Optional Environment Variables

```bash
# Custom metric directory (instead of default catalog)
export METRIC_DIR="/path/to/local/metrics"
# Or from a git repository
export METRIC_DIR="https://github.com/org/repo.git/path/to/metrics"
```

## Installation Options

### Option 1: Local Development

```bash
# Clone and build
git clone <repository-url>
cd compass-compute
make setup    # Installs development tools
make build    # Creates ./compass-compute binary

# Test it works
./compass-compute compute my-service --verbose
```

### Option 2: Docker

```bash
# Build Docker image
make docker-build

# Run with environment file
echo "COMPASS_API_TOKEN=your-token" > .env
echo "COMPASS_CLOUD_ID=your-cloud-id" >> .env
echo "GITHUB_TOKEN=your-github-token" >> .env

docker run --env-file .env compass-compute:latest compute my-service
```

## Verification

### Test the Setup

```bash
# Check if environment variables are set
./compass-compute compute --help

# Test with a real component (verbose mode shows all operations)
./compass-compute compute my-service --verbose
```

### Expected Output

```
Starting compass-compute with component: my-service
Found component 'my-service' (ID: comp-123, Type: service) with 3 metrics
Successfully cloned repository: my-service
Using metric directory: ./repos/of-catalog/config/grading-system
Processing metric: deployment-frequency
Evaluated metric 'deployment-frequency' with value: 5
Successfully processed 3 metrics for component 'my-service'
```

## Common Setup Issues

### "Component not found"
**Issue**: Component doesn't exist in Compass
**Fix**: Verify component name and check it exists in Compass UI

### "Failed to clone repository"
**Issue**: GitHub token lacks repository access
**Fix**: Ensure GitHub token has repo access permissions

### "AWS credentials not found"
**Issue**: Missing AWS credentials for Prometheus
**Fix**: Configure AWS CLI or set AWS environment variables

### "Metric directory not found"
**Issue**: Default metric catalog not accessible
**Fix**: Set custom `METRIC_DIR` or ensure GitHub access to catalog repo

## Next Steps

- [📊 Create your first metric](facts-guide.md)
- [🔌 Add custom data sources](extensions.md)
- [🐛 Debug common issues](debugging.md)