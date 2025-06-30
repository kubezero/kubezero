package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	configPath    string
	cloudProvider string
	region        string
	interactive   bool
	localMode     bool
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap a local KubeZero cluster with cloud provider setup",
	Long: `Bootstrap creates a local k3d cluster and prepares packages for your chosen cloud provider.

This command will:
1. Interactively ask for cloud provider and region (if not specified)
2. Prepare packages for the selected cloud provider
3. Create a k3d cluster using the configuration
4. Monitor the deployment until ArgoCD components are ready
5. Deploy cloud-specific packages to the registry

The --local flag automates local development setup by:
- Opening GitHub fork page in your browser
- Populating the registry/ directory with cloud-specific packages  
- Adding your fork as a git remote
- Updating all repoURL references to point to your fork
- Committing and pushing changes to your fork

Examples:
  kubezero bootstrap --cloud aws --region eu-west-1
  kubezero bootstrap --interactive
  kubezero bootstrap --local --cloud aws --region eu-west-1
  kubezero bootstrap  # Will prompt for cloud provider selection`,
	RunE: runBootstrap,
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
	bootstrapCmd.Flags().StringVarP(&configPath, "config", "c", "../bootstrap/k3d-bootstrap-cluster.yaml", "Path to k3d cluster configuration file")
	bootstrapCmd.Flags().StringVar(&cloudProvider, "cloud", "", "Cloud provider (aws, gcp)")
	bootstrapCmd.Flags().StringVar(&region, "region", "", "Cloud region (e.g., eu-west-1, us-central1)")
	bootstrapCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Force interactive mode even if flags are provided")
	bootstrapCmd.Flags().BoolVar(&localMode, "local", false, "Automate local development setup (opens GitHub fork, updates configs, pushes to your fork)")
}

