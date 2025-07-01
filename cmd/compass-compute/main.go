package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nUse '%s --help' for more information.\n", err, os.Args[0])
		os.Exit(1)
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nUse '%s --help' for more information.\n", err, os.Args[0])
		os.Exit(1)
	}
}

func run(config *Config) error {
	if config.Verbose {
		fmt.Printf("Starting compass-compute with component: %s\n", config.ComponentName)
	}

	if err := processComponent(config); err != nil {
		return fmt.Errorf("failed to process component: %w", err)
	}

	if config.Verbose {
		fmt.Println("Component processing completed successfully")
	}

	return nil
}

func processComponent(config *Config) error {
	fmt.Printf("Processing component: %s\n", config.ComponentName)

	// GET component ID
	// GET metricName, metricDefinitionID for all metrics associated with the component
	// GET FactDefinitions for all metrics
	// Clone component repository
	// For each metric:
	// 		compute all Facts to generate metric values for each metric
	// 		Push computed metric values to the Compass API

	if config.Verbose {
		fmt.Printf("  - Validating component '%s'\n", config.ComponentName)
		fmt.Printf("  - Executing component operations\n")
		fmt.Printf("  - Finalizing component processing\n")
	}

	fmt.Printf("Component '%s' processed successfully\n", config.ComponentName)
	return nil
}
