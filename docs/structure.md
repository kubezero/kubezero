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
  - Namespaces, CR templates, RBAC

**Example use:**

```yaml
# stacks/management/manifests/kustomization.yaml
resources:
  - ../../../modules/cert-manager
```
---

### `packages/` → **Environment/cluster-specific composition of stacks**

> 💡 _Think: higher-level bundles for specific platforms or customers_

- Groups together one or more `stacks/` to represent a complete platform, customer, or use-case
- Useful for multi-cluster, multi-tenant, or SaaS scenarios
- Contains:
  - References to `stacks/` directories
  - Optional package-level configuration
- **Never contains raw manifests** — always references `stacks/`

---

### `stacks/` → **Cluster/environment-specific composition of modules**

> 💡 _Think: logical environments or platform configurations_

- Represents platform stacks like `management`, `vcluster`, `production`
- Aggregates modules using Kustomize
- Contains:
  - `manifests/` — references to `modules/`
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

## Summary Table

| Directory     | Responsibility                                         | Contains                          |
|---------------|--------------------------------------------------------|------------------------------------|
| `modules/`     | Reusable IaC components (cert-manager, argo-cd, etc.) | Helm/kustomize + namespaces       |
| `packages/`    | Environment/cluster-specific composition of stacks     | References to stacks              |
| `stacks/`      | Ready to use composition of modules for specific use-cases  | manifests     |
| `registry/`    | Deployment targets for ArgoCD (actual environments)   | GitOps entrypoints (app list)     |
| `bootstrap/`   | Local single-use cluster to kick off KubeZero         | Minimal manifests + k3d YAML      |
| `controller/`  | ArgoCD `Application` definitions used by bootstrap    | App CRs pointing to registry/stacks |

---
