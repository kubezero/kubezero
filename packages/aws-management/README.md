# KubeZero AWS Management Package

This package provides a GitOps-ready Crossplane setup for managing an EKS cluster and its supporting AWS infrastructure on AWS, using Upbound's platform resources and Kustomize.

## Structure
- `infrastructure/xeks.yaml`: Defines the EKS cluster using the `XEKS` kind. Edit this file to set cluster version, node count, instance type, and other EKS parameters.
- `infrastructure/xnetwork.yaml`: Defines the VPC and subnets using the `XNetwork` kind. Edit this file to set VPC CIDR, subnet CIDRs, and availability zones.
- `infrastructure/kustomization.yaml`: Kustomize entrypoint for infrastructure resources.
- `provider/`: Contains Crossplane provider configuration and the Upbound configuration package manifest.
- `kustomization.yaml`: Top-level Kustomize entrypoint for the package.
- `gitops.yaml`: ArgoCD AppProject and Application definitions for GitOps deployment.

## Usage

To use this package:

1. **Copy the package**
   - Copy the entire `aws-management` folder into your `registry` directory (e.g., `registry/aws-management`).
     ```shell
     cp -a packages/aws-management registry/aws-management
     ```
   - (Optional) Commit it to your repository:
     ```shell
     git add registry/aws-management
     git commit -m "feat: enable aws-management package"
     ```

2. **Configure AWS Provider Credentials**
   - Ensure a Kubernetes secret named `aws-creds` exists in the `crossplane-system` namespace with your AWS credentials.
   - The provider config is defined in `provider/provider-config.yaml`.

3. **Edit EKS and Network Parameters**
   - Update `infrastructure/xeks.yaml` for EKS cluster settings (version, node count, instance type, etc).
   - Update `infrastructure/xnetwork.yaml` for VPC and subnet settings (CIDR blocks, availability zones, etc).

4. **Deploy with ArgoCD or Kustomize**
   - **ArgoCD**: Point your Application at the `registry/aws-management` directory or use the provided `gitops.yaml`.
   - **Kustomize**: From the package root, run:
     ```sh
     kustomize build . | kubectl apply -f -
     ```
## Notes
- All configuration is done by editing the YAML manifests directly.
- The Upbound configuration package is referenced in `provider-config/configuration.yaml`.
- For advanced customization, refer to the Upbound [configuration-aws-eks documentation](https://marketplace.upbound.io/providers/upbound/configuration-aws-eks).
