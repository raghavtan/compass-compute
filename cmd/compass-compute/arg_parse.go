package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func parseFlags() (*Config, error) {
	var (
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "compass-compute - A tool to manage compass components\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [OPTIONS] <component-name>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  component-name    The name of the component to process\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s my-component\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --verbose my-component\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --version\n", os.Args[0])
	}

	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		showVersionInfo()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		return nil, fmt.Errorf("component name is required")
	}

	if len(args) > 1 {
		return nil, fmt.Errorf("too many arguments: expected 1, got %d", len(args))
	}

	componentName := strings.TrimSpace(args[0])
	if componentName == "" {
		return nil, fmt.Errorf("component name cannot be empty")
	}

	if err := validateComponentName(componentName); err != nil {
		return nil, fmt.Errorf("invalid component name: %w", err)
	}

	return &Config{
		ComponentName: componentName,
		Verbose:       *verbose,
	}, nil
}

func showVersionInfo() {
	fmt.Printf("compass-compute\n")
	fmt.Printf("Version:    %s\n", Version)
	fmt.Printf("Built:      %s\n", BuildTime)
	fmt.Printf("Commit:     %s\n", CommitHash)
}
