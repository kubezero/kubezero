package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// ── Cloud config ──────────────────────────────────────────────────────────────

type cloudConfig struct {
	label   string
	regions []string
}

var clouds = map[string]cloudConfig{
	"aws": {
		label: "AWS (Amazon Web Services)",
		regions: []string{
			"us-east-1", "us-east-2", "us-west-1", "us-west-2",
			"eu-west-1", "eu-west-2", "eu-central-1",
			"ap-southeast-1", "ap-southeast-2", "ap-northeast-1",
		},
	},
	"gcp": {
		label: "GCP (Google Cloud Platform)",
		regions: []string{
			"us-central1", "us-east1", "us-west1", "us-west2",
			"europe-west1", "europe-west2", "europe-central2",
			"asia-southeast1", "asia-east1", "asia-northeast1",
		},
	},
	"digitalocean": {
		label: "DigitalOcean (DOKS)",
		regions: []string{
			"nyc1", "nyc3", "sfo2", "sfo3",
			"ams3", "fra1", "lon1",
			"sgp1", "blr1", "syd1",
		},
	},
	"virtual": {
		label: "Virtual (vCluster — no cloud required)",
		regions: []string{"in-cluster"},
	},
}

var cloudKeys = []string{"aws", "gcp", "digitalocean", "virtual"}

// ── Topology definitions ──────────────────────────────────────────────────────

type topology struct {
	label       string
	description string
	// roles lists the cluster slots the user must fill, in order.
	// Each slot is "<role>", e.g. "management", "worker", "worker (prod)", etc.
	roles []string
}

var topologies = []topology{
	{
		label:       "Single cluster (all-in-one)",
		description: "One management cluster; staging/dev run as vClusters inside it",
		roles:       []string{"management"},
	},
	{
		label:       "Two clusters",
		description: "Management cluster + one worker cluster",
		roles:       []string{"management", "worker"},
	},
	{
		label:       "Three clusters (recommended)",
		description: "Management + prod worker + non-prod worker (staging/dev as vClusters)",
		roles:       []string{"management", "worker (prod)", "worker (non-prod)"},
	},
	{
		label:       "Four clusters",
		description: "Management + prod + staging + dev (all physical clusters)",
		roles:       []string{"management", "worker (prod)", "worker (staging)", "worker (dev)"},
	},
}

// ── Command ───────────────────────────────────────────────────────────────────

var repoRoot string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactively configure your KubeZero platform topology",
	Long: `init guides you through selecting a deployment topology (from the five patterns
in docs/patterns.md) and choosing a cloud provider for each cluster role.
It then copies the correct packages from packages/ into registry/ so ArgoCD
can pick them up on the next sync.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&repoRoot, "repo", "..", "Path to the kubezero repository root")
}

func runInit(_ *cobra.Command, _ []string) error {
	fmt.Println("Welcome to KubeZero init")
	fmt.Println()

	// 1. Select topology
	topo, err := selectTopology()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Topology: %s\n", topo.label)
	fmt.Printf("Cluster roles to configure: %v\n", topo.roles)
	fmt.Println()

	// 2. For each role, select cloud + region → resolve package name
	type selection struct {
		role        string
		cloud       string
		region      string
		packageName string
	}
	var selections []selection

	for _, role := range topo.roles {
		fmt.Printf("── Configure cluster: %s ──\n", role)

		cloud, err := selectCloud()
		if err != nil {
			return err
		}

		region, err := selectRegion(cloud)
		if err != nil {
			return err
		}

		pkgName := resolvePackageName(cloud, role)
		fmt.Printf("  Package: %s (region: %s)\n\n", pkgName, region)

		selections = append(selections, selection{
			role:        role,
			cloud:       cloud,
			region:      region,
			packageName: pkgName,
		})
	}

	// 3. Confirm
	fmt.Println("Summary")
	fmt.Println("-------")
	for _, s := range selections {
		fmt.Printf("  %-22s  %s  (%s)\n", s.role, s.packageName, s.region)
	}
	fmt.Println()

	if !confirm("Activate these packages in registry/?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// 4. Copy packages to registry
	registryDir := filepath.Join(repoRoot, "registry")
	packagesDir := filepath.Join(repoRoot, "packages")

	for _, s := range selections {
		src := filepath.Join(packagesDir, s.packageName)
		dst := filepath.Join(registryDir, s.packageName)

		if _, err := os.Stat(dst); err == nil {
			fmt.Printf("  skip   %s  (already exists in registry)\n", s.packageName)
			continue
		}

		if err := copyDir(src, dst); err != nil {
			return fmt.Errorf("failed to copy %s: %w", s.packageName, err)
		}
		fmt.Printf("  added  registry/%s\n", s.packageName)
	}

	fmt.Println()
	fmt.Println("Done. Commit and push to trigger ArgoCD sync.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Create cloud credential secrets in the bootstrap cluster (see docs/getting-started.md)")
	fmt.Println("  2. Run: kubezero bootstrap")
	fmt.Println("  3. ArgoCD will provision your clusters automatically")

	return nil
}

// ── Prompts ───────────────────────────────────────────────────────────────────

func selectTopology() (topology, error) {
	items := make([]string, len(topologies))
	for i, t := range topologies {
		items[i] = fmt.Sprintf("%-36s  %s", t.label, t.description)
	}

	p := promptui.Select{
		Label: "Select deployment topology",
		Items: items,
		Size:  len(items),
	}
	i, _, err := p.Run()
	if err != nil {
		return topology{}, err
	}
	return topologies[i], nil
}

func selectCloud() (string, error) {
	items := make([]string, len(cloudKeys))
	for i, k := range cloudKeys {
		items[i] = clouds[k].label
	}

	p := promptui.Select{
		Label: "Select cloud provider",
		Items: items,
		Size:  len(items),
	}
	i, _, err := p.Run()
	if err != nil {
		return "", err
	}
	return cloudKeys[i], nil
}

func selectRegion(cloud string) (string, error) {
	regions := clouds[cloud].regions

	p := promptui.Select{
		Label: "Select region",
		Items: regions,
		Size:  10,
	}
	_, region, err := p.Run()
	return region, err
}

func confirm(label string) bool {
	p := promptui.Prompt{
		Label:     label + " [y/N]",
		IsConfirm: true,
	}
	result, err := p.Run()
	if err != nil {
		return false
	}
	return result == "y" || result == "Y"
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// resolvePackageName maps (cloud, role) → packages/ directory name.
// Worker roles all map to the "<cloud>-worker" package regardless of the
// descriptive label (prod/staging/dev differentiation is handled by the user
// patching region/name after init).
func resolvePackageName(cloud, role string) string {
	if role == "management" {
		return cloud + "-management"
	}
	return cloud + "-worker"
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
