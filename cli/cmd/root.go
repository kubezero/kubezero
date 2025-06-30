package cmd

import (
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "kubezero",
	Short: "KubeZero CLI - zero-friction Kubernetes platform setup",
	Long: `kubezero helps you configure a KubeZero platform by selecting a
deployment topology and activating the right packages in your registry.`,
}

func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}
