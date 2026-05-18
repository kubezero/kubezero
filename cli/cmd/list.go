package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available packages and active registry entries",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&repoRoot, "repo", "..", "Path to the kubezero repository root")
}

func runList(_ *cobra.Command, _ []string) error {
	packagesDir := filepath.Join(repoRoot, "packages")
	registryDir := filepath.Join(repoRoot, "registry")

	fmt.Println("Available packages (packages/):")
	if err := printDirs(packagesDir, "  "); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Active registry entries (registry/):")
	if err := printActiveRegistry(registryDir); err != nil {
		return err
	}

	return nil
}

func printDirs(dir, indent string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			fmt.Printf("%s%s\n", indent, e.Name())
		}
	}
	return nil
}

func printActiveRegistry(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", dir, err)
	}
	active := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// A registry entry is active if it contains a gitops.yaml (not _gitops.yaml)
		gitopsPath := filepath.Join(dir, e.Name(), "gitops.yaml")
		if _, err := os.Stat(gitopsPath); err == nil {
			fmt.Printf("  %s  (active)\n", e.Name())
			active++
		} else {
			fmt.Printf("  %s  (template — rename _gitops.yaml to activate)\n", e.Name())
		}
	}
	if active == 0 {
		fmt.Println("  none — run: kubezero init")
	}
	return nil
}
