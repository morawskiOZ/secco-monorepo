# Project Brief — Secco Studio

## Context

**Secco Studio** is an interior design studio in Wrocław, Poland. Website at www.seccostudio.com.

This mono-root (`secco/`) contains three separate, independently deployable repositories:

```
secco/
├── secco_LP/        # Public website + 14-question valuation form
├── secco_cms/       # Admin CMS (content editing, image uploads, deploy trigger)
└── secco_renderer/  # Shared npm package (markdown renderer, types)
```

The **master task list** is `task.md` at this root. Read it completely before starting any implementation work.

---

## v1.1 — What's Live

**secco_LP** is deployed and working:
- Landing page (`/`) — prerendered static HTML
- 14-question valuation form (`/form`) — SvelteKit SPA
- Privacy policy (`/polityka-prywatnosci`)
- Go backend embeds the static build, serves on `:8080`
- Docker → MicroK8s via ArgoCD

---

## v2.0 — What's Being Built

### secco_LP — New Content Pages

Expanding the existing website with prerendered content sections:

- **Navbar** — sticky, warm beige, scroll shadow, hamburger on mobile (✅ done)
- **Treści (`/tresci`)** — blog listing + article pages, tag filter, date sort (✅ done)
- **Projekty (`/projekty`)** — portfolio grid + project detail with Lightbox (✅ done)
- **Proces projektowy (`/proces-projektowy`)** — single design process page (✅ done)
- **Content pipeline** — GH Action downloads R2 snapshot before build (✅ done)
- **ResponsiveImage** — Cloudflare Image Transformations via srcset (✅ done)

All content pages are prerendered at build time from JSON files. No runtime database.

### secco_cms — New App

Admin-only CMS built from scratch:

- SvelteKit SPA + Go backend on `:8081`
- SQLite for content drafts + image metadata
- TipTap WYSIWYG editor → markdown output
- Preview pane using `@secco/render` — identical rendering to public website
- Single admin auth: bcrypt + JWT in httpOnly cookie
- R2 image upload via AWS S3 SDK
- **Deploy button:** exports published content → R2 snapshot → triggers GH Actions → site rebuilds

### secco_renderer — New Shared Package

npm package extracted from secco_LP to avoid rendering drift between website and CMS preview:

- `renderMarkdown()` — marked + R2-aware `<picture>` tags with CF Image Transformations
- `ContentRenderer.svelte` — Svelte component with typography CSS
- Shared TypeScript interfaces: `Article`, `Project`, `ProcessPage`
- Installed via git dependency (no npm registry): `npm install github:org/secco-render#v1.0.0`

---

## Key Architecture Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Content at runtime | Fully static — no DB on website | Perfect Lighthouse, zero infra |
| Content pipeline | CMS → R2 snapshot → GH Actions → prerender | Decoupled, rollback-friendly |
| Image hosting | Cloudflare R2 + Image Transformations | Cheapest, CDN included, S3-compatible |
| Preview rendering | Shared npm package (secco_renderer) | Single source of truth, no drift |
| CMS auth | Single admin, JWT + bcrypt, env vars | No user management complexity |
| CMS database | SQLite via modernc.org/sqlite | Pure Go, no CGO, zero infra |
| Package distribution | GitHub git dependency, tagged releases | No npm registry setup needed |
| CMS editor | TipTap (MIT free core) | WYSIWYG + markdown output, no lock-in |

---

## Deploy Flow

```
Admin clicks Deploy
       │
       ▼
CMS exports published content from SQLite
       │
       ▼
Uploads to R2:
  snapshots/{timestamp}.json   ← versioned backup
  snapshots/latest.json        ← overwritten (used by GH Action)
       │
       ▼
CMS calls GitHub Actions workflow dispatch API
  (input: snapshot_timestamp)
       │
       ▼
GH Action downloads snapshot from R2
  Splits into:
    src/lib/content/articles.json
    src/lib/content/projects.json
    src/lib/content/process.json
       │
       ▼
SvelteKit build — prerenders all content pages to static HTML
       │
       ▼
Go binary embeds build/ via embed.FS
Docker image pushed to GHCR
       │
       ▼
ArgoCD detects new image tag → deploys to K8s
```

---

## Design System

```
Primary background: #EBD5C9 (warm beige)
Accent:            #2E3192 (deep navy blue)
Secondary accent:  #7F7F80 (medium gray)
Page background:   #fafafa (off-white)
Text:              #454545 (charcoal gray)
Navbar height:     64px desktop / 56px mobile (--navbar-height CSS var)
Fonts:             "Open Sans", sans-serif
```

Design tokens live in `secco_LP/src/lib/styles/tokens.css`.

---

## Deployment Targets

| App | URL | Host |
|-----|-----|------|
| secco_LP | www.seccostudio.com | MicroK8s via ArgoCD |
| secco_cms | cms.seccostudio.com | MicroK8s via ArgoCD |
| Images/assets | assets.seccostudio.com | Cloudflare R2 (custom domain) |

K8s secrets managed in `appme-k8s` repo via SealedSecrets + `seal.sh`. Never commit raw secrets.

---

## Constraints

- **Lighthouse 100/100/100/100** on all prerendered pages
- **Minimal server resources** — Go backend ~10MB RAM, no runtime DB on website
- **Image quality** — 90-95% for portfolio/gallery, 85% for thumbnails
- **All UI text in Polish** (labels, page content, error messages)
- **GDPR compliant** — Art. 6(1)(b), info clause near submit, link to privacy policy
- **Mobile-first responsive design**

---

## Implementation Order

Follow the phases in `task.md` sequentially:

1. **Phases 1–6** — Navbar + content pages + image components (secco_LP) ✅
2. **Phase 6b** — Extract secco_renderer package
3. **Phases 7–14** — CMS app (secco_cms): scaffold, auth, editor, images, deploy
4. **Phase 15** — GH Action update (already code-complete)
5. **Phase 16** — Deployment configs (Dockerfile, K8s, ArgoCD)
6. **Phase 17** — End-to-end testing + polish

---

## Important Files

| File | Purpose |
|------|---------|
| `task.md` | Complete implementation checklist (phases 1–17) |
| `secco_LP/CLAUDE.md` | Website conventions, dev commands, env vars |
| `secco_cms/CLAUDE.md` | CMS stack, auth, schema, env vars |
| `secco_renderer/CLAUDE.md` | Package purpose, exports, versioning |
| `secco_LP/docs/cloudflare-r2.md` | R2 setup, image delivery, transformations |
| `secco_LP/docs/deploy-flow.md` | Content publishing pipeline details |
| `secco_LP/docs/cms-setup.md` | CMS architecture and deployment |
| `secco_LP/CHANGELOG.md` | Track all user-facing changes |
