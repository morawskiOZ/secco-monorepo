# Secco Studio — Website Expansion & CMS

> **Master task list for the mono-root.** Lives at `secco/task.md`.
> The individual repos (`secco_LP/`, `secco_cms/`, `secco_renderer/`) each have their own `CLAUDE.md`.

## Architecture Overview

```
                         ┌──────────────────────────┐
                         │   CMS (secco_cms)        │
                         │   SvelteKit + Go + SQLite │
                         │                          │
                         │  Edit → Preview → Deploy  │
                         └────┬──────────┬──────────┘
                              │          │
                    ┌─────────┘          └──────────┐
                    ▼                               ▼
           ┌────────────────┐             ┌─────────────────┐
           │ Cloudflare R2  │             │  GitHub Actions  │
           │                │             │                  │
           │ /images/       │◄────────────│  1. Download     │
           │  (originals)   │  fetch      │     snapshot     │
           │                │  content    │     from R2      │
           │ /snapshots/    │             │  2. SvelteKit    │
           │  latest.json   │             │     prerender    │
           │  {ts}.json     │             │  3. Go build     │
           └────────────────┘             │  4. Push image   │
                    │                     └────────┬─────────┘
                    │ CDN delivery                  │
                    ▼                               ▼
           ┌────────────────┐             ┌─────────────────┐
           │  Website       │             │  ArgoCD         │
           │  (secco_LP)    │◄────────────│  Detect new tag │
           │                │  deploy     │  Sync to K8s    │
           │  Go + static   │             └─────────────────┘
           │  All pages     │
           │  prerendered   │
           │  No database   │
           └────────────────┘
```

### Key Principle: Pre-built Static Content

**Content is baked into the build, not served dynamically.** The website has NO database and NO content API at runtime. All content pages (blog, projects, process) are prerendered to static HTML during the SvelteKit build, just like the landing page and privacy policy today.

