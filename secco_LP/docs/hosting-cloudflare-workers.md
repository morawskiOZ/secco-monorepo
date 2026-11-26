# Hosting on Cloudflare Workers — plan and runbook

Moves www.seccostudio.com off the paid Hetzner/Talos Kubernetes cluster to
Cloudflare Workers (Free plan). Status and open steps are at the bottom.

## Goal

Host the landing page and both forms for €0/month with the same behaviour as
today, and switch over without downtime or lost form submissions.

## What runs today

| Piece | Today (Hetzner cluster) |
|---|---|
| Static site | SvelteKit static build embedded in a Go binary, 2 pods |
| Forms | `POST /api/submit`, `POST /api/submit-preferences`: Turnstile check, 5/hour per IP, email via Gmail SMTP with the uploaded file attached |
| Health | `GET /api/health` |
| Unknown paths | SPA shell (`200.html`) with status 200 |
| Redirects | apex → www and http → https, both 308 (Traefik) |
| TLS / DNS | Let's Encrypt via cert-manager; DNS on Cloudflare, proxied |
| Deploy | GitHub Actions → Docker image on GHCR → ArgoCD |

## Options considered

| Option | Cost | Verdict |
|---|---|---|
| **Cloudflare Workers + Static Assets** | Free: 100k Worker requests/day, static files unlimited | **Chosen.** DNS is already there, no cold starts, no servers. Backend had to be ported (≈400 lines). |
| Google Cloud Run | Free tier | Go image would run unchanged, but needs a billing account, has cold starts and a second vendor. Fallback if Workers disappoints. |
| Oracle Always Free VM | Free | Still a server to patch; capacity/reclaim risk. |
| Render free | Free | Blocks outbound SMTP ports since Sept 2025 — not viable. |
| Vercel Hobby | Free | Not for commercial sites. |

Limits checked on 2026-11-25: Workers Free allows 10 ms CPU per request,
128 MB memory, 100 MB request bodies; outbound TCP is allowed on every port
except 25, so Gmail SMTP on 587 works.

## Design

```
browser ── Cloudflare ── Worker "secco-studio" (worker/src/index.ts)
                            ├─ apex / http  → 308 https://www.seccostudio.com
                            ├─ /api/health
                            ├─ /api/submit, /api/submit-preferences
                            │     ├─ Durable Object per IP: 5 per hour
                            │     ├─ Turnstile siteverify
                            │     └─ Gmail SMTP :587 STARTTLS (cloudflare:sockets)
                            └─ everything else → static assets (./build),
                               unknown paths → 200.html
```

Decisions:

- **Email stays on Gmail SMTP.** Same sender, recipient, Reply-To, subjects and
  body as before, no DNS changes. Cloudflare's own Email Service was the
  alternative; it is still beta, changes the sender address and needs MX/SPF
  records.
- **Rate limit in a Durable Object** (SQLite-backed, available on Free). The
  built-in rate-limit binding only supports 10 s / 60 s windows.
- **Upload cap 18 MB per request** (was 100 MB on paper). The form only allows
  one 10 MB file and Gmail rejects mail over 25 MB, so nothing that worked
  before is refused.
- **Go backend and Dockerfile stay in the repo** until the cluster is switched
  off, so rollback is one DNS change.

### Differences from today

- `/polityka-prywatnosci` is served as prerendered HTML (Go served the SPA
  shell and rendered it client-side). Same page, faster.
- Hashed files under `/_app/immutable/` get a one-year cache header.
- Urlencoded (non-multipart) posts get 400; a stricter multipart check than Go.
- Attachment base64 is wrapped at 76 characters (Go sent one long line).
- The test host `secco-studio.apps.stockkup.com` goes away; staging lives on
  `secco-studio-staging.<account>.workers.dev`.

### Known risk: CPU limit on large uploads

Free allows a nominal 10 ms CPU per request (time waiting on Gmail does not
count). Measured on staging on 2026-11-25: a 3 MB attachment used 24 ms and a
10 MB attachment 64 ms, and both were delivered — Cloudflare tolerates
occasional overruns, but does not guarantee it. If submissions with large
files ever fail with error 1102 (visible in the Worker logs), the fix is the
Workers Paid plan ($5/month, 30 s CPU) — still far below the cluster cost.

## Testing forms without mailing the studio (staging)

`https://secco-studio-staging.secco-studio.workers.dev` is the same code and
site as production with three differences:

- **Recipient** is the `RECIPIENT_EMAIL` *secret* of the staging Worker — the
  tester's own address — not the studio inbox. If the secret is missing,
  submissions fail with 500 rather than falling back to the real address.
  Change it with `npx wrangler secret put RECIPIENT_EMAIL --env staging`.
