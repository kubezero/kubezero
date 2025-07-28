# Vault SSO Stack

**Description:**
Secure identity-backed secret access using External Secrets Operator with Vault as the backend and SSO integration.

- **Cloud Agnostic:** This stack is designed to work on any Kubernetes cluster, regardless of the underlying cloud provider.
- **Uses Existing Infrastructure:** Leverages the existing External Secrets Operator module.
- **Vault Integration:** Connects to an existing Vault instance with SSO authentication.
- **Kubernetes Native:** Uses ExternalSecret resources for seamless integration.

## Architecture

This stack uses the External Secrets Operator to connect to Vault:
1. **Vault** (external) - Handles SSO authentication and secret storage
2. **External Secrets Operator** - Manages Kubernetes secret synchronization
3. **SecretStore** - Defines the connection to Vault
4. **ExternalSecret** - Defines which secrets to fetch and where to store them

## Usage

### Prerequisites
1. **Vault instance** with SSO configured (OIDC/SAML)
2. **External Secrets Operator** (already included in this stack)

### Setup Steps
1. **Configure Vault SSO** (OIDC/SAML) in your Vault instance
2. **Update `vault-secretstore.yaml`** with your Vault server URL and authentication details
3. **Create ExternalSecret resources** to define which secrets to fetch from Vault
4. **Apply the stack** using your preferred GitOps or Kustomize workflow

### Vault Configuration

In your Vault instance, enable the auth method and create a role for External Secrets:

```bash
# Enable Kubernetes auth (for service account authentication)
vault auth enable kubernetes

# Create a role for External Secrets
vault write auth/kubernetes/role/external-secrets \
  bound_service_account_names=external-secrets \
  bound_service_account_namespaces=external-secrets \
  policies=external-secrets-policy \
  ttl=1h

# Create a policy for External Secrets
vault policy write external-secrets-policy -<<EOF
path "secret/data/*" {
  capabilities = ["read"]
}
EOF
```

### Example ExternalSecret

The stack includes an example ExternalSecret that demonstrates how to fetch secrets:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: example-vault-secret
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-sso
    kind: SecretStore
  target:
    name: example-secret
  data:
    - secretKey: database-password
      remoteRef:
        key: my-app/database
        property: password
```

## Benefits

- ✅ **No Vault deployment overhead** - Uses existing Vault infrastructure
- ✅ **Kubernetes native** - Uses ExternalSecret CRDs
- ✅ **SSO integration** - Vault handles authentication, ESO handles sync
- ✅ **Standard pattern** - Follows External Secrets Operator best practices
- ✅ **Separation of concerns** - Clear boundaries between components

## References
- [External Secrets Operator](https://external-secrets.io/)
- [Vault Provider for ESO](https://external-secrets.io/v0.8.3/provider/vault/)
- [Vault OIDC Auth Method](https://developer.hashicorp.com/vault/docs/auth/jwt) 