func runBootstrap(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Starting KubeZero cluster bootstrap")
	fmt.Println()

	// Step 1: Cloud provider and region selection
	if err := handleCloudProviderSelection(); err != nil {
		return fmt.Errorf("cloud provider selection failed: %w", err)
	}

	// Step 2: Prepare packages for the selected cloud provider
	fmt.Printf("📦 Preparing packages for %s in region %s...\n", cloudProvider, region)
	if err := prepareCloudPackages(); err != nil {
		return fmt.Errorf("failed to prepare packages: %w", err)
	}

	// Step 2.5: Handle local mode configuration
	if localMode {
		fmt.Println("🔧 Configuring for local development...")
		if err := configureLocalMode(); err != nil {
			return fmt.Errorf("failed to configure local mode: %w", err)
		}
	}

	// Step 3: Check prerequisites
	fmt.Println("🔍 Checking prerequisites...")
	if err := checkPrerequisites(); err != nil {
		return err
	}

	// Step 4: Create k3d cluster
	fmt.Printf("📋 Creating cluster with config: %s\n", configPath)
	if err := createK3dCluster(configPath); err != nil {
		return fmt.Errorf("failed to create k3d cluster: %w", err)
	}

	fmt.Println("✅ Cluster created successfully!")

	// Step 5: Wait for cluster to be ready
	fmt.Println("⏳ Waiting for cluster to be ready...")
	if err := waitForClusterReady(); err != nil {
		return fmt.Errorf("cluster not ready: %w", err)
	}

	// Step 6: Monitor ArgoCD pods
	fmt.Println("🔍 Monitoring ArgoCD deployment...")
	if err := monitorArgoCDPods(); err != nil {
		return fmt.Errorf("ArgoCD monitoring failed: %w", err)
	}

	// Step 7: Show completion message
	fmt.Println()
	fmt.Println("🎉 KubeZero cluster bootstrap completed successfully")
	fmt.Println()
	fmt.Printf("☁️  Cloud Provider: %s\n", cloudProvider)
	fmt.Printf("🌍 Region: %s\n", region)
	if localMode {
		fmt.Printf("🏠 Local Mode: ✅ Enabled (using local git repository)\n")
	}
	fmt.Println("📍 ArgoCD URL: http://gitops.local.kubezero.io")
	fmt.Printf("📁 Packages deployed to: ../registry/\n")
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   • Check pod status: kubectl get pods -n kubezero")
	fmt.Println("   • View applications: kubectl get applications -n kubezero")
	fmt.Printf("   • Check %s packages: ls -la ../registry/\n", cloudProvider)
	if localMode {
		fmt.Println("   • All GitOps manifests now reference your local repository")
		fmt.Println("   • Make changes locally and commit to trigger ArgoCD sync")
	}

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
		"argo-cd-app-controller",
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

// Cloud provider configurations
type CloudConfig struct {
	Name    string
	Regions []string
}

var supportedClouds = map[string]CloudConfig{
	"aws": {
		Name: "Amazon Web Services (AWS)",
		Regions: []string{
			"us-east-1", "us-east-2", "us-west-1", "us-west-2",
			"eu-west-1", "eu-west-2", "eu-central-1",
			"ap-southeast-1", "ap-southeast-2", "ap-northeast-1",
		},
	},
	"gcp": {
		Name: "Google Cloud Platform (GCP)",
		Regions: []string{
			"us-central1", "us-east1", "us-west1", "us-west2",
			"europe-west1", "europe-west2", "europe-central2",
			"asia-southeast1", "asia-east1", "asia-northeast1",
		},
	},
}

func handleCloudProviderSelection() error {
	// If both cloud and region are provided and not in interactive mode, validate and use them
	if cloudProvider != "" && region != "" && !interactive {
		return validateCloudConfig(cloudProvider, region)
	}

	// Interactive mode or missing parameters
	if cloudProvider == "" || region == "" || interactive {
		fmt.Println("☁️  Cloud Provider Configuration")
		fmt.Println()
	}

	// Select cloud provider
	if cloudProvider == "" || interactive {
		if err := selectCloudProvider(); err != nil {
			return err
		}
	}

	// Select region
	if region == "" || interactive {
		if err := selectRegion(); err != nil {
			return err
		}
	}

	return validateCloudConfig(cloudProvider, region)
}

func selectCloudProvider() error {
	providers := make([]string, 0, len(supportedClouds))
	for key := range supportedClouds {
		providers = append(providers, key)
	}

	// Create options with descriptions
	var options []string
	for _, provider := range providers {
		config := supportedClouds[provider]
		options = append(options, fmt.Sprintf("%s (%s)", config.Name, provider))
	}

	prompt := promptui.Select{
		Label: "Select Cloud Provider",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U0001F449 {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "\U0001F44D {{ . | green }}",
		},
	}

	i, _, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select cloud provider: %w", err)
	}

	cloudProvider = providers[i]
	fmt.Printf("✅ Selected: %s\n", supportedClouds[cloudProvider].Name)
	return nil
}

func selectRegion() error {
	config := supportedClouds[cloudProvider]

	prompt := promptui.Select{
		Label: fmt.Sprintf("Select Region for %s", config.Name),
		Items: config.Regions,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U0001F449 {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "\U0001F44D {{ . | green }}",
		},
	}

	i, _, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select region: %w", err)
	}

	region = config.Regions[i]
	fmt.Printf("✅ Selected region: %s\n", region)
	return nil
}

func validateCloudConfig(provider, reg string) error {
	config, exists := supportedClouds[provider]
	if !exists {
		available := make([]string, 0, len(supportedClouds))
		for k := range supportedClouds {
			available = append(available, k)
		}
		return fmt.Errorf("unsupported cloud provider: %s. Available: %v", provider, available)
	}

	// Check if region is valid for the provider
	validRegion := false
	for _, validReg := range config.Regions {
		if validReg == reg {
			validRegion = true
			break
		}
	}

	if !validRegion {
		return fmt.Errorf("unsupported region '%s' for %s. Available regions: %v", reg, provider, config.Regions)
	}

	return nil
}

