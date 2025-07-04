package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose       bool
	allComponents bool
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.AddCommand(computeCmd)
}

var rootCmd = &cobra.Command{
	Use:   "compass-compute <component-name>",
	Short: "A tool to compute compass component metrics",
	Long:  `compass-compute is a CLI tool for compute compass component metrics`,
	Example: `  
	compass-compute compute my-component
	compass-compute compute my-component1,my-component2
  	compass-compute --verbose my-component
	compass-compute compute all
	compass-compute compute all --verbose
	`,
}
