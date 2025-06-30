package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var configPath string

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Create the local k3d bootstrap cluster",
	Long: `bootstrap runs k3d to create the local cluster defined in
bootstrap/k3d-bootstrap-cluster.yaml. ArgoCD is installed automatically
by the K3s HelmController and will start syncing the registry.`,
	RunE: runBootstrap,
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
	bootstrapCmd.Flags().StringVarP(&configPath, "config", "c",
		"../bootstrap/k3d-bootstrap-cluster.yaml",
		"Path to the k3d cluster configuration file")
}

func runBootstrap(_ *cobra.Command, _ []string) error {
	if _, err := exec.LookPath("k3d"); err != nil {
		return fmt.Errorf("k3d not found in PATH — install it first: https://k3d.io")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	fmt.Printf("Creating k3d cluster from %s ...\n", configPath)

	cmd := exec.Command("k3d", "cluster", "create", "--config", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("k3d cluster create failed: %w", err)
	}

	fmt.Println()
	fmt.Println("Cluster created. ArgoCD will be available shortly at:")
	fmt.Println("  http://gitops.local.kubezero.io")
	fmt.Println()
	fmt.Println("Watch progress:")
	fmt.Println("  kubectl get pods -n kubezero -w")
	fmt.Println("  kubectl get applications -n kubezero")

	return nil
}