**Deploy flow:**
1. Admin edits content in CMS (drafts stored in CMS's SQLite)
2. Admin clicks "Deploy" in CMS
3. CMS exports published content → uploads JSON snapshot to R2
4. CMS triggers GitHub Actions workflow via API
5. GH Action downloads snapshot from R2, feeds it into SvelteKit build
6. SvelteKit prerenders all content pages to static HTML
7. Go binary embeds all static files
8. Docker image pushed to GHCR
9. ArgoCD detects new image → deploys to K8s

### System Components

| Component | Role | Tech | Database |
|-----------|------|------|----------|
| secco_LP (website) | Public website + form handler | SvelteKit + Go | **None** (fully static) |
| secco_cms | Content editing + deploy trigger | SvelteKit + Go | SQLite (drafts + images metadata) |
| secco_renderer | Shared rendering package | npm (Svelte + marked) | N/A |
| Cloudflare R2 | Image storage + content snapshots | S3-compatible API | N/A |
| GitHub Actions | Build pipeline | Existing workflow | N/A |
| ArgoCD | K8s deployment | Existing | N/A |

### Routes (All Prerendered)

| Route | Type | Description |
|-------|------|-------------|
| `/` | Prerendered | Landing page (existing) |
| `/form` | SPA | Valuation form (existing, only SPA page) |
| `/polityka-prywatnosci` | Prerendered | Privacy policy (existing) |
| `/tresci` | Prerendered | Blog listing page |
| `/tresci/[slug]` | Prerendered | Individual blog article |
| `/projekty` | Prerendered | Projects portfolio grid |
| `/projekty/[slug]` | Prerendered | Individual project detail |
| `/proces-projektowy` | Prerendered | Design process page |

---

## Resolved Decisions

### Original (v1.0)

| Question | Decision |
|----------|----------|
| File uploads | Single file UI, backend accepts `[]File`. |
| SMTP provider | Gmail SMTP with app password. |
| "Other" on Q10 | Selecting "Other" reveals a free-text input. |
| Required fields | Only first name + email required. |
| GDPR | Art. 6(1)(b) — no consent checkbox. Info clause + privacy policy link. |
| Bot protection | Cloudflare Turnstile (form page only). |

### New (v2.0)

| Question | Decision |
|----------|----------|
| Content model | **Pre-built static.** Content baked into SvelteKit build. No runtime DB on website. |
| Content storage (CMS) | SQLite via `modernc.org/sqlite` (pure Go). Lives only in CMS app. |
| Image hosting | **Cloudflare R2** + Image Transformations. Cheapest option — ~free for our volume. |
| Image quality | 90-95% for portfolio/gallery, 85% for thumbnails. Originals stored in R2. |
| CMS approach | **Built from scratch.** SvelteKit + Go + TipTap editor. Full control over deploy flow. |
| CMS editor | **TipTap** (free MIT core). WYSIWYG with markdown output. |
| CMS auth | Single admin user, env var credentials (bcrypt), JWT sessions. |
| Deploy trigger | CMS "Deploy" button → R2 snapshot → GitHub Actions API → build → ArgoCD. |
| Content backup | R2 snapshots: `snapshots/latest.json` + timestamped versions. |
| SEO | All content pages prerendered — perfect SEO by default. OG tags baked in. |
| Lightbox | Yes — full-screen image viewer on project pages. |
| Shared renderer | **secco_renderer** npm package — single source of truth for both apps. |

---

## Database Schema (CMS only)

```sql
-- Lives in secco_cms SQLite, NOT in secco_LP

CREATE TABLE content (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    type         TEXT NOT NULL CHECK(type IN ('article', 'project', 'process')),
    slug         TEXT NOT NULL UNIQUE,
    title        TEXT NOT NULL,
    summary      TEXT,
    body         TEXT NOT NULL DEFAULT '',
    cover_image  TEXT,
    status       TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft', 'published')),
    sort_order   INTEGER DEFAULT 0,
    published_at TIMESTAMP,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE images (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    r2_key        TEXT NOT NULL UNIQUE,
    filename      TEXT NOT NULL,
    alt_text      TEXT,
    width         INTEGER,
    height        INTEGER,
    size_bytes    INTEGER,
    content_type  TEXT,
    public_url    TEXT NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE content_images (
    content_id INTEGER REFERENCES content(id) ON DELETE CASCADE,
    image_id   INTEGER REFERENCES images(id) ON DELETE CASCADE,
    PRIMARY KEY (content_id, image_id)
);

CREATE INDEX idx_content_type_status ON content(type, status);
CREATE INDEX idx_content_slug ON content(slug);
CREATE INDEX idx_content_type_sort ON content(type, sort_order);
```

---

## Content Snapshot Format

The CMS exports this JSON to R2 on deploy. The GH Action downloads it and places files in `src/lib/content/` for SvelteKit to prerender.

```json
{
  "exported_at": "2026-03-20T14:00:00Z",
  "articles": [
    {
      "slug": "jak-wybrac-styl-wnetrzarski",
      "title": "Jak wybrać styl wnętrzarski?",
      "summary": "Krótki przewodnik po najpopularniejszych stylach...",
      "body": "# Jak wybrać styl...\n\nMarkdown content with R2 image URLs...",
      "cover_image": "https://assets.seccostudio.com/images/cover-abc123.jpg",
      "published_at": "2026-03-19T10:00:00Z"
    }
  ],
  "projects": [
    {
      "slug": "mieszkanie-na-krakowskiej",
      "title": "Mieszkanie na Krakowskiej",
      "summary": "Nowoczesne wnętrze w centrum Wrocławia",
      "body": "## O projekcie\n\n![salon](https://assets.seccostudio.com/images/salon.jpg)\n\nOpis...",
      "cover_image": "https://assets.seccostudio.com/images/cover-def456.jpg",
      "sort_order": 1,
      "published_at": "2026-03-15T12:00:00Z"
    }
  ],
  "process": {
    "title": "Proces projektowy",
    "body": "## 1. Konsultacja\n\n![konsultacja](https://assets.seccostudio.com/images/konsultacja.jpg)\n\nOpis..."
  }
}
```

### SvelteKit Build-Time Integration

Content JSON is split into files that SvelteKit reads during prerendering:

```
src/lib/content/
├── articles.json      # array of articles
├── projects.json      # array of projects
└── process.json       # single process object
```

Pages use `+page.ts` load functions to import these:
```typescript
// src/routes/tresci/+page.ts
import articles from '$lib/content/articles.json';
export const prerender = true;
export function load() {
  return { articles };
}
```

Dynamic routes prerendered via `entries()`:
```typescript
// src/routes/tresci/[slug]/+page.ts
import articles from '$lib/content/articles.json';
export const prerender = true;
export function entries() {
  return articles.map(a => ({ slug: a.slug }));
}
export function load({ params }) {
  return { article: articles.find(a => a.slug === params.slug) };
}
```

---

## Implementation Phases

### Phase 1: Navbar & Layout Update ✅
**Goal:** Persistent navigation across all pages. Mobile-responsive hamburger menu.

- [x] Create `src/lib/components/Navbar.svelte`
  - Logo "secco.studio" (left) → links to `/`
  - Nav links (right): Projekty, Treści, Proces, Formularz wyceny
  - "Formularz wyceny" styled as CTA button (navy bg, white text)
  - Active link indicator (bottom border, navy color)
  - Sticky positioning, warm beige background (#EBD5C9)
  - Subtle box-shadow on scroll (via IntersectionObserver or scroll event)
- [x] Create `src/lib/components/MobileMenu.svelte`
  - Hamburger icon (3-line, CSS-only)
  - Full-height slide-down overlay
  - Links centered, larger font
  - Close on link click or overlay tap
  - Body scroll lock when open
- [x] Update `src/routes/+layout.svelte`
  - Include Navbar above `<slot />`
  - Adjust page padding-top for sticky navbar height
- [x] Responsive breakpoint: 768px (hamburger below, full links above)
- [x] Test navbar on all existing pages (landing, form, privacy)
- [ ] Ensure Lighthouse scores remain 100 on landing page

### Phase 2: Content Data Layer & Markdown Rendering ✅
**Goal:** Static content loading and rendering infrastructure.

- [x] Create `src/lib/content/` directory
  - `articles.json` — empty array placeholder `[]`
  - `projects.json` — empty array placeholder `[]`
  - `process.json` — null placeholder `null`
  - Add to `.gitignore` (generated at build time from R2 snapshot)
  - Add fallback/default files for dev mode when no content exists
- [x] Install `marked` for markdown → HTML conversion on website
  - Note: skipped DOMPurify — CMS is admin-only, content is trusted
- [x] Create `src/lib/utils/markdown.ts`
  - Configure `marked` renderer for custom image handling
  - Wrap R2 image URLs in responsive `<picture>` tags with CF Image Transformations
  - Add `loading="lazy"` and `decoding="async"` to all images
- [x] Create `src/lib/components/content/ContentRenderer.svelte`
  - Takes markdown string, renders to sanitized HTML
  - Applies typography styles (article-body class)
  - Responsive images via Cloudflare Image Transformations
- [x] Typography styles for rendered content:
  - Headings (h1-h4), paragraphs, lists, blockquotes
  - Code blocks, inline code
  - Image captions
  - Link styling

### Phase 3: Website Content Pages — Treści (Blog) ✅
**Goal:** Blog listing and article pages, prerendered at build time.

- [x] Create `src/lib/components/content/ArticleCard.svelte`
  - Cover image (thumbnail size), title, summary, date
  - Tags as clickable filter chips
  - Hover effect: subtle scale + shadow
  - Link to `/tresci/[slug]`
- [x] Create `src/routes/tresci/+page.ts`
  - `export const prerender = true`
  - Load function imports `articles.json`
- [x] Create `src/routes/tresci/+page.svelte`
  - 3-column grid (desktop), 2-column (tablet), 1-column (mobile)
  - Tag filter bar + newest/oldest date sort (client-side, URL-shareable)
  - Empty state: "Brak artykułów" message
  - SEO meta tags + comprehensive Polish interior design keywords
- [x] Create `src/routes/tresci/[slug]/+page.ts`
  - `export const prerender = true`
  - `entries()` returns all article slugs
  - Load function finds article by slug
- [x] Create `src/routes/tresci/[slug]/+page.svelte`
  - Hero cover image (full width, max-height 480px)
  - Title (h1), date, clickable tag links
  - Markdown content via ContentRenderer
  - Max-width container (~720px) for reading comfort
  - Back link "← Wróć do treści"
  - OG tags with cover image, title, summary, article:published_time
  - Keywords meta built from article tags + base interior design keywords
- [x] Handle missing articles gracefully (404)

### Phase 4: Website Content Pages — Projekty (Projects) ✅
**Goal:** Portfolio grid and project detail pages. Image-heavy layout.

- [x] Create `src/lib/components/content/ProjectCard.svelte`
  - Large cover image, title, short summary
  - Aspect ratio: 3:2 (landscape, interior design photos)
  - Hover: slight zoom + overlay with title/summary
  - Link to `/projekty/[slug]`
- [x] Create `src/lib/components/content/Lightbox.svelte`
  - Full-screen overlay with image
  - Close button, click-outside-to-close, Escape key
  - Previous/next navigation (arrows + keyboard)
  - Image counter (3/12)
  - High-quality image variant (w=2400, quality=95)
  - Swipe support on mobile (touch events)
- [x] Create `src/routes/projekty/+page.ts`
  - `export const prerender = true`
  - Load from `projects.json`
- [x] Create `src/routes/projekty/+page.svelte`
  - 2-column grid (desktop), 1-column (mobile)
  - Larger cards than blog (portfolio emphasis)
  - Empty state message
- [x] Create `src/routes/projekty/[slug]/+page.ts`
  - `export const prerender = true`
  - `entries()` returns all project slugs
- [x] Create `src/routes/projekty/[slug]/+page.svelte`
  - Title (h1), initial description
  - Markdown content with interspersed images and text
  - Click any image → opens Lightbox
  - Back link "← Wróć do projektów"
  - OG tags with cover image
- [x] Ensure project images use high-quality variants (quality=90+)

### Phase 5: Website Content Pages — Proces Projektowy ✅
**Goal:** Single page describing the design process.

- [x] Create `src/routes/proces-projektowy/+page.ts`
  - `export const prerender = true`
  - Load from `process.json`
- [x] Create `src/routes/proces-projektowy/+page.svelte`
  - Title "Proces projektowy" (h1)
  - Markdown content rendered via ContentRenderer
  - Sections with images (numbered steps of the design process)
  - Same max-width container as articles (~720px)
  - Empty state: placeholder text until content added via CMS
  - OG tags
- [x] Style process sections (via ContentRenderer generic markdown styles)

### Phase 6: Responsive Image Delivery (Cloudflare R2 + Transformations) ✅ (code done)
**Goal:** Images stored in R2, served via Cloudflare CDN with on-the-fly transformations.

- [x] Create `src/lib/components/content/ResponsiveImage.svelte`
  - Takes R2 public URL, alt text, optional aspect ratio
  - Renders `<picture>` with srcset using CF Image Transformations:
    - `(max-width: 480px)` → `/cdn-cgi/image/w=480,quality=90,format=auto/{url}`
    - `(max-width: 768px)` → `/cdn-cgi/image/w=768,quality=90,format=auto/{url}`
    - `(max-width: 1200px)` → `/cdn-cgi/image/w=1200,quality=90,format=auto/{url}`
    - Default → `/cdn-cgi/image/w=1920,quality=95,format=auto/{url}`
  - `loading="lazy"`, `decoding="async"`
  - Explicit width/height for CLS prevention
  - Fade-in on load
- [x] Update markdown renderer to detect R2 image URLs and wrap in responsive picture tags
- [ ] Configure Cloudflare zone for Image Transformations (infrastructure — manual):
  - Enable Image Resizing on the zone (requires paid plan or use Polish)
  - Custom domain on R2 bucket for public access
- [ ] Test image quality at various breakpoints (pending CF zone setup)

### Phase 6b: secco_renderer Package ✅ (local tarball)
**Goal:** Extract shared rendering primitives into a versioned npm package consumed by both apps.

> **Status (2026-04-15):** Package built as `@secco/render` v1.0.0, shipped to secco_LP as a local tarball (`file:../secco_renderer/secco-render-1.0.0.tgz`). GitHub publishing deferred until a remote repo exists. secco_cms install deferred until Phase 7.

- [x] Create `secco_renderer/` at mono-root (empty skeleton — no package.json yet)
- [x] Initialize `package.json` (name, version, exports, peerDependencies)
- [x] Move `renderMarkdown()` and `ContentRenderer.svelte` from secco_LP into package
- [x] Create `src/lib/types.ts` — Article, Project, ProcessPage interfaces
- [x] Create `src/lib/markdown.ts` — `renderMarkdown()` using marked, R2-aware picture tags
- [x] Create `src/lib/ContentRenderer.svelte` — Svelte component + typography CSS
- [x] Create `src/lib/index.ts` — re-exports all of the above
- [x] Configure `@sveltejs/package` for building; ship raw `.svelte` files
- [x] Pack locally with `npm pack` → `secco-render-1.0.0.tgz`
- [ ] Tag initial release: `v1.0.0` (deferred — no GitHub remote yet)
- [x] Install in secco_LP via local tarball (`file:../secco_renderer/secco-render-1.0.0.tgz`)
- [ ] Install in secco_cms (deferred — secco_cms still empty, picks up in Phase 7)
- [x] Remove duplicated rendering code from secco_LP

### Phase 7: CMS App — Scaffolding (secco_cms) ✅
**Goal:** Initialize the CMS app at `secco/secco_cms/`.

> **Status (2026-04-16):** Fully scaffolded. SvelteKit frontend (SPA mode, adapter-static, Svelte 5 runes) and Go backend (stdlib HTTP, SQLite, embed.FS) both compile and pass checks.

- [x] Create directory `secco/secco_cms/`
- [x] Initialize SvelteKit project with adapter-static (`package.json`, `svelte.config.js`, `vite.config.ts`, `app.css`, `app.html`, `+layout.svelte`, `+page.svelte`)
- [x] Initialize Go backend in `secco_cms/backend/` (`go.mod`, `main.go`)
- [x] Project structure — all routes, components, API clients, stores created
- [x] Copy design tokens from secco_LP for preview consistency
- [x] Set up Go backend with SQLite (modernc.org/sqlite, WAL mode, auto-migrate)
- [x] Vite proxy: `/api/*` → Go backend on :8081

### Phase 8: CMS — Authentication ✅
**Goal:** Simple, secure single-admin login.

> **Status (2026-04-16):** Auth fully implemented and tested (11 Go tests). JWT+bcrypt, httpOnly cookies, auth middleware, SvelteKit login page + route guard.

- [x] Go backend `auth.go`: login/logout/check endpoints, JWT HS256 24h, bcrypt, authMiddleware
- [x] Login page (`login/+page.svelte`): form + error + redirect
- [x] Auth store (`stores/auth.ts`): reactive Svelte 5 runes, checkAuth/login/logout
- [x] Route guard in `+layout.svelte`: redirect to /login when unauthenticated

### Phase 9: CMS — Content CRUD API ✅
**Goal:** Go backend for content management backed by SQLite.

> **Status (2026-04-16):** Full CRUD implemented and tested (15 Go tests including slugify subtests). Polish diacritics → ASCII slug generation, process uniqueness constraint, slug uniqueness validation.

- [x] `backend/database.go`: InitDB with WAL, foreign keys, schema migration (content, images, content_images)
- [x] `backend/content.go`: list (filter by type/status), get, create, update, delete, publish, draft
- [x] Auto-generate slug from title (Polish diacritics → ASCII)
- [x] Validate: only one `process` entry allowed, slug uniqueness
- [x] JSON request/response format with proper HTTP status codes
- [x] SvelteKit API client (`src/lib/api/content.ts`) with typed Content/ContentInput interfaces
- [x] Working content list pages (tresci, projekty) with status badges + actions

### Phase 10: CMS — Image Management (Cloudflare R2) ✅
**Goal:** Upload images to R2, manage metadata in SQLite.

> **Status (2026-04-16):** Go backend images.go with R2 S3 SDK, objectStore interface for testability, 15 tests. Frontend ImageUpload + ImagePicker + images page all working.

- [x] `backend/images.go`: upload (R2 via S3 SDK, UUID key, dimension extraction), list (pagination + search), delete (R2 + DB), update alt_text
- [x] `objectStore` interface for S3 mock testing
- [ ] R2 bucket configuration (infrastructure — manual): custom domain, CORS
- [x] Image list endpoint with pagination and search by filename
- [x] Frontend: ImageUpload.svelte (drag-drop, validation, multi-file)
- [x] Frontend: ImagePicker.svelte (modal, grid, inline upload, alt text, insert)
- [x] Frontend: images/+page.svelte (gallery, search, detail modal, editable alt, copy URL, delete)

### Phase 11: CMS — TipTap Editor ✅
**Goal:** WYSIWYG content editor with image upload and preview.

> **Status (2026-04-16):** TipTap v3.22.3 integrated with Svelte 5 vanilla JS API. ContentEditor shared component with auto-save, HTML↔Markdown via Turndown/Marked. All editor pages implemented.

- [x] Install TipTap dependencies (@tiptap/core, starter-kit, image, link, placeholder, pm)
- [x] Editor.svelte: TipTap WYSIWYG with toolbar (H1-H3, Bold, Italic, Lists, Quote, Link, Image, HR), active states
- [x] ContentEditor.svelte: shared editor form (title, slug, summary, cover image, sort_order, status, auto-save 5s debounce, unsaved indicator)
- [x] HTML↔Markdown conversion via Turndown + Marked (src/lib/utils/markdown.ts)
- [x] Preview.svelte: renders markdown via @secco/render's renderMarkdown()
- [x] Article editor pages (tresci/new, tresci/[id])
- [x] Project editor pages (projekty/new, projekty/[id])
- [x] Process editor page (proces — single entry, auto-loads or creates)

### Phase 12: CMS — Image Manager Page ✅
**Goal:** Standalone page to upload, browse, and manage images.

> **Status (2026-04-16):** Full image gallery with upload, search, detail modal, alt text editing, copy URL, delete.

- [x] images/+page.svelte: grid view, upload zone, detail modal, search, pagination
- [x] ImageUpload.svelte: drag-drop, file type/size validation, multi-file, status tracking
- [x] ImagePicker.svelte: modal grid, inline upload, alt text, insert/cancel

### Phase 13: CMS — Deploy System ✅
**Goal:** "Deploy" button exports content, backs up to R2, triggers GH Action.

> **Status (2026-04-16):** Go deploy.go with snapshot building, R2 upload, GH Actions trigger, deploys table. 9 tests. Frontend deploy button in sidebar + dashboard status.

- [x] `backend/deploy.go`: snapshot export (articles/projects/process), R2 upload (timestamped + latest), GH Actions workflow dispatch
- [x] `deploys` table in database.go for tracking deploy history
- [x] `GET /api/deploy/status`: last deploy info, published counts, has_changes flag
- [x] Deploy button in Sidebar.svelte (already wired)
- [x] Dashboard deploy status section

### Phase 14: CMS — Dashboard & Lists ✅
**Goal:** Dashboard overview and content listing pages.

> **Status (2026-04-16):** Sidebar, dashboard, and content list pages all implemented in Phases 7-9 and enhanced in this wave.

- [x] Sidebar.svelte: logo, nav links, active highlight, mobile hamburger, deploy button, logout
- [x] Dashboard: summary cards (articles, projects, images count), quick actions, recent changes, deploy status
- [x] Content list pages (tresci, projekty): table with status badges, publish/unpublish, edit, delete
- [ ] Reorder projects via drag-drop (sort_order) — deferred, manual sort_order field available in editor

### Phase 15: GitHub Actions Update ✅ (code done)
**Goal:** Update build pipeline to download content from R2 before building.

- [x] Update `.github/workflows/build-push.yml`:
  - Add workflow_dispatch input: `snapshot_timestamp` (optional, defaults to latest)
  - New step: Download content from R2
    - Uses AWS CLI
    - Downloads `snapshots/latest.json` (or specific timestamp)
    - Splits into `src/lib/content/articles.json`, `projects.json`, `process.json`
  - Existing steps: npm build, go build, docker push
- [ ] Add GitHub secrets (manual — done in GitHub repo settings):
  - `R2_ACCOUNT_ID`
  - `R2_ACCESS_KEY_ID`
  - `R2_SECRET_ACCESS_KEY`
  - `R2_BUCKET_NAME`
- [x] Add `src/lib/content/*.json` to `.gitignore` (generated at build time)
- [x] Create fallback content files for local development (committed as `src/lib/content/*.example.json`)

### Phase 16: Deployment Updates
**Goal:** Docker, Taskfile, CI/CD for secco_cms (K8s is in appme-k8s repo).

- [x] secco_LP Dockerfile — **no changes needed** (no DB, still just static + Go)
- [x] Create `Dockerfile` for secco_cms:
  - Multi-stage: node:24-alpine → golang:1.25-alpine → scratch
  - Pure Go SQLite driver, no extra deps needed in scratch
  - Copies @secco/render tarball, rewrites path via sed
  - Expose port 8081
- [x] Create `Taskfile.yml` for secco_cms:
  - Matches secco_LP pattern (login, build-push, dev tasks)
  - Added `prepare` task to copy @secco/render tarball before Docker build
- [x] Create `docker-compose.yml` for local dev:
  - Frontend (node:24-alpine, port 5175→5173) + backend (golang:1.25-alpine, port 8081)
  - Named volumes for node_modules and cms-data (SQLite)
  - Backend reads env from .env file
- [x] Create `.github/workflows/build-push.yml`:
  - workflow_dispatch with semver version input
  - Downloads @secco/render tarball from R2 before build
  - Uses Task for login + build-push, creates git tag
- [x] Create `.dockerignore` matching secco_LP pattern
- [x] Update `.gitignore` for build artifacts (tarball, Go binary)
- N/A K8s manifests, ArgoCD, DNS — managed in appme-k8s repo

### Phase 17: Polish & Testing
**Goal:** Ensure everything works end-to-end.

- [ ] End-to-end flow test:
  - Create article in CMS → save as draft → preview → publish
  - Click Deploy → R2 snapshot created → GH Action triggered → site rebuilt
  - New article visible on website as static page
  - Image uploaded → R2 → visible in content → transformations working
- [ ] Responsive testing: all new pages on mobile/tablet/desktop
- [ ] Lighthouse audit: all prerendered content pages (target 100/100/100/100)
- [ ] Accessibility: keyboard navigation, screen reader, ARIA
- [ ] Cross-browser: Chrome, Firefox, Safari
- [ ] Error handling: R2 upload failures, GH Action failures, network errors
- [ ] Loading states in CMS: skeletons/spinners
- [ ] Empty states on website: meaningful messages when no content
- [ ] 404 pages for non-existent slugs
- [ ] Rollback test: deploy older snapshot → site reverts

---

## Navbar Design

```
Desktop (>768px):
┌────────────────────────────────────────────────────────────────────────┐
│  secco.studio          Projekty    Treści    Proces    ┌────────────┐ │
│  (wordmark, link to /) (links, charcoal #454545)       │ Formularz  │ │
│                        (navy underline on active)      │  wyceny    │ │
│                                                        └────────────┘ │
│                                                        (CTA, navy bg) │
└────────────────────────────────────────────────────────────────────────┘

Mobile (<768px):
┌──────────────────────────────────┐
│  secco.studio               ☰   │
└──────────────────────────────────┘
         ↓ (on hamburger click)
┌──────────────────────────────────┐
│  secco.studio               ✕   │
├──────────────────────────────────┤
│                                  │
│          Projekty                │
│          Treści                  │
│          Proces                  │
│       Formularz wyceny           │
│                                  │
└──────────────────────────────────┘
```

**Styling details:**
- Position: `sticky`, top: 0, z-index: 100
- Background: `#EBD5C9` (warm beige)
- Height: 64px (desktop), 56px (mobile)
- Logo: font-weight 600, font-size 1.25rem, color `#454545`
- Links: font-weight 400, font-size 0.9rem, letter-spacing 0.5px, text-transform uppercase
- Active link: navy bottom border (2px)
- Hover: color transition to `#2E3192` (navy)
- CTA button: background `#2E3192`, color white, border-radius 4px, padding 8px 20px
- Mobile menu: full-viewport overlay, background `#EBD5C9`, links centered vertically
- Transition: menu slides down 300ms ease

---

## Page Layouts

### Treści (Blog) — `/tresci`
- Page title: "Treści" (h1)
- Grid: 3 columns desktop, 2 tablet, 1 mobile
- Cards: cover image (4:3), title, summary (2 lines max), date
- Gap: 32px between cards
- Max-width: 1200px container

### Article — `/tresci/[slug]`
- Cover image: full-width (max-height 480px, object-fit cover)
- Content container: max-width 720px, centered
- Title: h1, margin-bottom 8px
- Date: small gray text below title
- Body: rendered markdown with good typography
- Images in body: full container width, responsive via CF Transformations
- Back link: "← Wróć do treści"

### Projekty — `/projekty`
- Page title: "Projekty" (h1)
- Grid: 2 columns desktop, 1 mobile
- Cards: large cover image (3:2), title overlay on hover
- Gap: 24px
- Max-width: 1200px container

### Project Detail — `/projekty/[slug]`
- Title: h1
- Initial description paragraph
- Markdown content with interspersed images and text
- Click any image → lightbox with full-quality view
- Lightbox: arrows, counter, swipe, Escape to close
- Back link: "← Wróć do projektów"

### Proces Projektowy — `/proces-projektowy`
- Title: "Proces projektowy" (h1)
- Content container: max-width 720px, centered
- Sections with headings, descriptions, and images
- Similar layout to article page

---

## Cloudflare R2 — Image Delivery

### Storage
Images stored in R2 bucket with public custom domain: `assets.seccostudio.com`

Key format: `images/{uuid}-{filename}`

### Image Transformations (via Cloudflare zone)
```
Original:    https://assets.seccostudio.com/images/abc123-salon.jpg
Thumbnail:   /cdn-cgi/image/w=400,h=300,fit=cover,quality=85,format=auto/https://assets.seccostudio.com/images/abc123-salon.jpg
Medium:      /cdn-cgi/image/w=800,quality=90,format=auto/https://assets.seccostudio.com/images/abc123-salon.jpg
Large:       /cdn-cgi/image/w=1200,quality=90,format=auto/https://assets.seccostudio.com/images/abc123-salon.jpg
Hero:        /cdn-cgi/image/w=1920,quality=95,format=auto/https://assets.seccostudio.com/images/abc123-salon.jpg
Full:        /cdn-cgi/image/w=2400,quality=95,format=auto/https://assets.seccostudio.com/images/abc123-salon.jpg
OG:          /cdn-cgi/image/w=1200,h=630,fit=cover,quality=85,format=auto/https://assets.seccostudio.com/images/abc123-salon.jpg
```

`format=auto` serves WebP or AVIF based on browser `Accept` header.

### Content Snapshots
Snapshots stored in R2 under `snapshots/` prefix:
```
snapshots/latest.json                    → always overwritten (used by GH Action)
snapshots/2026-03-20T14-00-00Z.json      → timestamped backup (never overwritten)
```

---

## Environment Variables

### secco_LP (NO CHANGES from v1.1)
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=secco.studio@gmail.com
SMTP_PASS=<gmail-app-password>
RECIPIENT_EMAIL=secco.studio@gmail.com
PUBLIC_TURNSTILE_SITE_KEY=<sitekey>
TURNSTILE_SECRET_KEY=<secret>
PORT=8080
```

### secco_cms (new app)
```env
# Admin auth
CMS_ADMIN_USER=admin
CMS_ADMIN_PASS_HASH=<bcrypt-hash>
CMS_JWT_SECRET=<random-secret>

# Database
DATABASE_PATH=./data/cms.db

# Cloudflare R2
R2_ACCOUNT_ID=<cf-account-id>
R2_ACCESS_KEY_ID=<r2-access-key>
R2_SECRET_ACCESS_KEY=<r2-secret-key>
R2_BUCKET_NAME=secco-assets
R2_PUBLIC_URL=https://assets.seccostudio.com

# GitHub (for deploy trigger)
GITHUB_TOKEN=<personal-access-token>
GITHUB_REPO=owner/secco_LP
GITHUB_WORKFLOW_ID=build-push.yml

# Website preview
WEBSITE_PUBLIC_URL=https://seccostudio.com

# Server
PORT=8081
```

### GitHub Actions Secrets (new)
```
R2_ACCOUNT_ID
R2_ACCESS_KEY_ID
R2_SECRET_ACCESS_KEY
R2_BUCKET_NAME
```

---

## File Structure

### secco_LP (expanded — website)
```
secco_LP/
├── src/
│   ├── routes/
│   │   ├── +layout.svelte                      # Updated: includes Navbar
│   │   ├── +page.svelte                         # Landing (existing)
│   │   ├── +page.ts
│   │   ├── form/                                # Existing form (SPA)
│   │   ├── polityka-prywatnosci/                # Existing privacy policy
│   │   ├── tresci/
│   │   │   ├── +page.svelte                     # Blog listing (prerendered)
│   │   │   ├── +page.ts
│   │   │   └── [slug]/
│   │   │       ├── +page.svelte                 # Blog article (prerendered)
│   │   │       └── +page.ts
│   │   ├── projekty/
│   │   │   ├── +page.svelte                     # Projects grid (prerendered)
│   │   │   ├── +page.ts
│   │   │   └── [slug]/
│   │   │       ├── +page.svelte                 # Project detail (prerendered)
│   │   │       └── +page.ts
│   │   └── proces-projektowy/
│   │       ├── +page.svelte                     # Process page (prerendered)
│   │       └── +page.ts
│   ├── lib/
│   │   ├── components/
│   │   │   ├── Navbar.svelte                    # NEW
│   │   │   ├── MobileMenu.svelte                # NEW
│   │   │   ├── landing/                         # Existing
│   │   │   ├── form/                            # Existing
│   │   │   └── content/                         # NEW
│   │   │       ├── ContentRenderer.svelte
│   │   │       ├── ArticleCard.svelte
│   │   │       ├── ProjectCard.svelte
│   │   │       ├── Lightbox.svelte
│   │   │       └── ResponsiveImage.svelte
│   │   ├── content/                             # NEW (generated at build time, gitignored)
│   │   │   ├── articles.json
│   │   │   ├── projects.json
│   │   │   └── process.json
│   │   ├── utils/
│   │   │   └── markdown.ts                      # NEW
│   │   ├── config/
│   │   │   └── questions.ts                     # Existing
│   │   └── styles/
│   │       └── tokens.css                       # Existing
│   ├── app.css
│   ├── app.d.ts
│   └── app.html
├── static/                                      # Existing
├── backend/
│   ├── main.go                                  # Existing (no DB changes needed)
│   ├── handler.go                               # Existing
│   ├── email.go                                 # Existing
│   ├── turnstile.go                             # Existing
│   └── go.mod                                   # Existing (no new deps)
├── docs/
│   ├── cloudflare-r2.md
│   ├── cms-setup.md
│   └── deploy-flow.md
├── .github/workflows/
│   └── build-push.yml                           # Updated: R2 content download step
├── CHANGELOG.md
├── CLAUDE.md
├── Dockerfile                                   # No changes
└── docker-compose.yml                           # Updated: add CMS service
```

### secco_cms (new app)
```
secco_cms/
├── src/
│   ├── routes/
│   │   ├── +layout.svelte
│   │   ├── +page.svelte              # Dashboard
│   │   ├── login/+page.svelte
│   │   ├── tresci/
│   │   │   ├── +page.svelte          # Articles list
│   │   │   ├── new/+page.svelte
│   │   │   └── [id]/+page.svelte
│   │   ├── projekty/
│   │   │   ├── +page.svelte          # Projects list
│   │   │   ├── new/+page.svelte
│   │   │   └── [id]/+page.svelte
│   │   ├── proces/+page.svelte       # Process editor
│   │   └── images/+page.svelte       # Image manager
│   ├── lib/
│   │   ├── components/
│   │   │   ├── Sidebar.svelte
│   │   │   ├── Editor.svelte         # TipTap WYSIWYG
│   │   │   ├── Preview.svelte
│   │   │   ├── ImageUpload.svelte
│   │   │   └── ImagePicker.svelte
│   │   ├── api/
│   │   │   ├── content.ts
│   │   │   ├── images.ts
│   │   │   └── deploy.ts
│   │   └── stores/
│   │       └── auth.ts
│   └── app.css
├── backend/
│   ├── main.go
│   ├── database.go
│   ├── auth.go
│   ├── content.go
│   ├── images.go
│   ├── deploy.go
│   └── go.mod
├── data/                              # SQLite (gitignored)
│   └── cms.db
├── svelte.config.js
├── vite.config.ts
├── package.json
├── Dockerfile
├── CLAUDE.md
└── .env.example
```

---

## Modification Guide

| Change | Where to edit |
|--------|--------------|
| Add/edit form questions | `secco_LP/src/lib/config/questions.ts` |
| Change design tokens | `secco_LP/src/lib/styles/tokens.css` |
| Change navbar links | `secco_LP/src/lib/components/Navbar.svelte` |
| Change content page layout | `secco_LP/src/routes/tresci/`, `projekty/`, `proces-projektowy/` |
| Change markdown rendering | `secco_renderer/src/markdown.ts` |
| Change image variants | `secco_LP/src/lib/components/content/ResponsiveImage.svelte` |
| Change content types/interfaces | `secco_renderer/src/types.ts` |
| Change email format | `secco_LP/backend/email.go` |
| CMS editor behavior | `secco_cms/src/lib/components/Editor.svelte` |
| CMS deploy logic | `secco_cms/backend/deploy.go` |
| R2 bucket config | `secco_cms/backend/images.go` |
| GH Action build steps | `secco_LP/.github/workflows/build-push.yml` |
| Content snapshot format | `secco_cms/backend/deploy.go` (export) + GH Action (import) |
