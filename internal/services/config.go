package services

import (
	"os"
	"path/filepath"
)

const (
	CatalogRepo            = "of-catalog"
	GitHubOrg              = "motain"
	MetricPath             = "config/grading-system"
	CompassBaseURL         = "https://onefootball.atlassian.net/gateway/api"
	GraphQLEndpoint        = CompassBaseURL + "/graphql"
	MetricsEndpoint        = CompassBaseURL + "/compass/v1/metrics"
	ServiceSlugPrefix      = "svc-"
	LocalBasePath          = "./repos/"
	DefaultMetricLocalPath = LocalBasePath + CatalogRepo + "/" + MetricPath
)

func GetMetricLocalPath() string {
	metricDir := os.Getenv("METRIC_DIR")
	if metricDir != "" {
		return filepath.Join(LocalBasePath, "metrics")
	}
	return DefaultMetricLocalPath
}
