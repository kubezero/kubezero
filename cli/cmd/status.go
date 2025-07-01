package cmd

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the KubeZero cluster",
	Long:  `Display information about the current KubeZero cluster including nodes and ArgoCD pods.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 KubeZero Cluster Status")
	fmt.Println("==========================")

	// Check if k3d cluster exists
	if err := checkClusterExists(); err != nil {
		fmt.Printf("❌ No KubeZero cluster found: %v\n", err)
		return err
	}

	// Get Kubernetes client
	client, err := getKubernetesClient()
	if err != nil {
		fmt.Printf("❌ Failed to connect to cluster: %v\n", err)
		return err
	}

	ctx := context.Background()

	// Get nodes
	fmt.Println("\n📊 Cluster Nodes:")
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("❌ Failed to get nodes: %v\n", err)
	} else {
		for _, node := range nodes.Items {
			status := "NotReady"
			for _, condition := range node.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					status = "Ready"
					break
				}
			}
			fmt.Printf("  ✅ %s (%s)\n", node.Name, status)
		}
	}

	// Get ArgoCD pods
	fmt.Println("\n🚀 ArgoCD Components:")
	pods, err := client.CoreV1().Pods("kubezero").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("❌ Failed to get pods: %v\n", err)
	} else {
		argoCDPods := []string{
			"argo-cd-server",
			"argo-cd-application-controller",
			"argo-cd-repo-server",
		}

		for _, requiredPod := range argoCDPods {
			found := false
			for _, pod := range pods.Items {
				if matchesPodName(pod.Name, requiredPod) {
					found = true
					if isPodReady(pod) {
						fmt.Printf("  ✅ %s (Ready)\n", pod.Name)
					} else {
						fmt.Printf("  ⏳ %s (Phase: %s)\n", pod.Name, pod.Status.Phase)
					}
					break
				}
			}
			if !found {
				fmt.Printf("  ❌ %s (Not Found)\n", requiredPod)
			}
		}
	}

	fmt.Println("\n💡 Tip: Access ArgoCD at http://gitops.local.kubezero.io")

	return nil
}

func checkClusterExists() error {
	cmd := exec.Command("k3d", "cluster", "list", "kubezero")
	return cmd.Run()
}
