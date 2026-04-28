# CMS Setup & Deployment Guide

## Overview

The Secco CMS (`secco_cms`) is a separate application for managing website content. It provides a markdown editor with live preview, image upload to Cloudflare Images, and content publishing workflow.

**Location:** `/home/pm/code/secco_cms`
**Tech:** SvelteKit (frontend) + Go (backend)
**URL:** `cms.seccostudio.com` (or any subdomain you choose)

## Prerequisites

1. secco_forms backend running with content API enabled
2. Cloudflare Images configured (see `cloudflare-images.md`)
3. Node.js 20+ and Go 1.22+

## Step 1: Generate Secrets

### Content API Key
Used for CMS → website API communication:
```bash
openssl rand -hex 32
```
Set as `CONTENT_API_KEY` in secco_forms and `WEBSITE_API_KEY` in secco_cms.

### Preview Token
Used for viewing draft content on the website:
```bash
openssl rand -hex 32
```
Set as `PREVIEW_TOKEN` in both apps.

### JWT Secret
Used for CMS session management:
```bash
openssl rand -hex 32
```
Set as `JWT_SECRET` in secco_cms.

### Admin Password Hash
Generate a bcrypt hash for the admin password:
```bash
# Using htpasswd (Apache utils)
htpasswd -nbBC 12 "" "your-secure-password" | cut -d: -f2

# Or using Go:
# go run -mod=mod golang.org/x/crypto/bcrypt <<< "your-password"
```
Set as `CMS_ADMIN_PASS_HASH` in secco_cms.

## Step 2: Configure Environment Variables

### secco_forms (.env additions)
```env
# Content API
DATABASE_PATH=./data/secco.db
CONTENT_API_KEY=<generated-api-key>
PREVIEW_TOKEN=<generated-preview-token>

# Cloudflare Images
CLOUDFLARE_ACCOUNT_ID=<your-account-id>
CLOUDFLARE_IMAGES_API_TOKEN=<your-cf-api-token>
CLOUDFLARE_ACCOUNT_HASH=<your-account-hash>
```

### secco_cms (.env)
```env
# Admin credentials
CMS_ADMIN_USER=admin
CMS_ADMIN_PASS_HASH=<bcrypt-hash>
JWT_SECRET=<generated-jwt-secret>

# Website API connection
WEBSITE_API_URL=http://localhost:8080
WEBSITE_API_KEY=<same-as-CONTENT_API_KEY>
WEBSITE_PUBLIC_URL=https://seccostudio.com
PREVIEW_TOKEN=<same-as-PREVIEW_TOKEN>

# Cloudflare Images
CLOUDFLARE_ACCOUNT_ID=<your-account-id>
CLOUDFLARE_IMAGES_API_TOKEN=<your-cf-api-token>
CLOUDFLARE_ACCOUNT_HASH=<your-account-hash>

# Server
PORT=8081
```

## Step 3: Local Development

### Terminal 1 — secco_forms (website + API)
```bash
cd /home/pm/code/secco_forms
npm run dev          # SvelteKit on :5173
```

### Terminal 2 — secco_forms backend
```bash
cd /home/pm/code/secco_forms/backend
go run .             # Go API on :8080
```

### Terminal 3 — secco_cms
```bash
cd /home/pm/code/secco_cms
npm run dev          # CMS frontend on :5174
```

### Terminal 4 — secco_cms backend
```bash
cd /home/pm/code/secco_cms/backend
go run .             # CMS API on :8081
```

Or use docker-compose from secco_forms root:
```bash
docker compose up --build
```

## Step 4: Docker Deployment

### Build Images
```bash
# Website
cd /home/pm/code/secco_forms
docker build -t secco-studio:latest .

# CMS
cd /home/pm/code/secco_cms
docker build -t secco-cms:latest .
```

### Run with Docker Compose
```yaml
# docker-compose.yml (in secco_forms)
services:
  website:
    image: secco-studio:latest
    ports:
      - "8080:8080"
    volumes:
      - secco-data:/app/data
    env_file:
      - .env

  cms:
    image: secco-cms:latest
    ports:
      - "8081:8081"
    env_file:
      - ../secco_cms/.env
    depends_on:
      - website

volumes:
  secco-data:
```

## Step 5: Kubernetes Deployment

### Website (secco_forms)
Add PersistentVolumeClaim for SQLite:
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: secco-data
spec:
  accessModes: [ReadWriteOnce]
  resources:
    requests:
      storage: 1Gi
```

Mount in Deployment:
```yaml
volumes:
  - name: data
    persistentVolumeClaim:
      claimName: secco-data
containers:
  - name: secco-studio
    volumeMounts:
      - name: data
        mountPath: /app/data
    env:
      - name: DATABASE_PATH
        value: /app/data/secco.db
```

### CMS (secco_cms)
New Deployment + Service + Ingress:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: secco-cms
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt
spec:
  tls:
    - hosts: [cms.seccostudio.com]
      secretName: secco-cms-tls
  rules:
    - host: cms.seccostudio.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: secco-cms
                port:
                  number: 8081
```

### DNS
Add A/CNAME record for `cms.seccostudio.com` pointing to your cluster ingress.

## Step 6: Kubernetes Secrets

Use SealedSecrets (per project convention):
```bash
# Create secret
kubectl create secret generic secco-cms-secrets \
  --from-literal=CMS_ADMIN_PASS_HASH='<hash>' \
  --from-literal=JWT_SECRET='<secret>' \
  --from-literal=WEBSITE_API_KEY='<key>' \
  --from-literal=PREVIEW_TOKEN='<token>' \
  --from-literal=CLOUDFLARE_IMAGES_API_TOKEN='<token>' \
  --dry-run=client -o yaml | kubeseal --format yaml > k8s/sealed-secret-cms.yaml
```

## Security Considerations

1. **CMS should not be publicly accessible without auth** — the Go backend enforces JWT authentication
2. **API keys are separate from admin credentials** — compromised CMS login doesn't expose API key
3. **Preview tokens are time-independent** — rotate them periodically
4. **Cloudflare API token is scoped** — only Images:Edit permission, not full account access
5. **SQLite is single-writer** — only one instance of secco_forms should write to the DB (single pod in K8s)
6. **HTTPS required** — both website and CMS should be behind TLS (cert-manager or Cloudflare)

## Backup Strategy

SQLite database backup:
```bash
# Simple file copy (ensure no writes during copy)
cp /app/data/secco.db /backups/secco-$(date +%Y%m%d).db

# Or use SQLite backup command (safe during writes)
sqlite3 /app/data/secco.db ".backup /backups/secco-$(date +%Y%m%d).db"
```

Consider a CronJob in K8s for automated daily backups.

## CMS Features Quick Reference

| Feature | Description |
|---------|-------------|
| Dashboard | Overview of all content with quick actions |
| Article Editor | Markdown editor with live preview, image upload |
| Project Editor | Same as articles, plus sort order field |
| Process Editor | Single-entry editor for the design process page |
| Image Manager | Upload, browse, delete images on Cloudflare |
| Preview | View draft content as it appears on the website |
| Publish/Draft | Toggle content visibility on the website |