func prepareCloudPackages() error {
	// Find packages for the provider
	packageDir := "../packages"
	packages, err := findProviderPackages(packageDir, cloudProvider)
	if err != nil {
		return fmt.Errorf("failed to find packages: %w", err)
	}

	if len(packages) == 0 {
		fmt.Printf("⚠️  No specific packages found for provider: %s\n", cloudProvider)
		fmt.Println("   Using default configuration...")
		return nil
	}

	fmt.Printf("📦 Found %d package(s): %v\n", len(packages), packages)

	// Create registry directory if it doesn't exist
	registryDir := "../registry"
	if err := os.MkdirAll(registryDir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	// Prepare each package
	for _, pkg := range packages {
		fmt.Printf("🔧 Preparing package: %s\n", pkg)
		if err := preparePackageForRegistry(packageDir, pkg, registryDir); err != nil {
			return fmt.Errorf("failed to prepare package %s: %w", pkg, err)
		}
		fmt.Printf("✅ Package %s prepared successfully\n", pkg)
	}

	return nil
}

func checkPrerequisites() error {
	// Check if k3d is installed
	if err := checkK3dInstalled(); err != nil {
		return fmt.Errorf("k3d not found: %w", err)
	}

	// Check if config file exists
	if !fileExists(configPath) {
		return fmt.Errorf("configuration file not found: %s", configPath)
	}

	return nil
}

// Helper functions for package preparation
func findProviderPackages(packageDir, provider string) ([]string, error) {
	var packages []string

	if !fileExists(packageDir) {
		return packages, nil // No packages directory, that's okay
	}

	entries, err := os.ReadDir(packageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read packages directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(entry.Name(), provider) {
			packages = append(packages, entry.Name())
		}
	}

	return packages, nil
}

func preparePackageForRegistry(packageDir, packageName, registryDir string) error {
	srcDir := filepath.Join(packageDir, packageName)
	dstDir := filepath.Join(registryDir, packageName)

	// Create destination directory
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy package contents
	if err := copyDir(srcDir, dstDir); err != nil {
		return fmt.Errorf("failed to copy package: %w", err)
	}

	// Update configuration files if they exist
	if err := updatePackageConfig(dstDir); err != nil {
		return fmt.Errorf("failed to update package config: %w", err)
	}

	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func updatePackageConfig(packageDir string) error {
	// Look for common configuration files and update them
	configFiles := []string{"gitops.yaml", "kustomization.yaml", "values.yaml"}

	for _, configFile := range configFiles {
		filePath := filepath.Join(packageDir, configFile)
		if fileExists(filePath) {
			if err := updateConfigFile(filePath); err != nil {
				return fmt.Errorf("failed to update %s: %w", configFile, err)
			}
		}
	}

	// Handle infrastructure patches for cloud providers
	infraDir := filepath.Join(packageDir, "infrastructure")
	if fileExists(infraDir) {
		if err := updateInfrastructurePatches(infraDir); err != nil {
			return fmt.Errorf("failed to update infrastructure patches: %w", err)
		}
	}

	return nil
}

func updateConfigFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	updatedContent := string(content)

	// Replace common patterns that need updating
	replacements := map[string]string{
		"REGION_PLACEHOLDER":         region,
		"CLOUD_PROVIDER_PLACEHOLDER": cloudProvider,
		"${AWS_REGION}":              region,
		"${GCP_REGION}":              region,
		"${CLOUD_PROVIDER}":          cloudProvider,
	}

	// Apply replacements
	for old, new := range replacements {
		updatedContent = strings.ReplaceAll(updatedContent, old, new)
	}

	// Write the updated content back if changes were made
	if updatedContent != string(content) {
		if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
			return fmt.Errorf("failed to write updated config file: %w", err)
		}
		fmt.Printf("   📝 Updated: %s\n", filepath.Base(filePath))
	}

	return nil
}

