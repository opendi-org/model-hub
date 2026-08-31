# Deployment Guide

This guide describes a production deployment of OpenDI Model Hub on a single Linux VM using Docker Compose.

Target audience: developers / IT operations staff deploying for sponsor use.

## 1) Architecture and Deployment Mode

Production deployment uses `compose.prod.yaml` with four services:

- `nginx`: public entrypoint on ports `80` and `443`
- `web`: frontend static app (served behind nginx)
- `api`: Go backend API (internal container network)
- `db`: PostgreSQL database (internal container network + persisted Docker volume)

Public traffic flow:

- Browser -> `nginx` (HTTPS)
- `nginx` -> `web` for `/`
- `nginx` -> `api` for `/api/`

No test/dummy data is seeded by this deployment flow.

## 2) Prerequisites

On the target VM, install and verify:

- Docker Engine + Docker Compose plugin (`docker compose version`)
- Git (`git --version`)
- TLS certificate and key files for your production domain.
   - For prod, needs to be CA-issued.
- TLS certificate and key files for your production DB. Common Name should be `db`.
   - Use `db/generate-cert.sh` for db certs. It will set CN.
   - Self-signed cert is fine for prod.
- DNS A/AAAA record pointing your domain to this VM

Use your distro/cloud standard install method for Docker.

## 3) Build + Install Steps

### 3.1 Clone and enter repository

```bash
git clone --recurse-submodules <your-repo-url> opendi
cd opendi
```

If the repo was already cloned without submodules:

```bash
git submodule update --init --recursive
```

### 3.2 Create production environment file

Copy and edit:

```bash
cp .env.example .env
```

Set all required values in `.env`:

- `DB_NAME`
- `DB_USERNAME`
- `DB_PASSWORD`
- `JWT_SECRET`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URL`
- `API_URL`
- `SSL_CRT_PATH`
- `SSL_KEY_PATH`
- `DB_SSL_CRT_PATH`
- `DB_SSL_KEY_PATH`

Production-specific guidance:

- `API_URL`: set to your public HTTPS API path, for example `https://your-domain.example/api`
- `GOOGLE_REDIRECT_URL`: set to your public callback URL, for example `https://your-domain.example/auth/callback`
- `SSL_CRT_PATH` and `SSL_KEY_PATH`: (likely absolute) paths on the VM that exist and are readable by Docker

### 3.3 Start production stack

```bash
docker compose -f compose.prod.yaml up -d --build
```

### 3.4 Confirm services are up

```bash
docker compose -f compose.prod.yaml ps
docker compose -f compose.prod.yaml logs --tail=100 nginx api web db
```

## 4) Security Checklist (Production)

- Serve only HTTPS publicly (already enforced by `nginx.conf.template` HTTP->HTTPS redirect).
- Keep `DB_PASSWORD` and `JWT_SECRET` strong and private.
- Restrict VM inbound firewall to admin SSH + web traffic only.
- Do not publish DB port externally in production.
- Keep host OS and Docker packages patched; rotate TLS certs before expiration.
- Back up DB volume regularly.

## 5) First-Time Configuration / First Users

Authentication is Google OAuth-based:

1. Configure Google OAuth app credentials (`GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`).
2. Ensure authorized redirect URI exactly matches `GOOGLE_REDIRECT_URL`.
3. First user signs in through the web UI at your domain.

No manual bootstrap admin script is required by this deployment path.

### 5.1 Google OAuth setup (required)

In Google Cloud Console:

1. Create/select a project.
2. Open **APIs & Services -> OAuth consent screen**.
3. Configure app name/support email and add required test/production users.
4. Open **APIs & Services -> Credentials -> Create Credentials -> OAuth client ID**.
5. Choose **Web application** as client type.
6. Add your production callback URL to **Authorized redirect URIs**:
   - Example: `https://your-domain.example/auth/callback`
7. (Optional but recommended) Add your domain to **Authorized JavaScript origins**:
   - Example: `https://your-domain.example`
8. Save and copy the generated **Client ID** and **Client Secret**.

