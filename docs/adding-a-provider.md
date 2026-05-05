# Adding a New Cloud Provider

This guide walks through everything needed to add a new cloud provider (e.g. Azure AKS) to KubeZero following the same conventions as the existing providers.

## Overview

Adding a provider means adding four things in order:

```
modules/<cloud>/         ← Crossplane provider + cluster CRDs
stacks/<cloud>-cluster/  ← Combine module + cluster-secret (ArgoCD registration)
packages/<cloud>-*/      ← Ready-to-use management + worker bundles
cli/cmd/scaffold.go      ← Template for kubezero scaffold
```

---

## Step 1 — Add the Crossplane provider module

Create `modules/<cloud>/provider/`:

```
modules/<cloud>/
  provider/
    configuration.yaml   ← Crossplane Configuration package CR
    kustomization.yaml
```

**`configuration.yaml`** installs the Crossplane provider package:

```yaml
---
apiVersion: pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: configuration-<cloud>-<service>
spec:
  package: xpkg.upbound.io/upbound/configuration-<cloud>-<service>:vX.Y.Z
```

Find the latest version at [marketplace.upbound.io](https://marketplace.upbound.io).

---

## Step 2 — Add the cluster module

Create `modules/<cloud>/cluster/` with these files:

```
modules/<cloud>/cluster/
  kustomization.yaml
  provider-config.yaml   ← ProviderConfig CR (credentials reference)
  x<cluster>.yaml        ← Composite resource (XEKS / XGKE / XAKS …)
  xnetwork.yaml          ← VPC / network composite resource
```

### `provider-config.yaml`

```yaml
---
apiVersion: <cloud>.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: provider-<cloud>
  annotations:
    argocd.argoproj.io/sync-wave: "1"
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: <cloud>-creds    # REPLACE: create this secret manually
      key: creds
```

### `x<cluster>.yaml`

Use the high-level composite resource from the Upbound platform configuration:

```yaml
---
apiVersion: <cloud>.platform.upbound.io/v1alpha1
kind: X<Cluster>
metadata:
  name: <cloud>-cluster
spec:
  parameters:
    id: <cloud>-cluster
    region: <default-region>
    version: latest
    nodes:
      count: 1
      instanceType: <default-instance-type>
  writeConnectionSecretToRef:
    name: <cloud>-cluster-kubeconfig
    namespace: crossplane-system
```

### `kustomization.yaml`

```yaml
---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - provider-config.yaml
  - x<cluster>.yaml
  - xnetwork.yaml
```

---

## Step 3 — Add a cluster stack

Create `stacks/<cloud>-cluster/`:

```
stacks/<cloud>-cluster/
  kustomization.yaml
  cluster-secret.yaml
```

### `kustomization.yaml`

```yaml
---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../../modules/<cloud>/cluster
  - ./cluster-secret.yaml
```

### `cluster-secret.yaml`

This is the bridge between Crossplane and ArgoCD. It reads the kubeconfig Crossplane writes and transforms it into an ArgoCD cluster Secret.

```yaml
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: <cloud>-cluster-secret
  namespace: kubezero
  annotations:
    argocd.argoproj.io/sync-wave: "1"
spec:
  refreshInterval: 1h
  secretStoreRef:
    kind: ClusterSecretStore
    name: kubezero-management
  target:
    name: <cloud>-cluster-secret
    template:
      metadata:
        labels:
          argocd.argoproj.io/secret-type: cluster
      data:
        name: "{{ .name }}"
        server: "{{ .server }}"
        config: |
          {
            "tlsClientConfig": {
              "insecure": false,
              "caData": "{{ .caData }}"
            },
            "bearerToken": "{{ .token }}"
          }
  data:
    - secretKey: name
      remoteRef:
        key: <cloud>-cluster-kubeconfig  # matches writeConnectionSecretToRef
        property: endpoint               # adjust property names per provider
    - secretKey: server
      remoteRef:
        key: <cloud>-cluster-kubeconfig
        property: endpoint
    - secretKey: caData
      remoteRef:
        key: <cloud>-cluster-kubeconfig
        property: clusterCA
    - secretKey: token
      remoteRef:
        key: <cloud>-cluster-kubeconfig
        property: token
```

> **Note on auth:** Different clouds use different auth mechanisms in ArgoCD:
> - AWS EKS: use `awsAuthConfig` with `clusterName`
> - GCP GKE: use `execProviderConfig` with `argocd-k8s-auth gcp`
> - Others: bearer token + CA (shown above)

---

## Step 4 — Add packages

Create management and worker packages:

```bash
kubezero scaffold --cloud <cloud> --role management
kubezero scaffold --cloud <cloud> --role worker
```

This generates `packages/<cloud>-management/` and `packages/<cloud>-worker/` with:
- `gitops.yaml` — AppProject + Application CRs
- `infrastructure/kustomization.yaml` — references stack + namePrefix
- `applications/kustomization.yaml` — references k8s-essentials

Then update `cli/cmd/scaffold.go` to add a proper template for the new cloud under `scaffoldTemplates` so future `scaffold` calls generate correct patch files.

---

## Step 5 — Add the credential secret

The provider needs a Kubernetes Secret in `crossplane-system`. This is the one manual step that is intentionally not in the repo (never commit credentials).

Create a script or document it in your runbook:

```bash
# Example: create the cloud credential secret
kubectl create secret generic <cloud>-creds \
  --namespace crossplane-system \
  --from-file=creds=<path-to-credentials-file>
```

---

## Step 6 — Update the CLI

Add the new cloud to the `clouds` map in `cli/cmd/init.go`:

```go
"<cloud>": {
    label: "<Cloud> (<Service>)",
    regions: []string{
        "region-1", "region-2",
    },
},
```

And add it to `cloudKeys` so it appears in `kubezero init`.

---

## Step 7 — Validate

```bash
# Verify kustomize build works for new modules/stacks/packages
kustomize build modules/<cloud>/cluster
kustomize build stacks/<cloud>-cluster
kustomize build packages/<cloud>-management/infrastructure

# Full validation
make validate
```

---

## Conventions checklist

- [ ] `ProviderConfig` has `argocd.argoproj.io/sync-wave: "1"` annotation
- [ ] Cluster composite resource writes to `crossplane-system` via `writeConnectionSecretToRef`
- [ ] `cluster-secret.yaml` uses `ExternalSecret` pointing at `ClusterSecretStore/kubezero-management`
- [ ] ArgoCD cluster Secret has label `argocd.argoproj.io/secret-type: cluster`
- [ ] Package uses `namePrefix: <role>-` to avoid name collisions
- [ ] Package `gitops.yaml` references `registry/<name>/` paths (not `packages/`)
- [ ] Cloud added to `clouds` map and `cloudKeys` slice in `cli/cmd/init.go`
- [ ] Scaffold template added to `scaffoldTemplates` in `cli/cmd/scaffold.go`