func updateInfrastructurePatches(infraDir string) error {
	// Define region-specific availability zones
	regionAZMap := map[string][]string{
		// AWS regions
		"us-east-1":      {"us-east-1a", "us-east-1b", "us-east-1c"},
		"us-east-2":      {"us-east-2a", "us-east-2b", "us-east-2c"},
		"us-west-1":      {"us-west-1a", "us-west-1b"},
		"us-west-2":      {"us-west-2a", "us-west-2b", "us-west-2c"},
		"eu-west-1":      {"eu-west-1a", "eu-west-1b", "eu-west-1c"},
		"eu-west-2":      {"eu-west-2a", "eu-west-2b", "eu-west-2c"},
		"eu-central-1":   {"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		"ap-southeast-1": {"ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c"},
		"ap-southeast-2": {"ap-southeast-2a", "ap-southeast-2b", "ap-southeast-2c"},
		"ap-northeast-1": {"ap-northeast-1a", "ap-northeast-1b", "ap-northeast-1c"},

		// GCP regions
		"us-central1":     {"us-central1-a", "us-central1-b", "us-central1-c"},
		"us-east1":        {"us-east1-a", "us-east1-b", "us-east1-c"},
		"us-west1":        {"us-west1-a", "us-west1-b", "us-west1-c"},
		"us-west2":        {"us-west2-a", "us-west2-b", "us-west2-c"},
		"europe-west1":    {"europe-west1-a", "europe-west1-b", "europe-west1-c"},
		"europe-west2":    {"europe-west2-a", "europe-west2-b", "europe-west2-c"},
		"europe-central2": {"europe-central2-a", "europe-central2-b", "europe-central2-c"},
		"asia-southeast1": {"asia-southeast1-a", "asia-southeast1-b", "asia-southeast1-c"},
		"asia-east1":      {"asia-east1-a", "asia-east1-b", "asia-east1-c"},
		"asia-northeast1": {"asia-northeast1-a", "asia-northeast1-b", "asia-northeast1-c"},
	}

	// Get availability zones for the selected region
	availabilityZones, exists := regionAZMap[region]
	if !exists {
		fmt.Printf("   ⚠️  No predefined AZs for region %s, using default configuration\n", region)
		return nil
	}

	// Update XNetwork patch file
	if cloudProvider == "aws" {
		if err := updateXNetworkPatch(infraDir, availabilityZones); err != nil {
			return fmt.Errorf("failed to update XNetwork patch: %w", err)
		}
	}

	// Update XEKS patch file
	if err := updateXEKSPatch(infraDir); err != nil {
		return fmt.Errorf("failed to update XEKS patch: %w", err)
	}

	return nil
}

func updateXNetworkPatch(infraDir string, availabilityZones []string) error {
	patchFile := filepath.Join(infraDir, "patch-xnetwork.yaml")

	// Create a comprehensive patch for XNetwork with region and subnets
	patchContent := fmt.Sprintf(`---
apiVersion: aws.platform.upbound.io/v1alpha1
kind: XNetwork
metadata:
  name: aws-network
spec:
  parameters:
    providerConfigName: management-provider-aws
    region: %s
    vpcCidrBlock: 192.168.0.0/16
    subnets:
      - availabilityZone: %s
        type: public
        cidrBlock: 192.168.0.0/18
      - availabilityZone: %s
        type: public
        cidrBlock: 192.168.64.0/18
      - availabilityZone: %s
        type: private
        cidrBlock: 192.168.128.0/18
      - availabilityZone: %s
        type: private
        cidrBlock: 192.168.192.0/18
`, region, availabilityZones[0], availabilityZones[1], availabilityZones[0], availabilityZones[1])

	if err := os.WriteFile(patchFile, []byte(patchContent), 0644); err != nil {
		return fmt.Errorf("failed to write XNetwork patch: %w", err)
	}

	fmt.Printf("   📝 Updated: patch-xnetwork.yaml (region: %s)\n", region)
	return nil
}

func updateXEKSPatch(infraDir string) error {
	patchFile := filepath.Join(infraDir, "patch-xeks.yaml")

	// Determine package type from the path
	packageType := "management"
	if strings.Contains(infraDir, "worker") {
		packageType = "worker"
	}

	// Create a comprehensive patch for XEKS with region
	var patchContent string

	if cloudProvider == "aws" {
		patchContent = fmt.Sprintf(`---
apiVersion: aws.platform.upbound.io/v1alpha1
kind: XEKS
metadata:
  name: aws-eks
spec:
  parameters:
    providerConfigName: %s-provider-aws
    region: %s
  writeConnectionSecretToRef:
    name: %s-aws-eks-kubeconfig
    namespace: crossplane-system
`, packageType, region, packageType)
	} else if cloudProvider == "gcp" {
		patchContent = fmt.Sprintf(`---
apiVersion: gcp.platform.upbound.io/v1alpha1
kind: XGKE
metadata:
  name: gcp-gke
spec:
  parameters:
    providerConfigName: %s-provider-gcp
    region: %s
  writeConnectionSecretToRef:
    name: %s-gcp-gke-kubeconfig
    namespace: crossplane-system
`, packageType, region, packageType)
	}

	if err := os.WriteFile(patchFile, []byte(patchContent), 0644); err != nil {
		return fmt.Errorf("failed to write XEKS patch: %w", err)
	}

	fmt.Printf("   📝 Updated: patch-xeks.yaml (region: %s)\n", region)
	return nil
}

func configureLocalMode() error {
	fmt.Println("   🏠 Configuring for local development...")

	// Get the absolute path to the kubezero directory
	kubeZeroDir, err := filepath.Abs("../")
	if err != nil {
		return fmt.Errorf("failed to get kubezero directory path: %w", err)
	}

	// Initialize git repository if it doesn't exist and commit changes
	if err := initializeLocalGitRepo(kubeZeroDir); err != nil {
		return fmt.Errorf("failed to initialize local git repo: %w", err)
	}

	fmt.Println("   � Your local registry/ directory is now populated with:")

	// List the packages that were prepared
	registryDir := "../registry"
	if entries, err := os.ReadDir(registryDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				fmt.Printf("   • %s/\n", entry.Name())
			}
		}
	}

	fmt.Println("")
	fmt.Println("   🚀 Setting up GitHub fork workflow...")

	// Step 1: Open GitHub fork URL in browser
	forkURL := "https://github.com/kubezero/kubezero/fork"
	fmt.Printf("   🌐 Opening GitHub fork page: %s\n", forkURL)
	if err := openBrowser(forkURL); err != nil {
		fmt.Printf("   ⚠️  Could not open browser automatically: %v\n", err)
		fmt.Printf("   📝 Please manually open: %s\n", forkURL)
	}

	fmt.Println("")
	fmt.Println("   📋 Please:")
	fmt.Println("   1. Complete the fork creation in your browser")
	fmt.Println("   2. Note your GitHub username from the fork URL")
	fmt.Println("")

	// Step 2: Get GitHub username from user
	var githubUsername string
	for {
		fmt.Print("   👤 Enter your GitHub username: ")
		if _, err := fmt.Scanln(&githubUsername); err != nil {
			fmt.Printf("   ❌ Error reading input: %v\n", err)
			continue
		}
		if githubUsername == "" {
			fmt.Println("   ❌ Username cannot be empty, please try again")
			continue
		}
		break
	}

	fmt.Printf("   ✅ Using GitHub username: %s\n", githubUsername)

	// Step 3: Add fork as remote and update configurations
	forkRemoteURL := fmt.Sprintf("https://github.com/%s/kubezero.git", githubUsername)
	if err := setupForkWorkflow(kubeZeroDir, githubUsername, forkRemoteURL); err != nil {
		return fmt.Errorf("failed to setup fork workflow: %w", err)
	}

	// Step 4: Show completion message
	fmt.Println("")
	fmt.Println("   🎉 Local development setup completed!")
	fmt.Printf("   🔗 Your fork: https://github.com/%s/kubezero\n", githubUsername)
	fmt.Println("   📁 Registry populated with cloud-specific packages")
	fmt.Println("   🚀 ArgoCD manifests now point to your fork")
	fmt.Println("   💡 Future changes will sync from your fork automatically")

	return nil
}

