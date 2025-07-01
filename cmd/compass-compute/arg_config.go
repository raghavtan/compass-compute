package main

var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

type Config struct {
	ComponentName string
	Verbose       bool
}
