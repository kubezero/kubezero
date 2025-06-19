# KubeZero Controller Directory

This directory contains the manifests for the GitOps controller stack used by KubeZero to bootstrap and manage real clusters on cloud providers or local environments. It is typically deployed in the local k3d bootstrap cluster via ArgoCD.

## Purpose
- Provides the core building blocks for cluster bootstrapping: ArgoCD, Crossplane, and External Secrets.
- Enables automated provisioning and management of cloud or local clusters using GitOps workflows.

## Structure
- `argo-cd/`: ArgoCD controller and project setup.
- `crossplane/`: Crossplane controller and configuration for cloud provider integration.
- `external-secrets/`: External Secrets operator and secret store setup.
- `gitops/`: GitOps project and application definitions.
- `namespace/`: Namespace definitions for controller components.
- `kustomization.yaml`: Main Kustomize entrypoint for the controller stack.

## Usage
Running the KubeZero bootstrap command will create this GitOps controller (with ArgoCD, Crossplane, and External Secrets) out of the box, ready for KubeZero to start provisioning clusters:

```shell
k3d cluster create --config bootstrap/k3d-bootstrap-cluster.yaml
```

---
For more details, see the [structure.md](../docs/structure.md) file in the docs directory.