func initializeLocalGitRepo(repoPath string) error {
	// Check if .git directory exists
	gitDir := filepath.Join(repoPath, ".git")
	if fileExists(gitDir) {
		// Repository already exists, just add and commit any changes
		return commitChanges(repoPath)
	}

	// Initialize new git repository
	cmds := [][]string{
		{"git", "init"},
		{"git", "add", "."},
		{"git", "commit", "-m", "Initial commit for local KubeZero setup"},
	}

	for _, cmd := range cmds {
		execCmd := exec.Command(cmd[0], cmd[1:]...)
		execCmd.Dir = repoPath
		if output, err := execCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to run %v: %w\nOutput: %s", cmd, err, output)
		}
	}

	return nil
}

func commitChanges(repoPath string) error {
	// Add all changes including submodules
	addCmds := [][]string{
		{"git", "add", "."},
		{"git", "add", "-A"},
	}

	for _, cmd := range addCmds {
		execCmd := exec.Command(cmd[0], cmd[1:]...)
		execCmd.Dir = repoPath
		// Ignore errors for git add (might fail for various reasons)
		execCmd.CombinedOutput()
	}

	// Check if there are changes to commit
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = repoPath
	statusOutput, err := statusCmd.CombinedOutput()
	if err != nil {
		// If we can't check status, just try to commit anyway
		fmt.Printf("   ⚠️  Could not check git status, proceeding with commit attempt\n")
	} else if len(strings.TrimSpace(string(statusOutput))) == 0 {
		// No changes to commit
		return nil
	}

	// Try to commit changes
	commitCmd := exec.Command("git", "commit", "-m", "Update for local KubeZero development")
	commitCmd.Dir = repoPath
	// Ignore errors for git commit (might fail if no changes or other reasons)
	commitCmd.CombinedOutput()

	return nil
}

