# Secco Studio — Website Expansion & CMS

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                     USERS (Browser)                      │
│                                                          │
│   seccostudio.com              cms.seccostudio.com       │
│   (Public Website)             (Admin CMS)               │
└──────────┬───────────────────────────┬───────────────────┘
           │                           │
           ▼                           ▼
┌────────────────────┐      ┌────────────────────┐
│  secco_forms       │      │  secco_cms         │
│  Go Backend        │◄─────│  Go Backend        │
│  + SvelteKit       │ API  │  + SvelteKit       │
│  + SQLite          │      │  (no DB)           │
│                    │      │                    │
│  - Static pages    │      │  - CF Images API   │
│  - Form handler    │      │  - Content proxy   │
│  - Content API     │      │  - Markdown editor │
│  - Email (SMTP)    │      │  - Auth (JWT)      │
│  - Turnstile       │      │                    │
│  - SEO injection   │      │                    │
└──────┬─────────────┘      └──────┬─────────────┘
       │                           │
       ▼                           ▼
  ┌──────────┐            ┌────────────────┐
  │  SQLite  │            │  Cloudflare    │
  │  DB file │            │  Images API    │
  └──────────┘            └────────────────┘
```

### System Components

| Component | Role | Tech |
|-----------|------|------|
| secco_forms (website) | Public website + content API + form handler | SvelteKit + Go + SQLite |
| secco_cms | Admin panel for content management | SvelteKit + Go |
| Cloudflare Images | Image storage, optimization, CDN delivery | Cloudflare API |
| SQLite | Content storage (articles, projects, process) | modernc.org/sqlite (pure Go) |

### New Routes

| Route | Type | Description |
|-------|------|-------------|
| `/` | Prerendered | Landing page (existing) |
| `/form` | SPA | Valuation form (existing) |
| `/polityka-prywatnosci` | Prerendered | Privacy policy (existing) |
| `/tresci` | SPA | Blog listing page |
| `/tresci/[slug]` | SPA | Individual blog article |
| `/projekty` | SPA | Projects portfolio grid |
| `/projekty/[slug]` | SPA | Individual project detail |
| `/proces-projektowy` | SPA | Design process page |

### New API Endpoints

**Public (read):**
```
GET /api/content/tresci                       → list published articles
GET /api/content/tresci/:slug                 → single article
GET /api/content/projekty                     → list published projects
GET /api/content/projekty/:slug               → single project
GET /api/content/proces                       → design process content
GET /api/content/preview/:type/:slug?token=X  → preview draft content
```

**Protected (CMS — X-API-Key header):**
```
GET    /api/admin/content              → list all content (drafts included)
GET    /api/admin/content/:id          → get by ID
POST   /api/admin/content              → create
PUT    /api/admin/content/:id          → update
DELETE /api/admin/content/:id          → delete
PUT    /api/admin/content/:id/publish  → publish
PUT    /api/admin/content/:id/draft    → unpublish
POST   /api/admin/images/upload        → upload to Cloudflare Images
GET    /api/admin/images               → list all images
DELETE /api/admin/images/:id           → delete image
```

---

## Resolved Decisions (Original)

| Question | Decision |
|----------|----------|
| File uploads | Single file UI, backend accepts `[]File`. |
| SMTP provider | Gmail SMTP with app password. |
| "Other" on Q10 | Selecting "Other" reveals a free-text input. |
| Required fields | Only first name + email required. |
| GDPR | Art. 6(1)(b) — no consent checkbox. Info clause + privacy policy link. |
| Bot protection | Cloudflare Turnstile (form page only). |

## New Decisions

| Question | Decision |
|----------|----------|
| Content storage | SQLite via `modernc.org/sqlite` (pure Go, no CGo). Single DB file. |
| CMS architecture | Separate app, calls website content API. No own database. |
| Image hosting | Cloudflare Images with flexible variants. Auto WebP/AVIF. |
| Image quality | 90-95% for portfolio/gallery, 85% for thumbnails. Originals preserved. |
| SEO for content | Go injects meta tags + OG data into HTML shell before serving. |
| CMS auth | Session-based login (env var credentials), JWT for API calls. |
| Markdown rendering | Client-side via `marked` library. Sanitized with `DOMPurify`. |
| Content preview | Iframe in CMS pointing to website preview endpoint with auth token. |

---

## Database Schema

```sql
CREATE TABLE content (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    type        TEXT NOT NULL CHECK(type IN ('article', 'project', 'process')),
    slug        TEXT NOT NULL UNIQUE,
    title       TEXT NOT NULL,
    summary     TEXT,
    body        TEXT NOT NULL,
    cover_image TEXT,
    status      TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft', 'published')),
    sort_order  INTEGER DEFAULT 0,
    published_at TIMESTAMP,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE images (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    cloudflare_id  TEXT NOT NULL UNIQUE,
    filename       TEXT NOT NULL,
    alt_text       TEXT,
    width          INTEGER,
    height         INTEGER,
    size_bytes     INTEGER,
    content_type   TEXT,
    delivery_url   TEXT NOT NULL,
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

## Implementation Phases

### Phase 1: Navbar & Layout Update
**Goal:** Persistent navigation across all pages. Mobile-responsive hamburger menu.

- [ ] Create `src/lib/components/Navbar.svelte`
  - Logo "secco.studio" (left) → links to `/`
  - Nav links (right): Projekty, Treści, Proces, Formularz wyceny
  - "Formularz wyceny" styled as CTA button (navy bg, white text)
  - Active link indicator (bottom border, navy color)
  - Sticky positioning, warm beige background (#EBD5C9)
  - Subtle box-shadow on scroll (via IntersectionObserver or scroll event)
- [ ] Create `src/lib/components/MobileMenu.svelte`
  - Hamburger icon (3-line, CSS-only)
  - Full-height slide-down overlay
  - Links centered, larger font
  - Close on link click or overlay tap
  - Body scroll lock when open
- [ ] Update `src/routes/+layout.svelte`
  - Include Navbar above `<slot />`
  - Adjust page padding-top for sticky navbar height
- [ ] Responsive breakpoint: 768px (hamburger below, full links above)
- [ ] Test navbar on all existing pages (landing, form, privacy)
- [ ] Ensure Lighthouse scores remain 100 on landing page

### Phase 2: Go Backend — SQLite & Content API
**Goal:** Content storage and API in the Go backend.

- [ ] Add `modernc.org/sqlite` dependency to `backend/go.mod`
- [ ] Create `backend/database.go`
  - `InitDB(path string) (*sql.DB, error)` — open SQLite, run migrations
  - Schema creation (content, images, content_images tables)
  - Auto-migration on startup
- [ ] Create `backend/content.go` — content CRUD handlers
  - `GET /api/content/tresci` — list published articles, ordered by published_at DESC
  - `GET /api/content/tresci/:slug` — single article by slug
  - `GET /api/content/projekty` — list published projects, ordered by sort_order ASC
  - `GET /api/content/projekty/:slug` — single project by slug
  - `GET /api/content/proces` — latest published process entry
  - `GET /api/content/preview/:type/:slug` — draft preview (requires preview token)
  - JSON responses with proper content-type headers
- [ ] Create `backend/admin.go` — protected CRUD endpoints
  - API key middleware (`X-API-Key` header check)
  - `GET /api/admin/content` — list all (with drafts), filter by type
  - `POST /api/admin/content` — create (validate slug uniqueness, type)
  - `PUT /api/admin/content/:id` — update
  - `DELETE /api/admin/content/:id` — delete
  - `PUT /api/admin/content/:id/publish` — set status=published, published_at=now
  - `PUT /api/admin/content/:id/draft` — set status=draft
- [ ] Create `backend/images.go` — Cloudflare Images integration
  - `POST /api/admin/images/upload` — receive file, upload to CF, store metadata
  - `GET /api/admin/images` — list all images with delivery URLs
  - `DELETE /api/admin/images/:id` — delete from CF + DB
  - Cloudflare API: `POST /client/v4/accounts/{id}/images/v1`
  - Return delivery URL with account hash for frontend use
- [ ] Create `backend/seo.go` — HTML meta tag injection
  - Intercept requests to `/tresci/*`, `/projekty/*`, `/proces-projektowy`
  - Read content from DB by slug
  - Inject `<title>`, `<meta name="description">`, OG tags into HTML before serving
  - Fall back to default meta tags if content not found
- [ ] Update `backend/main.go`
  - Initialize SQLite on startup
  - Register new routes
  - Add CONTENT_API_KEY, PREVIEW_TOKEN, DATABASE_PATH env vars
  - Add CLOUDFLARE_ACCOUNT_ID, CLOUDFLARE_IMAGES_API_TOKEN, CLOUDFLARE_ACCOUNT_HASH env vars
  - Create `data/` directory if not exists
- [ ] Update `.env.example` with new env vars
- [ ] Add `data/` to `.gitignore`

### Phase 3: Website Content Pages — Treści (Blog)
**Goal:** Blog listing and article pages with markdown rendering.

- [ ] Install `marked` and `dompurify` npm dependencies
- [ ] Create `src/lib/utils/markdown.ts`
  - Configure `marked` renderer for custom image handling (CF Images responsive)
  - Wrap images in `<picture>` tags with srcset for different variants
  - Add `loading="lazy"` and `decoding="async"` to all images
  - Sanitize output with DOMPurify
- [ ] Create `src/lib/components/content/ContentRenderer.svelte`
  - Takes markdown string, renders to sanitized HTML
  - Applies typography styles (article-body class)
  - Responsive images with CF Images variants
- [ ] Create `src/lib/components/content/ArticleCard.svelte`
  - Cover image (thumbnail variant), title, summary, date
  - Hover effect: subtle scale + shadow
  - Link to `/tresci/[slug]`
- [ ] Create `src/routes/tresci/+page.svelte`
  - Fetch articles from `/api/content/tresci`
  - 3-column grid (desktop), 2-column (tablet), 1-column (mobile)
  - Empty state: "Brak artykułów" message
  - `ssr: false` in `+page.ts`
- [ ] Create `src/routes/tresci/+page.ts` — `export const ssr = false`
- [ ] Create `src/routes/tresci/[slug]/+page.svelte`
  - Fetch article from `/api/content/tresci/{slug}`
  - Hero cover image (full width, hero variant)
  - Title (h1), date, markdown content via ContentRenderer
  - Max-width container (~720px) for reading comfort
  - Back link to `/tresci`
  - 404 handling if article not found
- [ ] Create `src/routes/tresci/[slug]/+page.ts` — `export const ssr = false`
- [ ] Typography styles for article body (headings, paragraphs, lists, blockquotes, code)

### Phase 4: Website Content Pages — Projekty (Projects)
**Goal:** Portfolio grid and project detail pages. Image-heavy layout.

- [ ] Create `src/lib/components/content/ProjectCard.svelte`
  - Large cover image (large variant), title, short summary
  - Aspect ratio: 4:3 or 3:2 (landscape, interior design photos)
  - Hover: slight zoom (transform: scale 1.02) + overlay with title
  - Link to `/projekty/[slug]`
- [ ] Create `src/lib/components/content/ImageGallery.svelte`
  - Responsive grid for project images (2-3 columns)
  - Click image → lightbox (full-size view)
  - Lazy loading for below-fold images
  - Gap: 8px between tiles
- [ ] Create `src/lib/components/content/Lightbox.svelte`
  - Full-screen overlay with image
  - Close button, click-outside-to-close
  - Previous/next navigation (arrows + keyboard)
  - Image counter (3/12)
  - Uses `full` CF Images variant for maximum quality
  - Swipe support on mobile
- [ ] Create `src/routes/projekty/+page.svelte`
  - Fetch projects from `/api/content/projekty`
  - 2-column grid (desktop), 1-column (mobile)
  - Larger cards than blog (portfolio emphasis)
  - Empty state message
- [ ] Create `src/routes/projekty/+page.ts` — `export const ssr = false`
- [ ] Create `src/routes/projekty/[slug]/+page.svelte`
  - Fetch project from `/api/content/projekty/{slug}`
  - Title (h1), initial description
  - Markdown content with interspersed images and text
  - Image gallery sections rendered via ContentRenderer
  - Back link to `/projekty`
- [ ] Create `src/routes/projekty/[slug]/+page.ts` — `export const ssr = false`
- [ ] Ensure project images use high-quality CF variants (quality=90+)

### Phase 5: Website Content Pages — Proces Projektowy
**Goal:** Single page describing the design process.

- [ ] Create `src/routes/proces-projektowy/+page.svelte`
  - Fetch content from `/api/content/proces`
  - Title "Proces projektowy" (h1)
  - Markdown content rendered via ContentRenderer
  - Sections with images (typical: numbered steps of the design process)
  - Same max-width container as articles
  - Empty state: placeholder text until content is added via CMS
- [ ] Create `src/routes/proces-projektowy/+page.ts` — `export const ssr = false`
- [ ] Style process sections (numbered headings, full-width images between sections)

### Phase 6: Cloudflare Images Integration
**Goal:** Image upload, optimization, and responsive delivery.

- [ ] Create `src/lib/components/content/ResponsiveImage.svelte`
  - Takes CF delivery URL base, alt text, optional aspect ratio
  - Renders `<picture>` with srcset:
    - `(max-width: 480px)` → `w=480,quality=90`
    - `(max-width: 768px)` → `w=768,quality=90`
    - `(max-width: 1200px)` → `w=1200,quality=90`
    - Default → `w=1920,quality=95`
  - `loading="lazy"`, `decoding="async"`
  - Explicit width/height for CLS prevention
  - Fade-in on load
- [ ] Update markdown renderer to use ResponsiveImage for all images
  - Detect CF Images URLs in markdown image tags
  - Replace with responsive `<picture>` elements
- [ ] Configure CF Images flexible variants (in Cloudflare dashboard):
  - Enable flexible variants for the account
  - Set allowed transformations: w, h, fit, quality, format
- [ ] Go backend image upload handler (`backend/images.go`):
  - Accept multipart file upload
  - Validate: JPEG, PNG, WebP only; max 20MB
  - Upload to CF Images API
  - Extract and store: CF ID, delivery URL, dimensions, file size
  - Return image metadata JSON
- [ ] Image deletion: remove from CF + DB

### Phase 7: CMS App — Scaffolding (secco_cms)
**Goal:** Initialize the CMS app at `/home/pm/code/secco_cms`.

- [ ] Initialize SvelteKit project with adapter-static
- [ ] Initialize Go backend in `secco_cms/backend/`
- [ ] Project structure:
  ```
  secco_cms/
  ├── src/
  │   ├── routes/
  │   │   ├── +layout.svelte        # Admin layout (sidebar + content area)
  │   │   ├── +page.svelte          # Dashboard
  │   │   ├── login/+page.svelte    # Login page
  │   │   ├── tresci/
  │   │   │   ├── +page.svelte      # Articles list
  │   │   │   ├── new/+page.svelte  # New article editor
  │   │   │   └── [id]/+page.svelte # Edit article
  │   │   ├── projekty/
  │   │   │   ├── +page.svelte      # Projects list
  │   │   │   ├── new/+page.svelte  # New project editor
  │   │   │   └── [id]/+page.svelte # Edit project
  │   │   ├── proces/+page.svelte   # Edit process page
  │   │   └── images/+page.svelte   # Image manager
  │   ├── lib/
  │   │   ├── components/
  │   │   │   ├── Sidebar.svelte
  │   │   │   ├── Editor.svelte
  │   │   │   ├── Preview.svelte
  │   │   │   ├── ImageUpload.svelte
  │   │   │   └── ImagePicker.svelte
  │   │   ├── api/
  │   │   │   ├── content.ts
  │   │   │   └── images.ts
  │   │   └── stores/
  │   │       └── auth.ts
  │   └── app.css
  ├── backend/
  │   ├── main.go
  │   ├── auth.go
  │   ├── proxy.go
  │   ├── cloudflare.go
  │   └── go.mod
  ├── svelte.config.js
  ├── vite.config.ts
  ├── package.json
  ├── Dockerfile
  └── CLAUDE.md
  ```
- [ ] Copy design tokens from secco_forms (consistent styling)
- [ ] Set up Go backend with auth + proxy to secco_forms API
- [ ] Vite proxy: `/api/*` → Go backend on :8081

### Phase 8: CMS — Authentication
**Goal:** Simple, secure admin login.

- [ ] Go backend `auth.go`:
  - `POST /api/auth/login` — verify username/password, return JWT
  - `POST /api/auth/logout` — invalidate session
  - `GET /api/auth/check` — verify JWT validity
  - JWT with 24h expiry, HS256 signing
  - Credentials from env vars (CMS_ADMIN_USER, CMS_ADMIN_PASS bcrypt hash)
- [ ] Login page (`login/+page.svelte`):
  - Username + password form
  - Error message on invalid credentials
  - Redirect to dashboard on success
  - Clean, minimal design
- [ ] Auth store (`stores/auth.ts`):
  - JWT stored in httpOnly cookie (set by Go backend)
  - Auth state: `$state({ authenticated: boolean, checking: boolean })`
  - Check auth on app load
- [ ] Route guard in `+layout.svelte`:
  - If not authenticated → redirect to `/login`
  - Show loading spinner while checking auth

### Phase 9: CMS — Content Editor
**Goal:** Full-featured markdown editor with image upload and live preview.

- [ ] Create `Editor.svelte`:
  - Split layout: editor (left 50%) + preview (right 50%)
  - Markdown textarea with monospace font
  - Toolbar: H1, H2, H3, Bold, Italic, Link, Image, List, Quote, Code
  - Toolbar actions insert markdown syntax at cursor position
  - Line numbers (optional)
  - Auto-save every 5 seconds (debounced PUT to API)
  - Unsaved changes indicator
  - Tab key inserts spaces (not focus change)
- [ ] Create `Preview.svelte`:
  - Renders markdown to HTML using same renderer as website
  - Uses same CSS/typography as website content pages
  - Syncs scroll position with editor (approximate)
  - Shows "Preview" badge in corner
- [ ] Content form fields (above editor):
  - Title input (auto-generates slug)
  - Slug input (editable, validates uniqueness)
  - Summary textarea (2 rows)
  - Cover image picker (opens image manager modal)
  - Sort order (number input, for projects only)
  - Status badge (Draft / Published)
  - Publish / Save Draft buttons
- [ ] Create article editor page (`tresci/new/+page.svelte`, `tresci/[id]/+page.svelte`):
  - New: empty form, POST on save
  - Edit: load existing content, PUT on save
  - Delete button with confirmation modal
- [ ] Create project editor page (same pattern as articles)
- [ ] Create process editor page (`proces/+page.svelte`):
  - Single entry — auto-loads existing or creates new
  - Same editor, no slug/summary fields
- [ ] Image upload in editor:
  - "Insert Image" toolbar button → opens image picker modal
  - Drag-drop images directly into editor textarea
  - Uploads to CF Images via CMS backend
  - Inserts `![alt](delivery-url/large)` at cursor

### Phase 10: CMS — Image Manager
**Goal:** Upload, browse, and manage images across all content.

- [ ] Create `images/+page.svelte`:
  - Grid view of all uploaded images (thumbnail variant)
  - Upload zone (drag-drop + file picker, multi-file)
  - Image details panel (click to select):
    - Preview (medium variant)
    - Filename, dimensions, size, upload date
    - Alt text (editable, saved to DB)
    - Copy URL buttons (thumbnail, medium, large, full)
    - Delete button with confirmation
  - Search/filter by filename
  - Pagination or infinite scroll
- [ ] Create `ImageUpload.svelte` (reusable):
  - Drag-drop zone with visual feedback
  - File type validation (JPEG, PNG, WebP)
  - Size validation (max 20MB)
  - Upload progress indicator (percentage bar)
  - Multi-file support
  - Calls CMS backend → CF Images API
- [ ] Create `ImagePicker.svelte` (modal for editor):
  - Shows image grid (same as image manager)
  - Click to select → returns image URL
  - Upload new image inline
  - Alt text input before inserting
  - "Insert" button closes modal and inserts markdown

### Phase 11: CMS — Content Preview
**Goal:** Preview content exactly as it appears on the website.

- [ ] Website preview endpoint:
  - `GET /api/content/preview/:type/:slug?token=PREVIEW_TOKEN`
  - Returns draft content JSON (same format as public endpoints)
  - Token validated server-side
- [ ] Website preview page:
  - `/preview/[type]/[slug]?token=X` — SvelteKit page that renders content
  - Uses same ContentRenderer, same styles
  - Yellow "Preview" banner at top: "Viewing draft — not published"
- [ ] CMS preview button:
  - "Preview" button in editor toolbar
  - Saves current content first (auto-save)
  - Opens preview URL in new tab
  - URL: `https://seccostudio.com/preview/article/my-slug?token=PREVIEW_TOKEN`
- [ ] Inline preview in CMS:
  - Right panel shows rendered preview (using same CSS)
  - Switches between "Split" (editor + preview) and "Preview" (full width) modes

### Phase 12: CMS — Dashboard & Lists
**Goal:** Dashboard overview and content listing pages.

- [ ] Create `Sidebar.svelte`:
  - Logo: "secco.cms" or "Secco CMS"
  - Nav links: Dashboard, Treści, Projekty, Proces, Zdjęcia
  - Active link highlight
  - Collapse on mobile (hamburger)
  - Logout button at bottom
- [ ] Create dashboard (`+page.svelte`):
  - Summary cards: X articles, Y projects, Z images
  - Quick actions: "Nowy artykuł", "Nowy projekt", "Edytuj proces"
  - Recent content list (last 5 changes)
- [ ] Create content list pages (`tresci/+page.svelte`, `projekty/+page.svelte`):
  - Table: Title, Status, Date, Actions
  - Status badges: "Opublikowany" (green), "Szkic" (gray)
  - Actions: Edit, Preview, Publish/Unpublish, Delete
  - "Nowy artykuł" / "Nowy projekt" button
  - Reorder projects via drag-drop (sort_order)

### Phase 13: Deployment Updates
**Goal:** Update Docker, K8s, and CI/CD for new architecture.

- [ ] Update `Dockerfile` for secco_forms:
  - Add SQLite support (no CGo needed with modernc.org/sqlite)
  - Add `data/` volume mount point for SQLite file
  - Add new env vars to runtime
- [ ] Create `Dockerfile` for secco_cms:
  - Same multi-stage pattern (Node → Go → scratch)
  - Expose port 8081
- [ ] Update `docker-compose.yml`:
  - Add secco_cms service
  - Shared network for CMS → website API calls
  - Volume for SQLite data persistence
  - New env vars
- [ ] Update K8s manifests:
  - secco_forms: add PersistentVolumeClaim for SQLite data
  - secco_cms: new Deployment, Service, Ingress (cms.seccostudio.com)
  - Updated Secrets (new env vars)
  - NetworkPolicy: CMS can reach website API (internal)
- [ ] Update GitHub Actions:
  - Build both images
  - Tag both with same version
- [ ] Update `Taskfile.yml` with new commands

### Phase 14: Polish & Testing
**Goal:** Ensure everything works end-to-end, responsive, accessible.

- [ ] Responsive testing: all new pages on mobile/tablet/desktop
- [ ] Lighthouse audit: all new pages
- [ ] Accessibility check: keyboard navigation, screen reader, ARIA
- [ ] Cross-browser: Chrome, Firefox, Safari
- [ ] End-to-end test flows:
  - Create article in CMS → appears on website
  - Upload image → correct variants served
  - Preview draft → shows unpublished content
  - Edit and publish → live update
  - Delete content → removed from website
- [ ] Error handling: API failures, image upload failures, network errors
- [ ] Loading states: skeletons or spinners for content fetches
- [ ] Empty states: meaningful messages when no content exists
- [ ] 404 pages for non-existent slugs

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
- Background: `#EBD5C9` (warm beige, matching page)
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
- Cover image: full-width (hero variant, max-height 480px, object-fit cover)
- Content container: max-width 720px, centered
- Title: h1, margin-bottom 8px
- Date: small gray text below title
- Body: rendered markdown with good typography
- Images in body: full container width, responsive
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
- Image gallery: responsive grid (2-3 columns)
- Interspersed text blocks between image sections
- Click image → lightbox with full-quality view
- Back link: "← Wróć do projektów"

### Proces Projektowy — `/proces-projektowy`
- Title: "Proces projektowy" (h1)
- Content container: max-width 720px, centered
- Sections with headings, descriptions, and images
- Similar layout to article page

---

## Content Types

### Article (Treści)
```json
{
  "id": 1,
  "type": "article",
  "slug": "jak-wybrac-styl-wnetrzarski",
  "title": "Jak wybrać styl wnętrzarski?",
  "summary": "Krótki przewodnik po najpopularniejszych stylach...",
  "body": "# Jak wybrać styl...\n\nMarkdown content...",
  "cover_image": "https://imagedelivery.net/hash/cf-id/large",
  "status": "published",
  "sort_order": 0,
  "published_at": "2026-03-19T10:00:00Z",
  "created_at": "2026-03-18T15:00:00Z",
  "updated_at": "2026-03-19T09:30:00Z"
}
```

### Project (Projekty)
```json
{
  "id": 2,
  "type": "project",
  "slug": "mieszkanie-na-krakowskiej",
  "title": "Mieszkanie na Krakowskiej",
  "summary": "Nowoczesne wnętrze w centrum Wrocławia",
  "body": "## O projekcie\n\nOpis...\n\n![salon](cf-url/large)\n![kuchnia](cf-url/large)\n\n## Detale\n\nWięcej tekstu...\n\n![łazienka](cf-url/large)",
  "cover_image": "https://imagedelivery.net/hash/cf-id/large",
  "status": "published",
  "sort_order": 1,
  "published_at": "2026-03-15T12:00:00Z"
}
```

### Process (Proces projektowy)
```json
{
  "id": 3,
  "type": "process",
  "slug": "proces-projektowy",
  "title": "Proces projektowy",
  "summary": null,
  "body": "## 1. Konsultacja\n\n![konsultacja](cf-url/large)\n\nOpis...\n\n## 2. Projekt koncepcyjny\n\n...",
  "status": "published"
}
```

---

## Cloudflare Images — Delivery URL Pattern

Base URL: `https://imagedelivery.net/{account_hash}/{image_id}`

**Flexible variant examples:**
```
/w=400,h=300,fit=cover,quality=85      → thumbnail (listing cards)
/w=800,quality=90                       → medium (in-content images)
/w=1200,quality=90                      → large (project galleries)
/w=1920,quality=95                      → hero (cover images)
/w=2400,quality=95                      → full (lightbox)
/w=1200,h=630,fit=cover,quality=85      → og (social sharing)
```

Auto-format: Cloudflare serves WebP or AVIF based on browser `Accept` header. No extra config needed.

---

## Environment Variables

### secco_forms (updated)
```env
# Existing
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=secco.studio@gmail.com
SMTP_PASS=<gmail-app-password>
RECIPIENT_EMAIL=secco.studio@gmail.com
PUBLIC_TURNSTILE_SITE_KEY=<sitekey>
TURNSTILE_SECRET_KEY=<secret>
PORT=8080

# New — Content & Images
DATABASE_PATH=./data/secco.db
CONTENT_API_KEY=<random-secret-for-cms-access>
PREVIEW_TOKEN=<random-secret-for-draft-preview>
CLOUDFLARE_ACCOUNT_ID=<cf-account-id>
CLOUDFLARE_IMAGES_API_TOKEN=<cf-api-token>
CLOUDFLARE_ACCOUNT_HASH=<cf-account-hash>
```

### secco_cms (new)
```env
CMS_ADMIN_USER=admin
CMS_ADMIN_PASS_HASH=<bcrypt-hash>
JWT_SECRET=<random-secret>
WEBSITE_API_URL=http://localhost:8080
WEBSITE_API_KEY=<same-as-CONTENT_API_KEY>
WEBSITE_PUBLIC_URL=https://seccostudio.com
PREVIEW_TOKEN=<same-as-PREVIEW_TOKEN>
CLOUDFLARE_ACCOUNT_ID=<cf-account-id>
CLOUDFLARE_IMAGES_API_TOKEN=<cf-api-token>
CLOUDFLARE_ACCOUNT_HASH=<cf-account-hash>
PORT=8081
```

---

## File Structure (Expanded secco_forms)

```
secco_forms/
├── src/
│   ├── routes/
│   │   ├── +layout.svelte                      # Updated: includes Navbar
│   │   ├── +page.svelte                         # Landing page (existing)
│   │   ├── +page.ts
│   │   ├── form/                                # Existing form
│   │   ├── polityka-prywatnosci/                # Existing privacy policy
│   │   ├── tresci/
│   │   │   ├── +page.svelte                     # Blog listing
│   │   │   ├── +page.ts
│   │   │   └── [slug]/
│   │   │       ├── +page.svelte                 # Blog article
│   │   │       └── +page.ts
│   │   ├── projekty/
│   │   │   ├── +page.svelte                     # Projects grid
│   │   │   ├── +page.ts
│   │   │   └── [slug]/
│   │   │       ├── +page.svelte                 # Project detail
│   │   │       └── +page.ts
│   │   ├── proces-projektowy/
│   │   │   ├── +page.svelte                     # Design process
│   │   │   └── +page.ts
│   │   └── preview/
│   │       └── [type]/[slug]/
│   │           ├── +page.svelte                 # Draft preview
│   │           └── +page.ts
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
│   │   │       ├── ImageGallery.svelte
│   │   │       ├── Lightbox.svelte
│   │   │       └── ResponsiveImage.svelte
│   │   ├── utils/
│   │   │   └── markdown.ts                      # NEW
│   │   ├── config/
│   │   │   └── questions.ts                     # Existing
│   │   └── styles/
│   │       └── tokens.css                       # Updated: new typography tokens
│   ├── app.css
│   ├── app.d.ts
│   └── app.html
├── static/                                      # Existing
├── backend/
│   ├── main.go                                  # Updated: new routes, DB init
│   ├── handler.go                               # Existing
│   ├── email.go                                 # Existing
│   ├── turnstile.go                             # Existing
│   ├── database.go                              # NEW: SQLite init, migrations
│   ├── content.go                               # NEW: public content endpoints
│   ├── admin.go                                 # NEW: protected CRUD endpoints
│   ├── images.go                                # NEW: CF Images integration
│   ├── seo.go                                   # NEW: HTML meta tag injection
│   └── go.mod                                   # Updated: +sqlite dependency
├── data/                                        # NEW: SQLite database (gitignored)
│   └── secco.db
├── docs/                                        # NEW
│   ├── cloudflare-images.md
│   ├── content-api.md
│   └── cms-setup.md
├── CHANGELOG.md                                 # NEW
├── task.md                                      # This file
├── CLAUDE.md
├── Dockerfile                                   # Updated
├── docker-compose.yml                           # Updated
├── Taskfile.yml                                 # Updated
└── ...
```

---

## Modification Guide (Updated)

| Change | Where to edit |
|--------|--------------|
| Add/edit form questions | `src/lib/config/questions.ts` |
| Change design tokens | `src/lib/styles/tokens.css` |
| Change navbar links | `src/lib/components/Navbar.svelte` |
| Change content page layout | `src/routes/tresci/`, `src/routes/projekty/` |
| Change markdown rendering | `src/lib/utils/markdown.ts` |
| Change image variants | `src/lib/components/content/ResponsiveImage.svelte` |
| Add content API field | `backend/database.go` (schema) + `backend/content.go` (handler) |
| Change email format | `backend/email.go` |
| Update CF Images config | `backend/images.go` + CF dashboard |
| CMS editor behavior | `secco_cms/src/lib/components/Editor.svelte` |
