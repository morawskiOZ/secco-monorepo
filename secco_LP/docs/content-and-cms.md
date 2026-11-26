# Content sections and CMS — plan and runbook

Adds a portfolio (Projekty), articles (Treści) and a design-process page to
the landing page, with an editor for the studio, on the Cloudflare Workers
Free plan. Built and running on **stage2**; production is unchanged until the
switch described under "Going live". Status is at the bottom.

## Goal

- The studio can add projects, articles and photos without a developer.
- Everything stays static, free to host and at Lighthouse 100.
- Content can be collected now, reviewed on stage2 and published later
  without redoing it.
- Production keeps serving exactly what it serves today until we decide.

## What existed before (assessment, 2026-11-26)

| Piece | State | Verdict |
|---|---|---|
| Content pages in `secco_LP` (navbar, `/tresci`, `/projekty`, `/proces-projektowy`) | Ticked off in `task.md`, but not in the repository: only three empty JSON placeholders arrived with the mono-repo import | Rebuilt here |
| `secco_cms` — SvelteKit admin + Go API + SQLite, Docker image for Kubernetes | Code complete, never deployed. Needs a server with a disk; the cluster it targeted is being switched off (burn-in ends 2026-11-28). Its Go tests did not run from a clean checkout (`embed` of a missing `build/`) | Replaced |
| `secco_renderer` — npm package shared by site and CMS preview, shipped as a tarball | Only needed because two apps rendered the same Markdown | Replaced |
| Pipeline: CMS → JSON snapshot in R2 → GitHub Action → build | Three systems and five secrets to publish one article; R2 and Image Transformations were never set up | Replaced |
| `docs/cms-setup.md`, `content-api.md`, `cloudflare-images.md` | Describe an even older design (runtime content API, Cloudflare Images) | Removed |

## Options considered

| Option | Cost | Verdict |
|---|---|---|
| **Markdown files in this repository + Sveltia CMS (git-based editor) + photos resized at build** | Free. No server, database, bucket or login system of our own | **Chosen** |
| Port `secco_cms` to a Worker with D1 (database) and R2 (photos) | Free tier would cover it (D1: 5 GB, 100k writes/day; R2: 10 GB) | Rejected: ~2,500 lines of Go to rewrite, our own password login to keep secure, R2 may ask for a card, and backups/rollback to build. Git gives history and rollback for nothing |
| Keep `secco_cms` on a small VM or Cloud Run | From free to ~€5/month | Rejected: a server to patch and a SQLite file to back up, for one editor |
| Decap CMS instead of Sveltia | Free | Same file format; Sveltia has the better editor, Polish UI and built-in photo shrinking. Because the config is Decap-compatible, switching later is a one-line change |
| Hosted headless CMS (Sanity, Contentful, …) | Free tiers with limits | Rejected: content leaves the repository, another account and vendor |
| Cloudflare Image Transformations for photo sizes | Free up to 5,000 new sizes/month, then errors | Rejected: build-time resizing has no quota and needs no zone setting |

Limits checked on 2026-11-26 (Cloudflare docs): Workers Free allows 100,000
Worker requests/day and 10 ms CPU per request; requests for static files are
free and unlimited *unless* the Worker is set to run first for them; a
deployment may hold 20,000 files of up to 25 MiB each.

## Design

```
secco_LP/content/            ← the content, in git
  projects/<slug>.md           YAML front matter + Markdown body
  articles/<slug>.md
  pages/proces-projektowy.md
  media/<photo>                originals (≤ 2560 px)

editor ── /admin/ (Sveltia CMS, static page) ── GitHub API ── commit on main
                                                                  │
                         GitHub Actions "Deploy LP" (push to content/) ──► stage2
                         same workflow, run by hand                     ──► production

build:  scripts/prebuild.ts   photos → WebP in 480/800/1200/1920 px + manifest
        vite build            every page prerendered to HTML
        scripts/postbuild.ts  stage2 only: X-Robots-Tag noindex
```

- **Build switches** (`src/lib/flags.ts`): `SECCO_CONTENT=1` builds the content
  sections, the navigation bar and `/admin/`; `SECCO_DRAFTS=1` also builds
  entries marked `draft: true`. Stage2 sets both; production sets neither.
- **Drafts** are how content is collected ahead of launch: a draft is visible
  on stage2 and nowhere else. Unticking "Szkic" publishes it wherever the
  content sections are on.
- **No JavaScript on content pages** (`csr = false`). The mobile menu is a
  `<details>` element; the only script is the 1.7 kB photo viewer inlined on
  project pages, and without it photo links open the image.
- **Navigation shows a section once it has something to show**, so sections can
  go live one at a time.
- **Photos**: the CMS shrinks uploads to 2560 px WebP in the browser before
  committing; the build writes the smaller sizes and `width`/`height`, so pages
  do not shift while loading. A content file that names a missing photo fails
  the build instead of shipping a broken image.
- **Pure logic is separate from files and platform** (`src/lib/content/`:
  `frontmatter`, `model`, `markdown`, `images`, `sitemap`) and unit-tested.
