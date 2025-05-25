# Reference Architecture

## Bootstrap

```mermaid
flowchart LR
  %% Cluster 1 - Bootstrap.
  style k8sBoostrap stroke-dasharray: 2 2
  subgraph k8sBoostrap["Kubernetes Cluster - Local Bootstrap"]
    %% Initial bootstrap.
    K3D["K3D\nHelm Controller"] --> k8sBoostrap.GitOps

    %% Management components.
    subgraph k8sBoostrap.GitOps["Bootstrap"]
      direction TB
      k8sBoostrap.ArgoCD[Argo CD]
      k8sBoostrap.Crossplane[Crossplane]
    end
  end

  %% Cluster 2 - Management.
  style k8sManagement stroke-dasharray: 2 2
  subgraph k8sManagement["Kubernetes Cluster - Management"]

    %% Management components.
    subgraph k8sManagement.GitOps["GitOps"]
      k8sManagement.GitOps.ArgoCD[Argo CD]
    end

    subgraph k8sManagement.InfraManagement[Infrastructure Management]
      k8sManagement.InfraManagement.Crossplane[Crossplane]
    end

    %% Common utils.
    subgraph k8sManagement.Utils["Utils"]
      Utils.Ingress[Ingress Nginx]
      Utils.CertManager[Cert-Manager]
      Utils.ExternalDNS[External-DNS]
      Utils.ESO[External Secret Operator]
      Utils.Vault[Vault]
      Utils.Monitoring[Monitoring]
    end
  end

  %% Bootstrap connections.
  k8sBoostrap.Crossplane --"Build Management Cluster"--> k8sManagement
  k8sBoostrap.ArgoCD --"Deploy Management Argo CD"--> k8sManagement.GitOps.ArgoCD
  k8sBoostrap.ArgoCD --"Deploy Management Crossplane"--> k8sManagement.InfraManagement.Crossplane
  k8sBoostrap.ArgoCD --"Deploy Management Utils"--> k8sManagement.Utils
```

## 01. Single Cluster - Main cluster with all environments in virtual clusters

```mermaid
flowchart TD
  %% Cluster 1.
  style k8sMain stroke-dasharray: 2 2
  subgraph k8sMain["Kubernetes Cluster - Main"]
    %% Management components.
    subgraph k8sMain.GitOps["GitOps"]
      direction TB
      k8sMain.GitOps.ArgoCD[Argo CD]
    end

    subgraph k8sMain.InfraManagement[Infrastructure Management]
      direction TB
      k8sMain.InfraManagement.Crossplane[Crossplane]
    end

    subgraph k8sMain.Utils["Utils"]
      direction TB
      k8sMain.Utils.Components[Main Utils Components]
    end

    %% Managed Clusters.
    style k8sMain.Production stroke-dasharray: 2 2
    subgraph k8sMain.Production[Production vCluster]
      k8sMain.Production.ESO[External Secret Operator]
      k8sMain.Production.App["Application(s)"]
    end

    style k8sMain.Staging stroke-dasharray: 2 2
    subgraph k8sMain.Staging[Staging vCluster]
      k8sMain.Staging.ESO[External Secret Operator]
      k8sMain.Staging.App["Application(s)"]
    end

    style k8sMain.Development stroke-dasharray: 2 2
    subgraph k8sMain.Development[Development vCluster]
      k8sMain.Development.ESO[External Secret Operator]
      k8sMain.Development.APP["Application(s)"]
    end
  end

  %% Management connections.
  k8sMain.GitOps --> k8sMain.InfraManagement
  k8sMain.GitOps --> k8sMain.Utils
  k8sMain.InfraManagement --> k8sMain.Production
  k8sMain.InfraManagement --> k8sMain.Staging
  k8sMain.InfraManagement --> k8sMain.Development
```

## 02. Single Cluster - Main cluster with production objects and non-production virtual clusters

