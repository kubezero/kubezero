# KubeZero Directory Structure and Responsibilities

A clearly defined structure to support scalable, modular, GitOps-native Kubernetes infrastructure management.

---

## Directory Responsibilities

### `modules/` → **Reusable building blocks (primitives)**

> 💡 _Think: Terraform modules, but for Kubernetes manifests_

- Granular, self-contained Kustomize or Helm bundles (e.g. `argocd`, `cert-manager`)
- Not environment- or cluster-specific
- Designed to be imported into any stack
- May include:
  - Helm chart values
  - Kustomize overlays
  - Namespaces, CR templates, RBAC

**Example use:**

```yaml
# stacks/management/manifests/kustomization.yaml
resources:
  - ../../../modules/cert-manager
```

---

### `stacks/` → **Cluster/environment-specific composition of modules**

> 💡 _Think: logical environments or platform configurations_

- Represents platform stacks like `management`, `vcluster`, `production`
- Aggregates modules using Kustomize
- Contains:
  - `manifests/` — references to `modules/`
  - `overlays/` — environment-specific configuration
- **Avoid raw manifests** — always reference `modules/`

---

### `registry/` → **Runtime GitOps registry (ArgoCD control)**

> 💡 _Think: GitOps deployment targets_

- Reflects what is actually deployed or tracked by ArgoCD
- Maps environments like `management`, `preview`, `production`
- Contains:
  - `_gitops.yaml`: list of ArgoCD applications
  - `_kustomization.yaml`: ArgoCD setup config
- Should **only reference**, not define infrastructure

---

### `bootstrap/` → **Local one-time K3D-based bootstrapper**

> 💡 _Think: Terraform bootstrapper in Kubernetes_

- Sets up the initial local cluster
- Installs ArgoCD, Crossplane
- Kicks off GitOps loop via `controller/`
- Disposable after management cluster is online

---

### `controller/` → **Bootstrap ArgoCD applications**

- Defines ArgoCD `Application` & `AppProject` CRs
- Points to entries in `registry/` or `stacks/`
- Should **not contain manifests**, only GitOps references

---

## 🛠 Best Practice: No Duplication in Stacks

❌ Don’t create repeated directories like:

```
stacks/management/manifests/argocd/
```

✅ Instead, always reference shared modules:

```yaml
# stacks/management/manifests/kustomization.yaml
resources:
  - ../../../modules/argocd
  - ../../../modules/cert-manager
  - ../../../modules/crossplane/controller
```

This keeps your stacks thin, readable, and DRY.

---

## Multi-Namespace and Multi-Environment Overlays

Kustomize supports setting up resources in different namespaces and environments using overlays and bases. This is particularly useful for creating isolated environments like `dev`, `staging`, and `production`.

Referencing the [Kustomize multi-namespace example](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/multibases/multi-namespace.md), you can structure your overlays and bases as follows:

### Example Directory Structure

```
stacks/
  vcluster/
    base/
      kustomization.yaml
      namespace.yaml
      deployment.yaml
    overlays/
      dev/
        kustomization.yaml
        namespace.yaml
      staging/
        kustomization.yaml
        namespace.yaml
```

### Base Example

```yaml
# filepath: kubezero/stacks/vcluster/base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - namespace.yaml
  - deployment.yaml
```

### Overlay Example for `dev`

```yaml
# filepath: kubezero/stacks/vcluster/overlays/dev/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - namespace.yaml
```

```yaml
# filepath: kubezero/stacks/vcluster/overlays/dev/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: vcluster-dev
```

### Overlay Example for `staging`

```yaml
# filepath: kubezero/stacks/vcluster/overlays/staging/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - namespace.yaml
```

```yaml
# filepath: kubezero/stacks/vcluster/overlays/staging/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: vcluster-staging
```

This approach allows you to reuse the same base resources while customizing namespaces and other environment-specific configurations in overlays.

---

## Summary Table

| Directory     | Responsibility                                         | Contains                          |
|---------------|--------------------------------------------------------|------------------------------------|
| `modules/`     | Reusable IaC components (cert-manager, argo-cd, etc.) | Helm/kustomize + namespaces       |
| `stacks/`      | Environment/cluster-specific composition of modules   | References to modules + overlays  |
| `registry/`    | Deployment targets for ArgoCD (actual environments)   | GitOps entrypoints (app list)     |
| `bootstrap/`   | Local single-use cluster to kick off KubeZero         | Minimal manifests + k3d YAML      |
| `controller/`  | ArgoCD `Application` definitions used by bootstrap    | App CRs pointing to registry/stacks |

---
