package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print detailed version information for KubeZero CLI including build details.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("KubeZero CLI %s\n", version)
		fmt.Printf("Built from commit: %s\n", commit)
		fmt.Printf("Build date: %s\n", date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