```mermaid
flowchart TD
  %% Cluster 1.
  style k8sMain stroke-dasharray: 2 2
  subgraph k8sMain["Kubernetes Cluster - Main"]
    %% Management components.
    subgraph k8sMain.GitOps["GitOps"]
      direction TB
      k8sMain.GitOps.ArgoCD[Argo CD]
    end

    subgraph k8sMain.InfraManagement[Infrastructure Management]
      direction TB
      k8sMain.InfraManagement.Crossplane[Crossplane]
    end

    subgraph k8sMain.Utils["Utils"]
      direction TB
      k8sMain.Utils.Components[Main Utils Components]
    end

    %% Managed Clusters.
    subgraph k8sMain.Production[Production Objects]
      k8sMain.Production.App["Application(s)"]
    end

    style k8sMain.Staging stroke-dasharray: 2 2
    subgraph k8sMain.Staging[Staging vCluster]
      k8sMain.Staging.ESO[External Secret Operator]
      k8sMain.Staging.App["Application(s)"]
    end

    style k8sMain.Development stroke-dasharray: 2 2
    subgraph k8sMain.Development[Development vCluster]
      k8sMain.Development.ESO[External Secret Operator]
      k8sMain.Development.APP["Application(s)"]
    end
  end

  %% Management connections.
  k8sMain.GitOps --> k8sMain.Production
  k8sMain.GitOps --> k8sMain.InfraManagement
  k8sMain.GitOps --> k8sMain.Utils
  k8sMain.InfraManagement --> k8sMain.Staging
  k8sMain.InfraManagement --> k8sMain.Development
```

## 03. Two Clusters - Production cluster and  non-production cluster with virtual clusters

```mermaid
flowchart TD
  %% Cluster 1 - Main.
  style k8sMain stroke-dasharray: 2 2
  subgraph k8sMain["Kubernetes Cluster - Main"]
    %% Management components.
    subgraph k8sMain.GitOps["GitOps"]
      direction TB
      k8sMain.GitOps.ArgoCD[Argo CD]
    end

    subgraph k8sMain.InfraManagement[Infrastructure Management]
      direction TB
      k8sMain.InfraManagement.Crossplane[Crossplane]
    end

    subgraph k8sMain.Utils["Utils"]
      direction TB
      k8sMain.Utils.Components[Main Utils Components]
    end

    %% Managed Clusters.
    subgraph k8sMain.Production[Production Objects]
      k8sMain.Production.App["Application(s)"]
    end

    %% Management connections.
    k8sMain.GitOps --> k8sMain.Production
    k8sMain.GitOps --> k8sMain.InfraManagement
    k8sMain.GitOps --> k8sMain.Utils
  end

  %% Cluster 2 - Non-Production.
  style k8sNonProd stroke-dasharray: 2 2
  subgraph k8sNonProd["Kubernetes Cluster - Non-Production"]
    direction TB
    subgraph k8sNonProd.Utils["Utils"]
      k8sNonProd.Utils.Ingress[Ingress Nginx]
      k8sNonProd.Utils.Monitoring[Monitoring Agent]
    end

    style k8sNonProd.Staging stroke-dasharray: 2 2
    subgraph k8sNonProd.Staging[Staging vCluster]
      k8sNonProd.Staging.ESO[External Secret Operator]
      k8sNonProd.Staging.App["Application(s)"]
    end

    style k8sNonProd.Development stroke-dasharray: 2 2
    subgraph k8sNonProd.Development[Development vCluster]
      k8sNonProd.Development.ESO[External Secret Operator]
      k8sNonProd.Development.APP["Application(s)"]
    end
  end
  k8sMain.InfraManagement --> k8sNonProd
```

## 04. Three Clusters - Management cluster, production cluster and non-production cluster with multi-virtual clusters

