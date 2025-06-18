# KubeZero AWS Workers Package

This package provides a GitOps-ready Crossplane setup for managing an EKS cluster and its supporting AWS infrastructure on AWS, by importing the pre-configured `modules/aws/eks` module. All necessary AWS provider configuration and resources (XEKS, XNetwork) are handled by the imported module.

## Structure
- Imports `modules/aws/eks` for all AWS EKS and network resources.
- `kustomization.yaml`: Top-level Kustomize entrypoint for the package.
- `gitops.yaml`: ArgoCD AppProject and Application definitions for GitOps deployment.

## Usage

To use this package:

1. **Create the AWS Credentials Secret**
   - Ensure a Kubernetes secret named `aws-creds` exists in the `crossplane-system` namespace with your AWS credentials. For example:
     ```shell
     kubectl create secret generic aws-creds \
       -n crossplane-system \
       --from-literal=creds='{"accessKeyId":"<YOUR_ACCESS_KEY>","secretAccessKey":"<YOUR_SECRET_KEY>"}'
     ```

2. **Copy the package**
   - Copy the entire `aws-workers` folder into your `registry` directory (e.g., `registry/aws-workers`).
     ```shell
     cp -a packages/aws-workers registry/aws-workers
     ```
   - Commit it to your repository:
     ```shell
     git add registry/aws-workers
     git commit -m "feat: enable aws-workers package"
     ```

3. **Deploy with ArgoCD or Kustomize**
   - **ArgoCD**: Point your Application at the `registry/aws-workers` directory or use the provided `gitops.yaml`.
   - **Kustomize**: From the package root, run:
     ```shell
     kustomize build . | kubectl apply -f -
     ```

## Notes
- All AWS provider configuration and resource manifests are managed by the imported `modules/aws/eks` module.
- For advanced customization, edit the `modules/aws/eks` module or override its parameters as needed.
- For more details, see the KubeZero documentation and the Upbound [configuration-aws-eks documentation](https://marketplace.upbound.io/providers/upbound/configuration-aws-eks).