Set these in `.env`:

- `GOOGLE_CLIENT_ID=<client-id>`
- `GOOGLE_CLIENT_SECRET=<client-secret>`
- `GOOGLE_REDIRECT_URL=https://your-domain.example/auth/callback`

Important:

- `GOOGLE_REDIRECT_URL` must exactly match an Authorized redirect URI (including scheme, path, and trailing slash behavior).
- Use HTTPS URLs in production.

## 6) Verification Plan

Run these checks after installation and after every upgrade.

### 6.1 Infrastructure checks

```bash
docker compose -f compose.prod.yaml ps
docker compose -f compose.prod.yaml logs --tail=50 nginx api db
```

### 6.2 HTTPS and routing checks

From any machine that can reach the VM/domain:

```bash
curl -I http://your-domain.example
curl -I https://your-domain.example
curl -I https://your-domain.example/api/v0/search?q=test
```

Expected: HTTP redirects to HTTPS, site loads on HTTPS, and `/api` responds.

### 6.3 Functional checks (UI)

1. Open `https://your-domain.example`
2. Confirm UI loads fully
3. Test sign-in via Google
4. Create a new repository
5. Upload a model as a new tag in that repository
6. Confirm the uploaded tag is visible and retrievable

### 6.4 Data persistence check

1. Create or update data through normal app workflow
2. Restart stack:

```bash
docker compose -f compose.prod.yaml restart
```

3. Verify data is still present (confirms DB volume persistence)

## 7) Upgrades / Re-deploy

For future releases:

```bash
git pull
docker compose -f compose.prod.yaml up -d --build
docker image prune -f
```

## 8) Troubleshooting Guide

### Problem: containers do not start or keep restarting

```bash
docker compose -f compose.prod.yaml ps
docker compose -f compose.prod.yaml logs --tail=200 api db web nginx
```

Usually caused by invalid `.env` values or DB healthcheck failures.

### Problem: HTTPS does not come up

```bash
ls -l <path-to-cert> <path-to-key>
docker compose -f compose.prod.yaml logs --tail=100 nginx
```

Most common cause is bad `SSL_CRT_PATH` / `SSL_KEY_PATH` or cert/key mismatch.


### Problem: Server refused TLS connection, no encryption, or db not starting

Similar to previous, check for bad `DB_SSL_CRT_PATH` / `DB_SSL_KEY_PATH` or cert/key mismatch for the db.  
Consider re-generating DB certs via `generate-cert.sh`.

### Problem: Google login fails / redirect mismatch

Ensure `GOOGLE_REDIRECT_URL` exactly matches the redirect URI configured in Google OAuth console, then redeploy.

### Problem: frontend loads but API calls fail

```bash
docker compose -f compose.prod.yaml logs --tail=100 nginx api
curl -I https://your-domain.example/api/v0/search?q=test
```

Usually `API_URL` is wrong or API is unhealthy.

### Problem: DB connection errors in API logs

Check `DB_NAME`, `DB_USERNAME`, `DB_PASSWORD`, then restart:

```bash
docker compose -f compose.prod.yaml down
docker compose -f compose.prod.yaml up -d --build
```

## 9) CLI Deployment to PyPI

Use this when releasing the `opendi` CLI package for end users.

### 9.1 Prerequisites

- PyPI account with permission to publish `opendi`
- PyPI API token
- `uv` installed on release machine

### 9.2 Build and publish

From `cli/`:

```bash
cd cli
uv sync --extra dev
uv build
```

Set your PyPI token (recommended via environment variable):

```bash
export UV_PUBLISH_TOKEN="pypi-<your-token>"
```

Publish:

```bash
uv publish
```

If publishing to TestPyPI first:

```bash
uv publish --index testpypi
```

### 9.3 Post-publish verification (pipx)

Install from PyPI with pipx and verify CLI starts.

Windows:

```bash
py -m pip install --user pipx
py -m pipx install opendi
opendi --help
```

macOS / Linux:

```bash
python3 -m pip install --user pipx
python3 -m pipx install opendi
opendi --help
```