```mermaid
flowchart TD
  %% Cluster 1 - Main.
  style k8sMain stroke-dasharray: 2 2
  subgraph k8sMain["Kubernetes Cluster - Main"]
    %% Management components.
    subgraph k8sMain.GitOps["GitOps"]
      direction TB
      k8sMain.GitOps.ArgoCD[Argo CD]
    end

    subgraph k8sMain.InfraManagement[Infrastructure Management]
      direction TB
      k8sMain.InfraManagement.Crossplane[Crossplane]
    end

    subgraph k8sMain.Utils["Utils"]
      direction TB
      k8sMain.Utils.Components[Main Utils Components]
    end

    %% Management connections.
    k8sMain.GitOps --> k8sMain.InfraManagement
    k8sMain.GitOps --> k8sMain.Utils
  end

  %% Cluster 2 - Production
  style k8sProd stroke-dasharray: 2 2
  subgraph k8sProd["Kubernetes Cluster - Production"]
    direction TB
    subgraph k8sProd.Utils["Utils"]
      k8sProd.Utils.Ingress[Ingress Nginx]
      k8sProd.Utils.ESO[External Secret Operator]
      k8sProd.Utils.Monitoring[Monitoring Agent]
    end

    style k8sProd.Production stroke-dasharray: 2 2
    subgraph k8sProd.Production[Production Objects]
      k8sProd.Production.App["Application(s)"]
    end
  end

  %% Cluster 3 - Non-Prodction.
  style k8sNonProd stroke-dasharray: 2 2
  subgraph k8sNonProd["Kubernetes Cluster - Non-Production"]
    direction TB
    subgraph k8sNonProd.Utils["Utils"]
      k8sNonProd.Utils.Ingress[Ingress Nginx]
      k8sNonProd.Utils.Monitoring[Monitoring Agent]
    end

    style k8sNonProd.Staging stroke-dasharray: 2 2
    subgraph k8sNonProd.Staging[Staging vCluster]
      k8sNonProd.Staging.ESO[External Secret Operator]
      k8sNonProd.Staging.App["Application(s)"]
    end

    style k8sNonProd.Development stroke-dasharray: 2 2
    subgraph k8sNonProd.Development[Development vCluster]
      k8sNonProd.Development.ESO[External Secret Operator]
      k8sNonProd.Development.APP["Application(s)"]
    end
  end

  %% Management connections.
  k8sMain.InfraManagement --> k8sProd
  k8sMain.InfraManagement --> k8sNonProd
```

## 05. Four Clusters - Management cluster, production, staging, and development clusters

```mermaid
flowchart TD
  %% Cluster 1 - Main.
  style k8sMain stroke-dasharray: 2 2
  subgraph k8sMain["Kubernetes Cluster - Main"]
    %% Management components.
    subgraph k8sMain.GitOps["GitOps"]
      direction TB
      k8sMain.GitOps.ArgoCD[Argo CD]
    end

    subgraph k8sMain.InfraManagement[Infrastructure Management]
      direction TB
      k8sMain.InfraManagement.Crossplane[Crossplane]
    end

    subgraph k8sMain.Utils["Utils"]
      direction TB
      k8sMain.Utils.Components[Main Utils Components]
    end

    %% Management connections.
    k8sMain.GitOps --> k8sMain.InfraManagement
    k8sMain.GitOps --> k8sMain.Utils
  end

  %% Cluster 2 - Production
  style k8sProd stroke-dasharray: 2 2
  subgraph k8sProd["Kubernetes Cluster - Production"]
    direction TB
    subgraph k8sProd.Utils["Utils"]
      k8sProd.Utils.Ingress[Ingress Nginx]
      k8sProd.Utils.ESO[External Secret Operator]
      k8sProd.Utils.Monitoring[Monitoring Agent]
    end

    style k8sProd.Production stroke-dasharray: 2 2
    subgraph k8sProd.Production[Production Objects]
      k8sProd.Production.App["Application(s)"]
    end
  end

  %% Cluster 3 - Staging
  style k8sStage stroke-dasharray: 2 2
  subgraph k8sStage["Kubernetes Cluster - Staging"]
    direction TB
    subgraph k8sStage.Utils["Utils"]
      k8sStage.Utils.Ingress[Ingress Nginx]
      k8sStage.Utils.ESO[External Secret Operator]
      k8sStage.Utils.Monitoring[Monitoring Agent]
    end

    style k8sStage.Production stroke-dasharray: 2 2
    subgraph k8sStage.Production[Staging Objects]
      k8sStage.Production.App["Application(s)"]
    end
  end

  %% Cluster 4 - Development
  style k8sDev stroke-dasharray: 2 2
  subgraph k8sDev["Kubernetes Cluster - Development"]
    direction TB
    subgraph k8sDev.Utils["Utils"]
      k8sDev.Utils.Ingress[Ingress Nginx]
      k8sDev.Utils.ESO[External Secret Operator]
      k8sDev.Utils.Monitoring[Monitoring Agent]
    end

    style k8sDev.Production stroke-dasharray: 2 2
    subgraph k8sDev.Production[Development Objects]
      k8sDev.Production.App["Application(s)"]
    end
  end

  %% Management connections.
  k8sMain.InfraManagement --> k8sProd
  k8sMain.InfraManagement --> k8sStage
  k8sMain.InfraManagement --> k8sDev
```