- **Subject** starts with `[TEST] `.
- **Turnstile** uses Cloudflare's always-pass test keys (the real widget is
  bound to the production hostnames), and the per-IP limit is 50/hour.

Mail is still sent through the studio's Gmail account, so test messages show
up in that account's Sent folder. Production is a separate Worker with its own
secrets and counters; nothing done on staging touches it.

Deploy staging from the Actions workflow (environment: staging) or locally:

```bash
PUBLIC_TURNSTILE_SITE_KEY=1x00000000000000000000AA npm run build && npm run deploy:staging
```

## Runbook

Prerequisite (owner, once): a Cloudflare API token with **Workers Scripts:
Edit** on the account and **Workers Routes: Edit** + **DNS: Edit** on
`seccostudio.com` (the "Edit Cloudflare Workers" template covers it), plus the
account ID. The existing infra token is zone-only and cannot deploy Workers.

```bash
cd secco_LP
# wrangler only reads CLOUDFLARE_API_TOKEN: give it the Workers token
export CLOUDFLARE_API_TOKEN="$CLOUDFLARE_WORKERS_API_TOKEN"   CLOUDFLARE_ACCOUNT_ID=...
export KUBECONFIG=<path to infra/kubeconfig>
```

1. **CI secrets**
   ```bash
   gh secret set CLOUDFLARE_WORKERS_API_TOKEN -R morawskiOZ/secco-monorepo
   gh secret set CLOUDFLARE_ACCOUNT_ID -R morawskiOZ/secco-monorepo
   ```
2. **Deploy staging** (no domain attached): `npm ci && npm run build && npm run deploy:staging`
   (build with the real `PUBLIC_TURNSTILE_SITE_KEY`; add the workers.dev
   hostname to the Turnstile widget's allowed hostnames, or test with the
   Turnstile test keys).
3. **Copy secrets from the cluster** to both Workers — values never touch the terminal:
   ```bash
   for env in staging ""; do for key in SMTP_PASS TURNSTILE_SECRET_KEY; do
     kubectl -n apps get secret secco-studio-secret -o jsonpath="{.data.$key}" | base64 -d \
       | npx wrangler secret put "$key" --env="$env"
   done; done
   ```
   (The production Worker must exist first: `npx wrangler deploy --env="" --dry-run`
   does not create it; `wrangler secret put` offers to create it.)
4. **Verify staging**: all routes return what production returns; submit each
   form once — one without a file, one with a ~10 MB file — and confirm both
   emails arrive with a working Reply-To and attachment. Watch `npx wrangler tail --env staging`.
5. **Cut over**: delete the two proxied `A` records (`seccostudio.com`,
   `www.seccostudio.com`, pointing at the Hetzner load balancer), then `npm run deploy`. Wrangler
   creates the custom domains and certificates. Records are proxied, so there
   is no DNS propagation wait; expect well under a minute of errors. Do it at a
   quiet hour.
6. **Verify production**: the checks from step 4 against www.seccostudio.com,
   plus `curl -I https://seccostudio.com/x?y=1` → 308 to www.
7. **Burn-in 48 h.** The cluster keeps running untouched.
8. **Decommission**: remove `secco-studio/` from `K8S-Hetzner-Apps` (ArgoCD
   prunes it), then delete `backend/`, `Dockerfile`, `docker-compose.yml`,
   `Taskfile.yml` and `.github/workflows/build-lp.yml` here.

**Rollback** (any time before step 8): in the Cloudflare dashboard remove both
custom domains from the Worker and re-create the two proxied `A` records to
the Hetzner load balancer IP (in the infra repo, `NODE_PUBLIC_IPS`/LB). The cluster deployment is still serving.

## Day-to-day

- Deploy: GitHub → Actions → "Deploy LP to Cloudflare Workers" → staging or production.
- Local: `npm run build && npm run dev:worker` (secrets in `.dev.vars`, see `.env.example`).
- Tests: `npm test`, `npm run check:worker`.
- Logs: `npx wrangler tail`, or Workers → secco-studio → Logs in the dashboard.

## Status

- [x] Worker written and unit-tested (23 tests)
- [x] Verified locally with `wrangler dev`: routes, SPA fallback, redirects,
      validation, rate limit, Turnstile, and a real STARTTLS + AUTH handshake
      with smtp.gmail.com; landing page HTML identical to production
- [x] Deploy workflow
- [x] Cloudflare API token with Workers permissions
- [x] Staging deployed and verified 2026-11-25: routes, validation, rate limit, emails with no file / 3 MB / 10 MB
- [ ] CI secrets (step 1), cutover (steps 5–6), burn-in, decommission (steps 7–8)
