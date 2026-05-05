package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// ── scaffold templates ────────────────────────────────────────────────────────

// scaffoldTemplates defines the files generated for each cloud provider.
// Keys are file paths relative to the new package root.
type scaffoldTemplate struct {
	gitops         string
	infraKust      string
	appKust        string
	extraFiles     map[string]string // extra files relative to infrastructure/
}

var scaffoldTemplates = map[string]scaffoldTemplate{
	"aws": {
		gitops: `---
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: {name}
spec:
  description: {role} EKS cluster resources
  clusterResourceWhitelist:
    - group: '*'
      kind: '*'
  destinations:
    - namespace: '*'
      server: '*'
  sourceRepos:
    - https://github.com/kubezero/kubezero
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {name}
  annotations:
    argocd.argoproj.io/sync-wave: '100'
    argocd.argoproj.io/depends-on: aws-provider
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: {name}
  source:
    repoURL: https://github.com/kubezero/kubezero
    path: registry/{name}/infrastructure
    targetRevision: main
  destination:
    name: in-cluster
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - SkipDryRunOnMissingResource=true
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {name}-applications
  annotations:
    argocd.argoproj.io/sync-wave: '100'
    argocd.argoproj.io/depends-on: aws-provider
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: {name}
  source:
    repoURL: https://github.com/kubezero/kubezero
    path: registry/{name}/applications
    targetRevision: main
  destination:
    name: {name}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - SkipDryRunOnMissingResource=true
`,
		infraKust: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: {name}-

resources:
  - ../../../modules/aws/eks/provider
  - ../../../stacks/eks-cluster

patches:
  - path: patch-xeks.yaml
    target:
      kind: XEKS
      name: aws-eks
  - path: patch-xnetwork.yaml
    target:
      kind: XNetwork
      name: aws-network
`,
		extraFiles: map[string]string{
			"patch-xeks.yaml": `---
apiVersion: aws.platform.upbound.io/v1alpha1
kind: XEKS
metadata:
  name: aws-eks
spec:
  providerConfigName: {name}-provider-aws
  writeConnectionSecretToRef:
    name: {name}-aws-eks-kubeconfig
    namespace: crossplane-system
`,
			"patch-xnetwork.yaml": `---
apiVersion: aws.platform.upbound.io/v1alpha1
kind: XNetwork
metadata:
  name: aws-network
spec:
  parameters:
    providerConfigName: {name}-provider-aws
`,
		},
	},
	"gcp": {
		gitops: `---
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: {name}
spec:
  description: {role} GKE cluster resources
  clusterResourceWhitelist:
    - group: '*'
      kind: '*'
  destinations:
    - namespace: '*'
      server: '*'
  sourceRepos:
    - https://github.com/kubezero/kubezero
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {name}
  annotations:
    argocd.argoproj.io/sync-wave: '100'
    argocd.argoproj.io/depends-on: gcp-provider
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: {name}
  source:
    repoURL: https://github.com/kubezero/kubezero
    path: registry/{name}/infrastructure
    targetRevision: main
  destination:
    name: in-cluster
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - SkipDryRunOnMissingResource=true
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {name}-applications
  annotations:
    argocd.argoproj.io/sync-wave: '100'
    argocd.argoproj.io/depends-on: gcp-provider
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: {name}
  source:
    repoURL: https://github.com/kubezero/kubezero
    path: registry/{name}/applications
    targetRevision: main
  destination:
    name: {name}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - SkipDryRunOnMissingResource=true
`,
		infraKust: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: {name}-

resources:
  - ../../../modules/gcp/gke/provider
  - ../../../stacks/gke-cluster

patches:
  - path: patch-xgke.yaml
    target:
      kind: XGKE
      name: gcp-gke
`,
		extraFiles: map[string]string{
			"patch-xgke.yaml": `---
apiVersion: gcp.platform.upbound.io/v1alpha1
kind: XGKE
metadata:
  name: gcp-gke
spec:
  parameters:
    providerConfigName: {name}-provider-gcp
  writeConnectionSecretToRef:
    name: {name}-gcp-gke-kubeconfig
    namespace: crossplane-system
`,
		},
	},
	"digitalocean": {
		gitops: `---
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: {name}
spec:
  description: {role} DOKS cluster resources
  clusterResourceWhitelist:
    - group: '*'
      kind: '*'
  destinations:
    - namespace: '*'
      server: '*'
  sourceRepos:
    - https://github.com/kubezero/kubezero
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {name}
  annotations:
    argocd.argoproj.io/sync-wave: '100'
    argocd.argoproj.io/depends-on: digitalocean-provider
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: {name}
  source:
    repoURL: https://github.com/kubezero/kubezero
    path: registry/{name}/infrastructure
    targetRevision: main
  destination:
    name: in-cluster
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - SkipDryRunOnMissingResource=true
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {name}-applications
  annotations:
    argocd.argoproj.io/sync-wave: '100'
    argocd.argoproj.io/depends-on: digitalocean-provider
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: {name}
  source:
    repoURL: https://github.com/kubezero/kubezero
    path: registry/{name}/applications
    targetRevision: main
  destination:
    name: {name}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - SkipDryRunOnMissingResource=true
`,
		infraKust: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: {name}-

resources:
  - ../../../modules/digitalocean/provider
  - ../../../stacks/doks-cluster
`,
		extraFiles: map[string]string{},
	},
	"virtual": {
		gitops: `---
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: {name}
spec:
  description: {role} vCluster resources
  clusterResourceWhitelist:
    - group: '*'
      kind: '*'
  destinations:
    - namespace: '*'
      server: '*'
  sourceRepos:
    - https://github.com/kubezero/kubezero
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {name}
  annotations:
    argocd.argoproj.io/sync-wave: '100'
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: {name}
  source:
    repoURL: https://github.com/kubezero/kubezero
    path: registry/{name}/infrastructure
    targetRevision: main
  destination:
    name: in-cluster
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - SkipDryRunOnMissingResource=true
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {name}-applications
  annotations:
    argocd.argoproj.io/sync-wave: '100'
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: {name}
  source:
    repoURL: https://github.com/kubezero/kubezero
    path: registry/{name}/applications
    targetRevision: main
  destination:
    name: {name}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - SkipDryRunOnMissingResource=true
`,
		infraKust: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../../stacks/virtual-cluster
  - namespace.yaml
`,
		extraFiles: map[string]string{
			"namespace.yaml": `---
apiVersion: v1
kind: Namespace
metadata:
  name: {name}
`,
		},
	},
}