// Enhanced fork workflow functions
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch {
	case fileExists("/usr/bin/open"): // macOS
		cmd = "open"
		args = []string{url}
	case fileExists("/usr/bin/xdg-open"): // Linux
		cmd = "xdg-open"
		args = []string{url}
	default:
		return fmt.Errorf("no suitable browser opener found")
	}

	return exec.Command(cmd, args...).Start()
}

func setupForkWorkflow(repoPath, githubUsername, forkRemoteURL string) error {
	// Step 1: Add fork as remote
	fmt.Println("   📡 Adding fork as remote...")
	if err := addForkRemote(repoPath, forkRemoteURL); err != nil {
		return fmt.Errorf("failed to add fork remote: %w", err)
	}

	// Step 2: Update repoURL references to use the fork
	fmt.Println("   📝 Updating manifest references to use your fork...")
	if err := updateRepoURLsToFork(githubUsername); err != nil {
		return fmt.Errorf("failed to update repoURL references: %w", err)
	}

	// Step 3: Commit changes and push to fork
	fmt.Println("   📤 Committing and pushing changes to your fork...")
	if err := commitAndPushToFork(repoPath); err != nil {
		return fmt.Errorf("failed to commit and push to fork: %w", err)
	}

	return nil
}

func addForkRemote(repoPath, forkRemoteURL string) error {
	// Check if fork remote already exists
	checkCmd := exec.Command("git", "remote", "get-url", "fork")
	checkCmd.Dir = repoPath
	if existingOutput, err := checkCmd.CombinedOutput(); err == nil {
		existingURL := strings.TrimSpace(string(existingOutput))
		if existingURL == forkRemoteURL {
			fmt.Printf("   ✅ Fork remote already exists with correct URL: %s\n", existingURL)
			return nil
		}
		
		// Remote exists but with different URL, remove it first
		fmt.Println("   🔧 Updating existing fork remote...")
		removeCmd := exec.Command("git", "remote", "remove", "fork")
		removeCmd.Dir = repoPath
		if output, err := removeCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to remove existing fork remote: %w\nOutput: %s", err, output)
		}
	}

	// Add fork remote
	fmt.Printf("   ➕ Adding fork remote: %s\n", forkRemoteURL)
	cmd := exec.Command("git", "remote", "add", "fork", forkRemoteURL)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add fork remote: %w\nOutput: %s", err, output)
	}

	return nil
}

