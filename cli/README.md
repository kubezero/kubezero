# KubeZero CLI

A command-line tool to bootstrap and manage KubeZero platform infrastructure.

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
