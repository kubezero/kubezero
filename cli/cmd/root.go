package cmd

import (
	"github.com/spf13/cobra"
)

// Version information - set from main.go
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "kubezero",
	Short: "KubeZero CLI - Platform infrastructure made simple",
	Long: `KubeZero CLI is a command line tool to help initialize and manage
KubeZero platform infrastructure with GitOps principles.

This tool helps you bootstrap a local k3d cluster with ArgoCD
and monitor the deployment status of core components.`,
}

// SetVersionInfo sets the version information from the main package
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Configuration initialization if needed
}
