# Bootstrap KubeZero

The bootstrap process uses the same pattern used in [Crossplane Bootstrapper](https://github.com/DevOpsHiveHQ/crossplane-bootstrapper) but it will load ArgoCD and Crossplane to build the management cluster.

## Purpose
The bootstrap directory contains the manifests for the GitOps controller stack used by KubeZero to bootstrap and manage real clusters on cloud providers or local environments. It is typically deployed in the local k3d bootstrap cluster via ArgoCD.

The controller stack provides:
- Core building blocks for cluster bootstrapping: ArgoCD, Crossplane, and External Secrets.
- Automated provisioning and management of cloud or local clusters using GitOps workflows.

## Structure
- `k3d-bootstrap-cluster.yaml`: K3d cluster configuration for the bootstrap environment.
- `kubezero-bootstrap-manifests.yaml`: Bootstrap manifests for initial controller deployment.
- `kustomization.yaml`: Main Kustomize entrypoint for the bootstrap stack.

The controller directory contains:
- `argo-cd/`: ArgoCD controller and project setup.
- `crossplane/`: Crossplane controller and configuration for cloud provider integration.
- `external-secrets/`: External Secrets operator and secret store setup.
- `gitops/`: GitOps project and application definitions.
- `namespace/`: Namespace definitions for controller components.

## Usage
Running the KubeZero bootstrap command will create this GitOps controller (with ArgoCD, Crossplane, and External Secrets) out of the box, ready for KubeZero to start provisioning clusters:

```shell
k3d cluster create --config bootstrap/k3d-bootstrap-cluster.yaml
```

This will set up the local k3d bootstrap cluster with all necessary controllers to manage and provision real clusters on cloud providers.

---
For more details, see the [structure.md](../docs/structure.md) file in the docs directory.
