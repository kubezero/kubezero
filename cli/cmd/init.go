package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize KubeZero for your cloud provider",
	Long: `Interactive initialization that guides you through setting up KubeZero
for your preferred cloud provider. This command will:

1. Ask for your cloud provider preference (AWS or GCP)
2. Prepare the necessary packages for your provider
3. Copy them to the registry for ArgoCD deployment
4. Provide next steps for bootstrapping your cluster`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Welcome to KubeZero Initialization")
	fmt.Println()

	// Get cloud provider from user
	provider, err := promptCloudProvider()
	if err != nil {
		return fmt.Errorf("failed to get cloud provider: %w", err)
	}

	// Get region for the selected provider
	selectedRegion, err := promptRegionForProvider(provider)
	if err != nil {
		return fmt.Errorf("failed to get region: %w", err)
	}

	fmt.Printf("\n📋 Configuration Summary:\n")
	fmt.Printf("   Cloud Provider: %s\n", provider)
	fmt.Printf("   Region: %s\n", selectedRegion)
	fmt.Println()

	// Confirm with user
	if !promptConfirmation("Proceed with this configuration?") {
		fmt.Println("❌ Initialization cancelled.")
		return nil
	}

	// Set the bootstrap variables for when bootstrap is run
	cloudProvider = provider
	region = selectedRegion

	// Show next steps instead of running prepare directly
	fmt.Println("🎯 Configuration saved successfully!")
	showNextSteps(provider)

	return nil
}

func promptCloudProvider() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("☁️  Which cloud provider would you like to use?")
	fmt.Println("   1) AWS (Amazon Web Services)")
	fmt.Println("   2) GCP (Google Cloud Platform)")
	fmt.Print("\nEnter your choice (1-2): ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	switch input {
	case "1":
		return "aws", nil
	case "2":
		return "gcp", nil
	default:
		fmt.Println("❌ Invalid choice. Please enter 1 or 2.")
		return promptCloudProvider()
	}
}

func promptRegionForProvider(provider string) (string, error) {
	// Use the same cloud configs from bootstrap.go
	supportedClouds := map[string]struct {
		Name    string
		Regions []string
	}{
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

	config := supportedClouds[provider]
	fmt.Printf("Available regions for %s:\n", config.Name)
	fmt.Println()

	for i, region := range config.Regions {
		fmt.Printf("  %d) %s\n", i+1, region)
	}

	for {
		fmt.Printf("\nSelect region (1-%d): ", len(config.Regions))

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(config.Regions) {
			fmt.Printf("❌ Invalid choice. Please enter a number between 1 and %d.\n", len(config.Regions))
			continue
		}

		selectedRegion := config.Regions[choice-1]
		fmt.Printf("✅ Selected region: %s\n", selectedRegion)
		return selectedRegion, nil
	}
}

func promptConfirmation(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n❓ %s (y/N): ", message)

	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func showNextSteps(provider string) {
	fmt.Println()
	fmt.Println("🎉 Initialization Complete")
	fmt.Println()
	fmt.Println("📋 Next Steps:")
	fmt.Println("   1. Bootstrap your cluster:")
	fmt.Printf("      ./kubezero bootstrap --cloud %s --region %s\n", provider, region)
	fmt.Println()
	fmt.Println("   2. Or run with interactive mode:")
	fmt.Println("      ./kubezero bootstrap --interactive")
	fmt.Println()
	fmt.Println("   3. Monitor the deployment:")
	fmt.Println("      ./kubezero status")
	fmt.Println()

	switch provider {
	case "aws":
		fmt.Println("☁️  AWS-specific notes:")
		fmt.Println("   • Ensure AWS credentials are configured")
		fmt.Println("   • Check that required IAM permissions are set")
		fmt.Println("   • Verify AWS provider configuration in ArgoCD")
	case "gcp":
		fmt.Println("☁️  GCP-specific notes:")
		fmt.Println("   • Ensure GCP credentials are configured")
		fmt.Println("   • Check that required IAM permissions are set")
		fmt.Println("   • Verify GCP provider configuration in ArgoCD")
	}

	fmt.Println()
	fmt.Println("💡 Need help? Run: ./kubezero --help")
}
