package compute

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		showVersionInfo()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func showVersionInfo() {
	fmt.Printf("compass-compute\n")
	fmt.Printf("Version:    %s\n", Version)
	fmt.Printf("Built:      %s\n", BuildTime)
	fmt.Printf("Commit:     %s\n", CommitHash)
}
