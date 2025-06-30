package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	forceDelete bool
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Delete the KubeZero cluster",
	Long:  `Remove the local k3d KubeZero cluster and all associated resources.`,
	RunE:  runCleanup,
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force delete without confirmation")
}

func runCleanup(cmd *cobra.Command, args []string) error {
	fmt.Println("🧹 KubeZero Cluster Cleanup")

	// Check if cluster exists
	if err := checkClusterExists(); err != nil {
		fmt.Println("ℹ️  No KubeZero cluster found to delete.")
		return nil
	}

	// Confirmation prompt unless forced
	if !forceDelete {
		fmt.Print("⚠️  This will permanently delete the KubeZero cluster. Continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("❌ Cleanup cancelled.")
			return nil
		}
	}

	// Delete the cluster
	fmt.Println("🗑️  Deleting k3d cluster...")
	cmd_exec := exec.Command("k3d", "cluster", "delete", "kubezero")
	cmd_exec.Stdout = os.Stdout
	cmd_exec.Stderr = os.Stderr

	if err := cmd_exec.Run(); err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	fmt.Println("✅ KubeZero cluster deleted successfully!")
	fmt.Println("💡 You can create a new cluster anytime with: kubezero bootstrap")

	return nil
}