var appKustManagement = `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: {name}-

resources:
  - ../../../stacks/k8s-essentials/helm-chart
  - ../../../controller
`

var appKustWorker = `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: {name}-

resources:
  - ../../../stacks/k8s-essentials/helm-chart
`

// ── Command ───────────────────────────────────────────────────────────────────

var (
	scaffoldCloud string
	scaffoldRole  string
	scaffoldForce bool
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate a new package from a template",
	Long: `scaffold creates a new package under packages/ for the given cloud and role.
It generates all required files following the same conventions as the built-in packages.

Examples:
  kubezero scaffold --cloud aws --role worker
  kubezero scaffold --cloud digitalocean --role management
  kubezero scaffold --cloud virtual --role worker`,
	RunE: runScaffold,
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)
	scaffoldCmd.Flags().StringVar(&scaffoldCloud, "cloud", "", "Cloud provider (aws, gcp, digitalocean, virtual)")
	scaffoldCmd.Flags().StringVar(&scaffoldRole, "role", "", "Cluster role (management or worker)")
	scaffoldCmd.Flags().BoolVar(&scaffoldForce, "force", false, "Overwrite if package already exists")
	scaffoldCmd.Flags().StringVar(&repoRoot, "repo", "..", "Path to the kubezero repository root")
	_ = scaffoldCmd.MarkFlagRequired("cloud")
	_ = scaffoldCmd.MarkFlagRequired("role")
}

func runScaffold(_ *cobra.Command, _ []string) error {
	cloud := strings.ToLower(scaffoldCloud)
	role := strings.ToLower(scaffoldRole)

	if _, ok := clouds[cloud]; !ok {
		return fmt.Errorf("unknown cloud %q — valid options: aws, gcp, digitalocean, virtual", cloud)
	}
	if role != "management" && role != "worker" {
		return fmt.Errorf("role must be 'management' or 'worker', got %q", role)
	}

	tmpl, ok := scaffoldTemplates[cloud]
	if !ok {
		return fmt.Errorf("no scaffold template for cloud %q", cloud)
	}

	name := cloud + "-" + role
	pkgDir := filepath.Join(repoRoot, "packages", name)

	if _, err := os.Stat(pkgDir); err == nil {
		if !scaffoldForce {
			return fmt.Errorf("package %s already exists — use --force to overwrite", name)
		}
		fmt.Printf("Overwriting existing package %s\n", name)
	}

	// Render all template variables
	render := func(s string) string {
		s = strings.ReplaceAll(s, "{name}", name)
		s = strings.ReplaceAll(s, "{role}", role)
		s = strings.ReplaceAll(s, "{cloud}", cloud)
		return s
	}

	appKust := appKustWorker
	if role == "management" {
		appKust = appKustManagement
	}

	files := map[string]string{
		"gitops.yaml":                        render(tmpl.gitops),
		"infrastructure/kustomization.yaml":  render(tmpl.infraKust),
		"applications/kustomization.yaml":    render(appKust),
	}
	for relPath, content := range tmpl.extraFiles {
		files["infrastructure/"+relPath] = render(content)
	}

	for relPath, content := range files {
		fullPath := filepath.Join(pkgDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", relPath, err)
		}
		fmt.Printf("  wrote  packages/%s/%s\n", name, relPath)
	}

	fmt.Println()
	fmt.Printf("Package packages/%s created.\n", name)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Review packages/%s/ and adjust any values\n", name)
	fmt.Printf("  2. Activate it:  make registry-add PACKAGE=%s\n", name)
	fmt.Println("  3. Commit and push — ArgoCD will pick it up automatically")

	return nil
}
