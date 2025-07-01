package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available packages and registry contents",
	Long: `List shows available packages in the packages directory and what's currently
prepared in the registry directory. This helps you understand what's available
and what's ready for deployment.`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	fmt.Println("📦 KubeZero Package Overview")
	fmt.Println("============================")
	fmt.Println()

	// List available packages
	fmt.Println("🗂️  Available Packages:")
	packages, err := listPackages("packages")
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	if len(packages) == 0 {
		fmt.Println("   No packages found in packages/ directory")
	} else {
		for provider, pkgs := range groupPackagesByProvider(packages) {
			fmt.Printf("   %s:\n", strings.ToUpper(provider))
			for _, pkg := range pkgs {
				fmt.Printf("     • %s\n", pkg)
			}
		}
	}

	fmt.Println()

	// List registry contents
	fmt.Println("🏭 Registry Contents:")
	registryPackages, err := listPackages("registry")
	if err != nil {
		return fmt.Errorf("failed to list registry: %w", err)
	}

	if len(registryPackages) == 0 {
		fmt.Println("   No packages in registry/ directory")
		fmt.Println("   💡 Run 'kubezero init' or 'kubezero prepare [provider]' to prepare packages")
	} else {
		for provider, pkgs := range groupPackagesByProvider(registryPackages) {
			fmt.Printf("   %s:\n", strings.ToUpper(provider))
			for _, pkg := range pkgs {
				fmt.Printf("     ✅ %s (ready for deployment)\n", pkg)
			}
		}
	}

	fmt.Println()

	// Show supported providers
	fmt.Println("☁️  Supported Providers:")
	fmt.Println("   • AWS (Amazon Web Services)")
	fmt.Println("   • GCP (Google Cloud Platform)")
	fmt.Println()

	// Show next steps
	if len(registryPackages) == 0 {
		fmt.Println("🚀 Quick Start:")
		fmt.Println("   1. Initialize for your provider: ./kubezero init")
		fmt.Println("   2. Bootstrap cluster: ./kubezero bootstrap")
	} else {
		fmt.Println("🎯 Ready for deployment!")
		fmt.Println("   Run: ./kubezero bootstrap")
	}

	return nil
}

func listPackages(directory string) ([]string, error) {
	var packages []string

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return packages, nil
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s directory: %w", directory, err)
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && entry.Name() != "test" {
			// Check if it's a valid package (has gitops.yaml or other indicators)
			gitopsPath := filepath.Join(directory, entry.Name(), "gitops.yaml")
			if _, err := os.Stat(gitopsPath); err == nil {
				packages = append(packages, entry.Name())
			} else {
				// Check for other package indicators
				infraPath := filepath.Join(directory, entry.Name(), "infrastructure")
				appsPath := filepath.Join(directory, entry.Name(), "applications")
				if _, err1 := os.Stat(infraPath); err1 == nil {
					packages = append(packages, entry.Name())
				} else if _, err2 := os.Stat(appsPath); err2 == nil {
					packages = append(packages, entry.Name())
				}
			}
		}
	}

	return packages, nil
}

func groupPackagesByProvider(packages []string) map[string][]string {
	grouped := make(map[string][]string)

	for _, pkg := range packages {
		provider := "other"

		if strings.Contains(pkg, "aws") {
			provider = "aws"
		} else if strings.Contains(pkg, "gcp") {
			provider = "gcp"
		} else if strings.Contains(pkg, "virtual") {
			provider = "virtual"
		}

		grouped[provider] = append(grouped[provider], pkg)
	}

	return grouped
}