func updateRepoURLsToFork(githubUsername string) error {
	forkRepoURL := fmt.Sprintf("https://github.com/%s/kubezero", githubUsername)

	// Files to update
	filesToUpdate := []string{
		"../bootstrap/kubezero-bootstrap-manifests.yaml",
		"../controller/argo-cd/application.yaml",
		"../controller/crossplane/application.yaml",
		"../controller/external-secrets/application.yaml",
		"../controller/gitops/application.yaml",
	}

	// Also update registry files
	registryFiles, _ := findRegistryGitOpsFiles("../registry")
	filesToUpdate = append(filesToUpdate, registryFiles...)

	for _, filePath := range filesToUpdate {
		if fileExists(filePath) {
			if err := updateRepoURLInFile(filePath, forkRepoURL); err != nil {
				fmt.Printf("   ⚠️  Warning: Could not update %s: %v\n", filePath, err)
			}
		}
	}

	return nil
}

func findRegistryGitOpsFiles(registryDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(registryDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "gitops.yaml") || strings.HasSuffix(path, "_gitops.yaml") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func updateRepoURLInFile(filePath, newRepoURL string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Replace GitHub kubezero/kubezero URLs with the fork URL
	oldContent := string(content)
	newContent := strings.ReplaceAll(oldContent, "https://github.com/kubezero/kubezero", newRepoURL)

	// Also update sourceRepos sections
	lines := strings.Split(newContent, "\n")
	for i, line := range lines {
		if strings.Contains(line, "- https://github.com/kubezero/kubezero") {
			lines[i] = strings.Replace(line, "https://github.com/kubezero/kubezero", newRepoURL, 1)
		}
	}
	newContent = strings.Join(lines, "\n")

	if newContent != oldContent {
		return os.WriteFile(filePath, []byte(newContent), 0644)
	}
	return nil
}

func commitAndPushToFork(repoPath string) error {
	// Get the current branch name with fallback handling
	currentBranch, err := getCurrentBranch(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	fmt.Printf("   📋 Current branch: %s\n", currentBranch)

	// Check current git status first
	fmt.Println("   🔍 Checking git status...")
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = repoPath
	statusOutput, err := statusCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w\nOutput: %s", err, statusOutput)
	}

	// If there are any changes, show them
	if len(strings.TrimSpace(string(statusOutput))) > 0 {
		fmt.Printf("   📊 Files to be committed:\n%s\n", string(statusOutput))
	}

	// For submodules and complex repositories, we need to be more careful
	// First, add all tracked files and new files
	fmt.Println("   📦 Adding changes to git...")

	// Add all changes, including new files and modifications
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = repoPath
	if output, err := addCmd.CombinedOutput(); err != nil {
		fmt.Printf("   ⚠️  Warning during git add -A: %v\nOutput: %s\n", err, output)
	}

	// Check if there are changes to commit after adding
	statusCmd = exec.Command("git", "diff", "--cached", "--quiet")
	statusCmd.Dir = repoPath
	if err := statusCmd.Run(); err == nil {
		fmt.Println("   ℹ️  No changes staged for commit")

		// Check if there are untracked files or other issues
		statusCmd = exec.Command("git", "status", "--porcelain")
		statusCmd.Dir = repoPath
		statusOutput, err = statusCmd.CombinedOutput()
		if err == nil && len(strings.TrimSpace(string(statusOutput))) > 0 {
			fmt.Printf("   ⚠️  There are still unstaged changes:\n%s\n", string(statusOutput))

			// Try to force add everything including ignored files if needed
			fmt.Println("   🔧 Attempting to add all files...")
			forceAddCmd := exec.Command("git", "add", ".", "--force")
			forceAddCmd.Dir = repoPath
			if output, err := forceAddCmd.CombinedOutput(); err != nil {
				fmt.Printf("   ⚠️  Warning during force add: %v\nOutput: %s\n", err, output)
			}
		}
	}

	// Final check if there are changes to commit
	statusCmd = exec.Command("git", "diff", "--cached", "--quiet")
	statusCmd.Dir = repoPath
	if err := statusCmd.Run(); err == nil {
		fmt.Println("   ℹ️  No changes to commit after adding files")
		return nil
	}

	// Commit changes
	fmt.Println("   💾 Committing changes...")
	commitMsg := "feat: Configure for local development with populated registry\n\n- Added cloud-specific packages to registry/\n- Updated repoURL references to use fork\n- Ready for local GitOps workflow"
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = repoPath
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit changes: %w\nOutput: %s", err, output)
	}
	fmt.Println("   ✅ Changes committed successfully")

	// Check if fork remote exists before pushing
	remoteCmd := exec.Command("git", "remote", "-v")
	remoteCmd.Dir = repoPath
	remoteOutput, err := remoteCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check git remotes: %w\nOutput: %s", err, remoteOutput)
	}

	if !strings.Contains(string(remoteOutput), "fork") {
		return fmt.Errorf("fork remote not found. Available remotes:\n%s", string(remoteOutput))
	}

	// Push to fork using the current branch name
	fmt.Printf("   📤 Pushing to fork (branch: %s)...\n", currentBranch)

	// First try to push, if it fails because the branch doesn't exist on remote, push with --set-upstream
	pushCmd := exec.Command("git", "push", "fork", currentBranch)
	pushCmd.Dir = repoPath
	if output, err := pushCmd.CombinedOutput(); err != nil {
		// If the push failed, try with --set-upstream to create the branch on the remote
		if strings.Contains(string(output), "does not exist") || strings.Contains(string(output), "no upstream branch") {
			fmt.Printf("   🔧 Branch doesn't exist on fork, creating it...\n")
			upstreamPushCmd := exec.Command("git", "push", "--set-upstream", "fork", currentBranch)
			upstreamPushCmd.Dir = repoPath
			if upstreamOutput, upstreamErr := upstreamPushCmd.CombinedOutput(); upstreamErr != nil {
				return fmt.Errorf("failed to push to fork with --set-upstream: %w\nOutput: %s", upstreamErr, upstreamOutput)
			}
		} else {
			return fmt.Errorf("failed to push to fork: %w\nOutput: %s", err, output)
		}
	}

	fmt.Println("   ✅ Successfully pushed to fork")
	return nil
}

// getCurrentBranch gets the current git branch name with fallback handling
func getCurrentBranch(repoPath string) (string, error) {
	// Try git branch --show-current (newer git versions)
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = repoPath
	if branchOutput, err := branchCmd.CombinedOutput(); err == nil {
		branch := strings.TrimSpace(string(branchOutput))
		if branch != "" {
			return branch, nil
		}
	}

	// Fallback: try git symbolic-ref HEAD (older git versions)
	symRefCmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	symRefCmd.Dir = repoPath
	if symRefOutput, err := symRefCmd.CombinedOutput(); err == nil {
		branch := strings.TrimSpace(string(symRefOutput))
		if branch != "" {
			return branch, nil
		}
	}

	// Last fallback: try git rev-parse --abbrev-ref HEAD
	revParseCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	revParseCmd.Dir = repoPath
	if revParseOutput, err := revParseCmd.CombinedOutput(); err == nil {
		branch := strings.TrimSpace(string(revParseOutput))
		if branch != "" && branch != "HEAD" {
			return branch, nil
		}
	}

	// If all else fails, default to main/master
	return "main", fmt.Errorf("could not determine current branch, defaulting to 'main'")
}