- **Worker-first routing narrowed**: `/_app/*`, `/media/*`, `/images/*` and
  `/fonts/*` are served without invoking the Worker, so a page view costs one
  Worker request however many photos it shows.

### Editor sign-in

Sveltia CMS talks to GitHub directly from the browser. Two ways in, no server
of ours either way:

1. **Access token** (works today): "Sign in with token" with a fine-grained
   GitHub token limited to this repository, *Contents: read and write*.
2. **"Sign in with GitHub" button** (optional, nicer for a non-technical
   editor): needs a GitHub OAuth app plus the free
   [sveltia-cms-auth](https://github.com/sveltia/sveltia-cms-auth) Worker, then
   `base_url` in `cms/config.yml`.

The editor needs a GitHub account with write access to this repository.
Developers can also use "Work with local repository" (Chromium browsers) or
edit the Markdown files directly.

### Differences from today

For production these apply from its next deploy, even with the content
sections off:

- `/polityka-prywatnosci` ships no JavaScript (it had nothing interactive).
- `sitemap.xml` is generated at build time; byte-identical while there is no
  published content (covered by a test).
- Missing files under `/_app/`, `/media/`, `/images/`, `/fonts/` return 404
  instead of the app shell, and requests for existing files there on the apex
  host are served directly instead of redirected to www.
- Landing page HTML: same markup and CSS rules; the footer's CSS is emitted as
  its own block and there is one extra empty HTML comment.
- The deploy workflow no longer downloads a content snapshot from R2.

Only with the content sections on (stage2 now, production after going live):
the navigation bar on the landing page and privacy policy, the new pages, and
`/admin/`.

### Known risks

| Risk | Fallback |
|---|---|
| Sveltia CMS is pre-1.0 with one maintainer | Version is pinned in `cms/index.html`. Content is plain Markdown: Decap CMS reads the same config, and files can always be edited by hand |
| The repository is public, so drafts and unpublished photos are readable on GitHub | Accepted for now (nothing confidential planned). If that changes: make the repository private (Actions minutes are still free at this volume) |
| Repository grows with photos | ~0.5–1 MB per photo after shrinking; 300 photos ≈ 150–300 MB, inside GitHub's 1 GB guidance. Beyond that: move originals to R2 |
| Build time grows with photos | Each photo is resized on every CI run (~0.1 s per size). Add an Actions cache for `static/media` if a run passes ~3 minutes |
| A saved entry with a mistake (e.g. no date) fails the build | Stage2 keeps serving the previous version; the failed run names the file and field |

## Free-plan check (measured 2026-11-26)

| Limit (Workers Free) | Use |
|---|---|
| 100,000 Worker requests/day | 1 per page view (HTML only) + form posts. Photos, scripts and fonts do not count |
| 10 ms CPU per request | Content pages are static files; the Worker only passes them through. Form uploads are unchanged (see `hosting-cloudflare-workers.md`) |
| 20,000 files per deployment | Stage2 today: see Status. Each photo adds up to 4 files → room for ~4,900 photos |
| 25 MiB per file | Largest generated photo is well under 1 MB |
| 100 Workers per account | 3 (production, staging, stage2) |
| GitHub Actions | Public repository: free |

No paid Cloudflare product is used: no R2, D1, Images or KV.

## Runbook

```bash
cd secco_LP
npm ci && npm test && npm run check && npm run check:worker
npm run build:stage2 && npx wrangler dev --env stage2     # local preview
```

1. **Deploy stage2**: Actions → "Deploy LP to Cloudflare Workers" →
   environment `stage2` (any branch). After the PR is merged, every commit
   under `secco_LP/content/` on `main` deploys stage2 by itself.
2. **Edit content**: open `/admin/` on stage2, sign in, add or change entries.
   Saving commits to `main`; stage2 shows the change after the workflow run
   (about a minute).
3. **Forms on stage2** (optional): they answer 500 until the mail secrets are
   set, and then only ever mail the tester:
   `npx wrangler secret put RECIPIENT_EMAIL --env stage2` and
   `npx wrangler secret put SMTP_PASS --env stage2`.

### Going live (not done — needs the owner's go-ahead)

1. Studio replaces the sample entries with real ones and unticks "Szkic" on
   what should be public; check them on stage2.
2. In `.github/workflows/deploy-lp.yml`, "Build site" step: export
   `SECCO_CONTENT=1` for `production` as well (not `SECCO_DRAFTS`).
3. Deploy `staging`, compare with stage2, run Lighthouse on the new pages.
4. Deploy `production`. Add `https://www.seccostudio.com/admin/` as the
   editor's bookmark.

**Rollback:** `npx wrangler rollback --env=""` (or `--env stage2`) returns to
the previous deployment in one step. A bad content change is undone with
`git revert` of its commit. Nothing here touches DNS, secrets or data.

## Status

- [x] Assessment, options and plan (this document) 2026-11-26
- [x] Content model, Markdown rendering, photo sizes, sitemap: unit-tested
- [x] Pages, navigation bar, photo viewer; checked locally at desktop and phone width
- [x] Build with content off compared with production HTML (differences listed above)
- [x] CMS configuration (`cms/config.yml`), Polish labels
- [x] Sample content marked as drafts; content guide in `content/README.md`
