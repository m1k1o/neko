# Deploying Neko to Azure Container Apps

This guide walks you through deploying [Neko](https://github.com/m1k1o/neko) (your fork at `networkneil/neko`) to Azure Container Apps with GitHub Actions CI/CD and federated credentials.

## Architecture

```
GitHub Push → GitHub Actions → Build Image → Push to ACR → Deploy to Container Apps
                                                              ↓
                                                    HTTPS ingress (:443)
                                                    WebRTC media (UDP ports)
```

**Key decisions:**
- **Azure Container Apps** — serverless containers, auto-scaling, pay-per-use
- **Azure Container Registry (ACR)** — private Docker image registry
- **Terraform** — infrastructure as code
- **Azure Workload Identity Federation** — GitHub Actions authenticates to Azure without storing secrets

## Prerequisites

1. [Azure subscription](https://azure.microsoft.com/free/)
2. [GitHub account](https://github.com) (your fork is at `github.com/networkneil/neko`)
3. [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.5
4. [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli) (`az`)
5. [Docker](https://docs.docker.com/get-docker/) (for local testing)

## Step 1: Authenticate to Azure

```bash
az login
# Or for browser-based auth:
az login --use-device-code
```

Set your subscription:

```bash
az account set --subscription "<YOUR_SUBSCRIPTION_ID>"
```

## Step 2: Deploy Infrastructure with Terraform

Navigate to the infra directory:

```bash
cd infra
```

Initialize and apply:

```bash
terraform init
terraform plan -out=tfplan
terraform apply tfplan
```

This creates:
- **Resource Group** — all resources live here
- **Azure Container Registry (ACR)** — stores your Docker images
- **Container Apps Environment** — networking & logging
- **Log Analytics Workspace** — container logs
- **Container App** — runs the Neko container

Outputs will be printed after apply. Save them for later:

```bash
# Copy these values to GitHub vars (Step 4)
terraform output -json | jq -r 'to_entries[] | "\(.key)=\(.value)"'
```

## Step 3: Create Azure Service Principal + Federated Credential

This enables GitHub Actions to authenticate to Azure **without storing secrets**.

### 3a. Create a Service Principal

```bash
SP_NAME="neko-github-deploy"
RESOURCE_GROUP="<your-resource-group-name>"

az ad sp create-for-rbac \
  --name "$SP_NAME" \
  --role Contributor \
  --scopes "/subscriptions/<YOUR_SUBSCRIPTION_ID>/resourceGroups/$RESOURCE_GROUP" \
  --json > azure-sp.json

cat azure-sp.json
```

This outputs `appId`, `password`, and `tenant`. **Save the `appId` (client ID) and `tenant` — you'll need them next.**

### 3b. Create Federated Credential

First, get your GitHub OIDC issuer URL:

```bash
# This is always the same for GitHub:
ISSUER_URL="https://token.actions.githubusercontent.com"
WEBHOOK_SECRET=$(gh api user/repos/networkneil/neko/hooks -X POST \
  --field name 'config' \
  --field config '{"content":""}' \
  --jq '.webhook_secret' 2>/dev/null || echo "")

# Actually, use the simpler approach:
```

The federated credential links your GitHub repo/workflow to the service principal:

```bash
CLIENT_ID=$(az ad sp list --display-name "$SP_NAME" --query '[0].appId' -o tsv)

az ad app federated-credential create \
  --id "$CLIENT_ID" \
  --params '{
    "name": "github-actions",
    "issuer": "https://token.actions.githubusercontent.com",
    "subject": "repo:networkneil/neko:ref:refs/heads/main",
    "description": "GitHub Actions workflow for deploying Neko to Azure",
    "audiences": ["api://AzureADTokenExchange"]
  }'
```

### 3c. Grant ACR Pull Role

The GitHub workflow needs permission to pull from ACR:

```bash
ACR_ID=$(az acr show --name <your-acr-name> --query id -o tsv)
SP_OBJECT_ID=$(az ad sp list --display-name "$SP_NAME" --query '[0].id' -o tsv)

# Actually, we need the SP's object ID in Azure AD:
SP_OBJECT_ID=$(az ad sp show --id "$CLIENT_ID" --query id -o tsv)

az role assignment create \
  --role "AcrPull" \
  --scope "$ACR_ID" \
  --assignee-object-id "$SP_OBJECT_ID" \
  --assignee-principal-type ServicePrincipal
```

## Step 4: Configure GitHub Repository Settings

### 4a. Enable OIDC in Azure AD (if not already)

Your Azure AD tenant must allow GitHub as an OIDC provider:

1. Go to [Azure Portal > Enterprise Apps](https://portal.azure.com/#blade/Microsoft_AAD_IAM/EnterpriseAppsMenuBlade/overview)
2. Find your service principal (`neko-github-deploy`)
3. Verify the federated credential is listed under **Certificates & secrets**

### 4b. Set GitHub Repository Variables

```bash
# These are NOT secrets — they're public repo config
gh variable set RESOURCE_GROUP -b "<your-resource-group>"
gh variable set ACR_NAME -b "<your-acr-name>"
gh variable set CONTAINER_APP_NAME -b "<your-container-app-name>"
gh variable set ACR_LOGIN_SERVER -b "<your-acr>.azurecr.io"
```

### 4c. Set GitHub Repository Secrets

```bash
# These ARE secrets — only visible to workflows
gh secret set AZURE_CLIENT_ID -b "<service-principal-app-id>"
gh secret set AZURE_TENANT_ID -b "<azure-tenant-id>"
gh secret set AZURE_SUBSCRIPTION_ID -b "<your-subscription-id>"

# ACR credentials (generated by Terraform or Azure CLI)
ACR_PASSWORD=$(az acr credential show --name <your-acr-name> --query passwords[0].value -o tsv)
gh secret set ACR_USERNAME -b "<your-acr-name>"
gh secret set ACR_PASSWORD -b "$ACR_PASSWORD"
```

## Step 5: Trigger the First Deploy

Push to main to trigger the workflow:

```bash
git add infra/ .github/workflows/deploy-azure.yml AZURE_DEPLOYMENT.md
git commit -m "feat: add Azure Container Apps deployment infrastructure"
git push origin feat/azure-container-apps
# Then create a PR and merge to main
```

Or trigger manually:

```bash
gh workflow run deploy-azure.yml
```

## Step 6: Access Neko

After the workflow completes, get your URL:

```bash
URL=$(az containerapp show \
  --name <container-app-name> \
  --resource-group <resource-group> \
  --query 'properties.configuration.ingress.fqdn' \
  -o tsv)

echo "https://$URL"
```

## WebRTC Networking Note

Azure Container Apps exposes HTTP/HTTPS ingress natively. For raw UDP ports (52000-52100), you have two options:

### Option A: Use a TURN Server (Recommended for Container Apps)

Configure Neko to relay all media through HTTPS via a TURN server. Add this to your environment variables or config:

```yaml
NEKO_WEBRTC_ICESERVERS_BACKEND: '[{"urls":["turns:turn.your-domain.com:443"],"username":"neko","credential":"your-secret"}]'
```

You can deploy a TURN server (like [Coturn](https://github.com/coturn/coturn)) on a separate Azure VM or Container Instance.

### Option B: Use Azure VM Instead

If you need direct UDP port exposure for low-latency WebRTC, consider deploying to an **Azure VM** instead of Container Apps. The existing Docker setup works as-is — just run the container with `-p 8080:8080 -p 52000-52100:52000/udp`.

## Managing Resources

### Destroy everything (clean up):

```bash
cd infra
terraform destroy
```

### Scale manually:

```bash
az containerapp scale \
  --name <container-app-name> \
  --resource-group <resource-group> \
  --min-replicas 1 \
  --max-replicas 3
```

### View logs:

```bash
az containerapp logs show \
  --name <container-app-name> \
  --resource-group <resource-group> \
  --follow
```

## Cost Estimate (East US)

| Resource | Monthly Cost |
|----------|-------------|
| Container Apps (0-2 replicas) | ~$0 (pay per GB-sec used) |
| ACR (Standard) | ~$15/month |
| Log Analytics (PerGB2018) | ~$5-20/month (depends on log volume) |
| **Total** | **~$20-35/month** |

Set `min_replicas = 0` in `terraform.tfvars` to scale to zero when idle.
