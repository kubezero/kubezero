# KubeZero GCP Workers Package

This package provides a GitOps-ready Crossplane setup for managing an GKE cluster and its supporting GCP infrastructure on GCP, by importing the pre-configured `modules/gcp/gke` module. All necessary GCP provider configuration and resources (XGKE, XNetwork) are handled by the imported module.

## Structure
- Imports `modules/gcp/gke` for all GCP GKE and network resources.
- `kustomization.yaml`: Top-level Kustomize entrypoint for the package.
- `gitops.yaml`: ArgoCD AppProject and Application definitions for GitOps deployment.

## Usage

To use this package:

1. **Create the GCP Credentials Secret**
   - Ensure a Kubernetes secret named `gcp-creds` exists in the `crossplane-system` namespace with your GCP credentials. For example:
    kubectl create secret generic gcp-creds \
      -n crossplane-system \
      --from-file=creds=</path/to/your/gcp-service-account.json>
     ```

2. **Copy the package**
   - Copy the entire `gcp-workers` folder into your `registry` directory (e.g., `registry/gcp-workers`).
     ```shell
     cp -a packages/gcp-workers registry/gcp-workers
     ```
   - Commit it to your repository:
     ```shell
     git add registry/gcp-workers
     git commit -m "feat: enable gcp-workers package"
     ```

3. **Deploy with ArgoCD or Kustomize**
   - **ArgoCD**: Point your Application at the `registry/gcp-workers` directory or use the provided `gitops.yaml`.
   - **Kustomize**: From the package root, run:
     ```shell
     kustomize build . | kubectl apply -f -
     ```

## Notes
- All GCP provider configuration and resource manifests are managed by the imported `modules/gcp/gke` module.
- For advanced customization, edit the `modules/gcp/gke` module or override its parameters as needed.
- For more details, see the KubeZero documentation and the Upbound [configuration-gcp-gke documentation](https://marketplace.upbound.io/providers/upbound/configuration-gcp-gke).
