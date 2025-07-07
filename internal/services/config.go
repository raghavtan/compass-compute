package services

const (
	CatalogRepo       = "of-catalog"
	GitHubOrg         = "motain"
	MetricPath        = "config/grading-system"
	CompassBaseURL    = "https://onefootball.atlassian.net/gateway/api"
	GraphQLEndpoint   = CompassBaseURL + "/graphql"
	MetricsEndpoint   = CompassBaseURL + "/compass/v1/metrics"
	ServiceSlugPrefix = "svc-"
	LocalBasePath     = "./repos/"
	MetricLocalPath   = LocalBasePath + CatalogRepo + "/" + MetricPath
)
