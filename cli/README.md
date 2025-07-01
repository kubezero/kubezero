# KubeZero CLI

A modern Go-based CLI tool for bootstrapping and managing KubeZero platform infrastructure with GitOps principles.

## 🚀 Features

- **Interactive Setup**: Friendly prompts for cloud provider and region selection
- **Multi-Cloud Support**: AWS and GCP with region-specific configurations
- **Local Development Mode**: Automated GitHub fork workflow for local development
- **Package Management**: Dynamic preparation and copying of cloud-specific packages
- **ArgoCD Integration**: Automated cluster bootstrap with ArgoCD monitoring
- **Configuration Patching**: Intelligent updates of infrastructure patches for regions and providers

## 📦 Installation

### Download from Releases

Visit the [releases page](https://github.com/kubezero/kubezero/releases) and download the appropriate binary for your platform:

#### Linux/macOS
```bash
# Download the appropriate binary for your platform
wget https://github.com/kubezero/kubezero/releases/latest/download/kubezero-linux-amd64
chmod +x kubezero-linux-amd64
sudo mv kubezero-linux-amd64 /usr/local/bin/kubezero
```

#### Windows
Download `kubezero-windows-amd64.exe` and add it to your PATH.

### Using Docker

```bash
# Run with Docker
docker run --rm -it ghcr.io/kubezero/kubezero/kubezero-cli:latest

# Mount your local directory for persistent changes
docker run --rm -it -v $(pwd):/workspace -w /workspace ghcr.io/kubezero/kubezero/kubezero-cli:latest
```

### Build from Source

```bash
git clone https://github.com/kubezero/kubezero.git
cd kubezero/cli
go build -o kubezero .
```

## 🛠️ Prerequisites

- **k3d**: For local cluster creation
- **kubectl**: For Kubernetes cluster interaction
- **Git**: For repository operations (when using `--local` mode)

## 📋 Usage

### Basic Bootstrap

```bash
# Interactive mode - prompts for cloud provider and region
kubezero bootstrap

# Specify cloud provider and region directly
kubezero bootstrap --cloud aws --region eu-west-1
kubezero bootstrap --cloud gcp --region us-central1

# Force interactive mode even with flags
kubezero bootstrap --interactive --cloud aws --region eu-west-1
```

### Local Development Mode

The `--local` flag automates the entire local development setup:

```bash
kubezero bootstrap --local --cloud aws --region eu-west-1
```

This will:
1. 📦 Prepare cloud-specific packages for your selected provider/region
2. 🌐 Open GitHub fork page in your browser
3. 👤 Prompt for your GitHub username
4. 📡 Add your fork as a git remote
5. 📝 Update all ArgoCD Application manifests to point to your fork
6. 💾 Commit and push changes to your fork
7. 🚀 Continue with standard cluster bootstrap

### Advanced Options

```bash
# Custom k3d configuration file
kubezero bootstrap --config /path/to/k3d-config.yaml --cloud aws --region eu-west-1

# Get help
kubezero --help
kubezero bootstrap --help

# Check version
kubezero version
```

## 🌍 Supported Cloud Providers

### AWS
- **Regions**: us-east-1, us-east-2, us-west-1, us-west-2, eu-west-1, eu-west-2, eu-central-1, ap-southeast-1, ap-southeast-2, ap-northeast-1

### GCP
- **Regions**: us-central1, us-east1, us-west1, us-west2, europe-west1, europe-west2, europe-central2, asia-southeast1, asia-east1, asia-northeast1

## 📁 What Gets Created

When you run the bootstrap command, the following happens:

1. **Registry Population**: Cloud-specific packages are copied from `../packages/` to `../registry/`
2. **Configuration Patching**: Infrastructure patches are updated with correct regions and availability zones
3. **k3d Cluster**: A local Kubernetes cluster is created using k3d
4. **ArgoCD Deployment**: ArgoCD is deployed and monitored until ready
5. **GitOps Setup**: (with `--local`) All manifests are configured to sync from your GitHub fork

## 🔧 Local Development Workflow

The CLI is designed to streamline local KubeZero development:

1. Run `kubezero bootstrap --local --cloud aws --region eu-west-1`
2. Complete the GitHub fork in your browser
3. Enter your GitHub username when prompted
4. The CLI handles all git operations automatically
5. Your local cluster now syncs from your personal fork
6. Make changes locally and commit to trigger ArgoCD sync

## 🐳 Container Usage

The CLI is also available as a container image with all prerequisites included:

```bash
# Interactive use
docker run --rm -it ghcr.io/kubezero/kubezero/kubezero-cli:latest

# Mount your workspace
docker run --rm -it \
  -v $(pwd):/workspace \
  -v $HOME/.kube:/home/kubezero/.kube \
  -w /workspace \
  ghcr.io/kubezero/kubezero/kubezero-cli:latest bootstrap --local --cloud aws --region eu-west-1
```

## 🎯 Next Steps After Bootstrap

After successful bootstrap, you can:

```bash
# Check pod status
kubectl get pods -n kubezero

# View ArgoCD applications
kubectl get applications -n kubezero

# Access ArgoCD UI
echo "ArgoCD URL: http://gitops.local.kubezero.io"

# Check prepared packages
ls -la ../registry/
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes in the `cli/` directory
4. Test your changes
5. Submit a pull request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](../LICENSE) file for details.

## Features

- Bootstrap a local k3d cluster with ArgoCD
- Monitor deployment status of core components
- Check cluster and pod status
- Clean up clusters when done
- Automated cluster initialization and validation

## Prerequisites

- [k3d](https://k3d.io/) installed and available in PATH
- [kubectl](https://kubernetes.io/docs/tasks/tools/) configured
- Go 1.21+ (for building from source)

## Installation

### Build from source

```bash
cd cli
go mod tidy
go build -o kubezero .
```

### Usage

#### Bootstrap a new KubeZero cluster:

```bash
./kubezero bootstrap
```

This command will:
1. Create a k3d cluster using `bootstrap/k3d-bootstrap-cluster.yaml`
2. Wait for the cluster to be ready
3. Monitor ArgoCD pods until they are healthy:
   - `argo-cd-server`
   - `argo-cd-application-controller` 
   - `argo-cd-repo-server`

#### Check cluster status:

```bash
./kubezero status
```

#### Clean up the cluster:

```bash
./kubezero cleanup
```

#### Show version:

```bash
./kubezero version
```

### Options

#### Bootstrap command
- `--config, -c`: Path to k3d cluster configuration file (default: "bootstrap/k3d-bootstrap-cluster.yaml")

#### Cleanup command  
- `--force, -f`: Force delete without confirmation

### Examples

Bootstrap with custom configuration:
```bash
./kubezero bootstrap --config /path/to/custom-config.yaml
```

Check current cluster status:
```bash
./kubezero status
```

Force cleanup without confirmation:
```bash
./kubezero cleanup --force
```

## Development

Run without building:
```bash
go run . bootstrap
```

Run tests:
```bash
go test ./...
```

## Architecture

The CLI is built using:
- [Cobra](https://github.com/spf13/cobra) for command-line interface
- [Kubernetes client-go](https://github.com/kubernetes/client-go) for cluster interaction
- Native Go exec for k3d commands

## Troubleshooting

- Ensure k3d is installed: `k3d version`
- Check cluster status: `kubectl get nodes`
- Monitor pods: `kubectl get pods -n kubezero`
- View ArgoCD logs: `kubectl logs -n kubezero deployment/argo-cd-server`
