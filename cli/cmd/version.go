package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of KubeZero CLI",
	Long:  `Print the version information for KubeZero CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("KubeZero CLI v0.1.0")
		fmt.Println("Built with Go")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
