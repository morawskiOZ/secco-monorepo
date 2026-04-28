# Secco Studio — Project Root

Mono-root for all Secco Studio repositories. Three separate repos, each independently deployable.

## Repository Structure

```
secco/
├── secco_LP/        # Public website + valuation form (SvelteKit + Go)
├── secco_cms/       # Admin CMS for content management (SvelteKit + Go + SQLite)
├── secco_renderer/  # Shared npm package: types, markdown renderer, ContentRenderer
├── task.md          # Master implementation todo list
└── PROMPT.md        # Project brief and architectural decisions
```

## How the repos relate

```
secco_renderer  ──(npm)──►  secco_LP      (renders content pages)
secco_renderer  ──(npm)──►  secco_cms     (renders preview pane)

secco_cms  ──(R2 snapshot)──►  GitHub Actions  ──(build)──►  secco_LP Docker image
```

## Key Architectural Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Content at runtime | None — fully static | No DB on website, perfect Lighthouse |
| Content pipeline | CMS → R2 snapshot → GH Actions build | Decoupled, rollback-friendly |
| Image hosting | Cloudflare R2 + Image Transformations | Cheapest, CDN included |
| Preview rendering | Shared npm package (secco_renderer) | Single source of truth, no drift |
| CMS auth | Single admin, JWT + bcrypt, env vars | No user management needed |
| CMS database | SQLite via modernc.org/sqlite | Pure Go, zero infra |
| Package distribution | GitHub git dependency, tagged releases | No npm registry needed |

## Deployment Targets

| App | URL | Host |
|-----|-----|------|
| secco_LP | www.seccostudio.com | MicroK8s via ArgoCD |
| secco_cms | cms.seccostudio.com | MicroK8s via ArgoCD |
| Images/assets | assets.seccostudio.com | Cloudflare R2 (custom domain) |

## Design Tokens

```
Primary background: #EBD5C9 (warm beige)
Accent:            #2E3192 (deep navy)
Secondary accent:  #7F7F80 (medium gray)
Page background:   #fafafa (off-white)
Text:              #454545 (charcoal)
Navbar height:     64px desktop / 56px mobile (--navbar-height CSS var)
Fonts:             "Open Sans", sans-serif
```

## K8s Secrets

Managed in the `appme-k8s` repo using SealedSecrets + `seal.sh` script. Never commit raw secrets.

## Task Tracking

See `task.md` in this root for the full implementation checklist.
