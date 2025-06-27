package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	configPath string
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap a local KubeZero cluster",
	Long: `Bootstrap creates a local k3d cluster using the configuration in bootstrap/k3d-bootstrap-cluster.yaml
and monitors the deployment until ArgoCD components are ready.`,
	RunE: runBootstrap,
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
	bootstrapCmd.Flags().StringVarP(&configPath, "config", "c", "bootstrap/k3d-bootstrap-cluster.yaml", "Path to k3d cluster configuration file")
}

func runBootstrap(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Starting KubeZero cluster bootstrap...")

	// Check if k3d is installed
	if err := checkK3dInstalled(); err != nil {
		return fmt.Errorf("k3d not found: %w", err)
	}

	// Check if config file exists
	if !fileExists(configPath) {
		return fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Create k3d cluster
	fmt.Printf("📋 Creating cluster with config: %s\n", configPath)
	if err := createK3dCluster(configPath); err != nil {
		return fmt.Errorf("failed to create k3d cluster: %w", err)
	}

	fmt.Println("✅ Cluster created successfully!")

	// Wait for cluster to be ready
	fmt.Println("⏳ Waiting for cluster to be ready...")
	if err := waitForClusterReady(); err != nil {
		return fmt.Errorf("cluster not ready: %w", err)
	}

	// Monitor ArgoCD pods
	fmt.Println("🔍 Monitoring ArgoCD deployment...")
	if err := monitorArgoCDPods(); err != nil {
		return fmt.Errorf("ArgoCD monitoring failed: %w", err)
	}

	fmt.Println("🎉 KubeZero cluster bootstrap completed successfully!")
	fmt.Println("📍 ArgoCD should be available at: http://gitops.local.kubezero.io")
	fmt.Println("💡 Run 'kubectl get pods -n kubezero' to check pod status")

	return nil
}

func checkK3dInstalled() error {
	_, err := exec.LookPath("k3d")
	if err != nil {
		return fmt.Errorf("k3d is not installed or not in PATH")
	}
	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func createK3dCluster(configPath string) error {
	cmd := exec.Command("k3d", "cluster", "create", "--config", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForClusterReady() error {
	// Wait a bit for the cluster to initialize
	time.Sleep(10 * time.Second)

	client, err := getKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx := context.Background()
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for cluster to be ready")
		case <-ticker.C:
			// Check if we can list nodes
			nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err != nil {
				fmt.Printf("⏳ Waiting for cluster API... (%v)\n", err)
				continue
			}

			if len(nodes.Items) > 0 {
				fmt.Printf("✅ Cluster API is ready with %d node(s)\n", len(nodes.Items))
				return nil
			}
		}
	}
}

func monitorArgoCDPods() error {
	client, err := getKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx := context.Background()
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	requiredPods := []string{
		"argo-cd-server",
		"argo-cd-application-controller",
		"argo-cd-repo-server",
	}

	fmt.Println("👀 Waiting for ArgoCD pods to be ready...")

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for ArgoCD pods to be ready")
		case <-ticker.C:
			pods, err := client.CoreV1().Pods("kubezero").List(ctx, metav1.ListOptions{})
			if err != nil {
				fmt.Printf("⏳ Waiting for kubezero namespace... (%v)\n", err)
				continue
			}

			readyPods := make(map[string]bool)
			for _, pod := range pods.Items {
				for _, requiredPod := range requiredPods {
					if matchesPodName(pod.Name, requiredPod) {
						if isPodReady(pod) {
							readyPods[requiredPod] = true
							fmt.Printf("✅ %s is ready\n", pod.Name)
						} else {
							fmt.Printf("⏳ %s is not ready yet (Phase: %s)\n", pod.Name, pod.Status.Phase)
						}
					}
				}
			}

			// Check if all required pods are ready
			allReady := true
			for _, requiredPod := range requiredPods {
				if !readyPods[requiredPod] {
					allReady = false
					break
				}
			}

			if allReady {
				fmt.Println("🎯 All ArgoCD pods are ready!")
				return nil
			}
		}
	}
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func matchesPodName(podName, expectedPrefix string) bool {
	return len(podName) >= len(expectedPrefix) && podName[:len(expectedPrefix)] == expectedPrefix
}

func isPodReady(pod corev1.Pod) bool {
	// Check if pod phase is Running and all containers are ready
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}
