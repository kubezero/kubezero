# KubeZero Management EKS Package

This package provisions the AWS EKS management cluster using Crossplane, following KubeZero's modular, GitOps-native structure.

## Structure
- `infrastructure/`: Crossplane manifests for AWS EKS, VPC, NodeGroup, ProviderConfig, etc.
- `kustomization.yaml`: Kustomize entrypoint for the package, referencing infrastructure and stacks.
- `gitops.yaml`: ArgoCD Application/AppProject for GitOps enablement.

## Usage
- Copy or symlink this package to the `registry/management` directory to enable it.
- Ensure AWS credentials are available as a Kubernetes Secret referenced by ProviderConfig.

## References
- [KubeZero Directory Structure](../docs/structure.md)
- [Crossplane AWS Provider](https://github.com/crossplane-contrib/provider-aws)
