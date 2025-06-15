# KubeZero Management AWS Package

This package provisions the AWS EKS management cluster using Crossplane, following KubeZero's modular, GitOps-native structure.

## Structure
- `infrastructure/`: Crossplane manifests for AWS EKS, VPC, NodeGroup, ProviderConfig, etc.
- `kustomization.yaml`: Kustomize entrypoint for the package, referencing infrastructure and stacks.
- `gitops.yaml`: ArgoCD Application/AppProject for GitOps enablement.

## Usage
- Copy or symlink this package to the `registry/management` directory to enable it.
- Ensure AWS credentials are available as a Kubernetes Secret referenced by ProviderConfig.

## Parameterization with Kustomize

This package uses Kustomize variables for AWS-specific and cluster configuration values. To set or override these values:

**Edit `kustomization.yaml`**
   - In the `configMapGenerator.literals` section, update the following variables with your environment-specific values:
     - `AWS_ACCOUNT_ID`
     - `VPC_ID`
     - `SUBNET_RESOURCE_1`, `SUBNET_RESOURCE_2`
     - `AWS_REGION`
     - `EKS_VERSION`
     - `NODE_SIZE`
     - `MIN_NODE_COUNT`, `MAX_NODE_COUNT`

   Example:
   ```yaml
   configMapGenerator:
     - name: argocd-params
       literals:
         - AWS_ACCOUNT_ID=123456789012
         - VPC_ID=vpc-xyz789
         - SUBNET_RESOURCE_1=subnet-abc123
         - SUBNET_RESOURCE_2=subnet-def456
         - AWS_REGION=us-west-1
         - EKS_VERSION=1.29
         - NODE_SIZE=medium
         - MIN_NODE_COUNT=2
         - MAX_NODE_COUNT=5
   ```

## References
- [KubeZero Directory Structure](../docs/structure.md)
- [Crossplane AWS Provider](https://github.com/crossplane-contrib/provider-aws)
