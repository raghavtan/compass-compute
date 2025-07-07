# 🧭 compass-compute CLI

> **Transform your component metrics with intelligent automation**

A powerful command-line tool that intelligently evaluates and submits component metrics to Atlassian Compass. Whether you're tracking code quality, deployment frequency, or custom business metrics, compass-compute makes it effortless.

---

## 🚀 Quick Start

```bash
# Install and run in 30 seconds
make setup
make  build
./compass-compute compute my-service

# Or with Docker
make docker-build
docker run compass-compute:latest compute my-service
```

**That's it!** Your component metrics are now automatically evaluated and submitted to Compass.

## 🏃‍♂️ Getting Started

### Prerequisites
- Go 1.18+
- Docker (optional)
- Git
- Access to Atlassian Compass

### Installation

**Option 1: Build from Source** (Recommended for contributors)
```bash
git clone <repository-url>
cd compass-compute
make setup    # Install development tools
make build    # Build the binary
```

**Option 2: Docker** (Great for CI/CD)
```bash
make docker-build
```

### Configuration

Set up your environment variables:
```bash
export COMPASS_API_TOKEN="your-compass-token"
export COMPASS_CLOUD_ID="your-cloud-id"
export GITHUB_TOKEN="your-github-token"
export AWS_REGION="us-west-2"  # For Prometheus integration
export PROMETHEUS_WORKSPACE_URL="your-prometheus-url"
```

---

## 🎮 Usage Examples

### Basic Usage
```bash
# Process metrics for a single component
./compass-compute compute my-service

# Process multiple components
./compass-compute compute service-a,service-b,service-c

# Enable detailed logging
./compass-compute compute my-service --verbose
```

### Advanced Scenarios
```bash
# Process with custom metric definitions
METRIC_PATH=./custom-metrics ./compass-compute compute my-service

# Run in CI/CD pipeline
docker run --env-file .env compass-compute:latest compute $SERVICE_NAME
```

---

## 📚 Documentation Hub

### 👋 **New Here?**
- [📖 User Guide](docs/user-guide.md) - Complete walkthrough for end users
- [🔧 Setup Guide](docs/setup.md) - Environment configuration and troubleshooting

### 🛠️ **Building Metrics?**
- [📊 Facts & Metrics Guide](docs/facts-and-metrics.md) - Everything about creating and managing metrics

### 🚀 **Extending the Tool?**
- [🔌 Extension Guide](docs/extending.md) - Add custom fact types, data sources, and rules
- [🏗️ Architecture Guide](docs/architecture.md) - Understanding the codebase

---

## 🏗️ Architecture at a Glance

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   compass-compute   │────│   Facts Engine   │────│   Data Sources  │
│        CLI          │    │                  │    │  • GitHub       │
└─────────────────┘    │  • Extract          │    │  • Prometheus   │
                       │  • Validate         │    │  • APIs         │
                       │  • Aggregate        │    │  • Custom...    │
                       └──────────────────┘    └─────────────────┘
                                │
                       ┌──────────────────┐
                       │ Atlassian Compass │
                       │   (Metrics API)   │
                       └──────────────────┘
```

**Key Components:**
- **CLI Interface**: Simple command-line interface for users
- **Facts Engine**: Intelligent metric evaluation with dependency resolution
- **Data Sources**: Pluggable extractors for different data sources
- **Extensibility**: Add custom processors without modifying core code

---

## 🤔 Common Use Cases

### 📈 **Code Quality Metrics**
- Test coverage from GitHub Actions
- Code complexity from static analysis
- Security scan results from CI/CD

### 🚀 **Deployment Metrics**
- Deployment frequency from CI/CD systems
- Lead time from issue tracking
- MTTR from incident management tools

### 📊 **Runtime Metrics**
- Error rates from Prometheus
- Response times from APM tools
- Resource utilization from monitoring

### 💼 **Business Metrics**
- Feature adoption from analytics
- Customer satisfaction scores
- SLA compliance metrics

