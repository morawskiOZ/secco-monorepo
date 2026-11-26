# Secco Studio - Website & Valuation Form

## Project Overview

Complete website for Secco Studio (interior design studio in Wroclaw, Poland). Single app with landing page, valuation and preference forms, and — behind a build switch — a portfolio, articles and a design-process page.

**Domain:** www.seccostudio.com
**Routes:** `/` (landing page), `/form` (valuation form), `/form/preferences` (preference survey), `/polityka-prywatnosci` (privacy policy); with `SECCO_CONTENT=1` also `/projekty`, `/tresci`, `/proces-projektowy`, `/admin/`
**Business purpose:** Lead generation — form submissions are emailed to the studio for follow-up with project offers.

## Architecture

**Frontend:** SvelteKit (static adapter)
- Landing page (`/`) and privacy policy prerendered to static HTML with no JavaScript — perfect Lighthouse
- Form page (`/form`) client-side interactive — Typeform-style one-question-at-a-time flow
- Questions defined in a config file (`src/lib/config/questions.ts`) for easy modification
- Plain CSS with design tokens — no UI framework, no Tailwind

**Backend:** Cloudflare Worker (`worker/`, TypeScript) on the Workers Free plan
- Static build served by Workers Static Assets; the Worker handles `/api/*`, the apex → www redirect and the SPA fallback
- `POST /api/submit`, `POST /api/submit-preferences` — multipart form data, Turnstile check, 5/hour per IP (Durable Object), email through Gmail SMTP
- Reply-To header set to customer's email for easy reply
- No database, no file storage (files in-memory during request only)
- `backend/` (Go) and the Dockerfile are the previous hosting, kept until the old cluster is decommissioned

**Bot protection:** Cloudflare Turnstile (form page only)
- Client-side: explicit render mode, token stored in Svelte state
- Server-side: the Worker verifies the token against Cloudflare siteverify API before processing

**Legal:** GDPR Art. 6(1)(b) — pre-contractual measures (no consent checkbox)
- Short info clause near submit + link to `/polityka-prywatnosci` privacy policy page

**Key constraints:**
- Perfect Lighthouse score (100/100 on all categories)
- Minimal server resources (~10MB RAM)
- No persistent storage

## Design Tokens (from current seccostudio.com)

```
Primary background: #EBD5C9 (warm beige)
Accent:            #2E3192 (deep navy blue)
Secondary accent:  #7F7F80 (medium gray)
Page background:   #fafafa (off-white)
Text:              #454545 (charcoal gray)
Fonts:             "Open Sans", "Roboto", "Montserrat", sans-serif
```

## Landing Page Content (from current site)

- **Title:** "secco.studio"
- **Subtitle:** "Studio projektowe zlokalizowane we Wrocławiu"
- **Description:** "Zajmujemy się kompleksową aranżacją wnętrz domów i mieszkań. Projektujemy, doradzamy, nadzorujemy i staramy się sprawnie przeprowadzić Państwa przez cały proces projektowy."
- **Phone:** +48 665 895 432
- **Email:** secco.studio@gmail.com
- **Social:** Instagram (@secco.studio), Issuu portfolio
- **SEO keywords:** architektura, wnętrza, architekt wnętrz, wrocław, legnica, projekt, kuchnia, projektowanie wnętrz, wizualizacje
- **Hero image:** interior design bathroom photo
- **Decorative elements:** navy circle (top-right), gray circle (bottom-center)

## Form Questions (14 questions)

All question text is in Polish and lives in `src/lib/config/questions.ts`. (`task.md` is the original 2026 build plan; its CMS phases are superseded by `docs/content-and-cms.md`.)

## Tech Stack

- Frontend: SvelteKit 2, Svelte 5, TypeScript, plain CSS
- Static adapter: prerender landing page, SPA mode for form
- Backend: Cloudflare Worker (TypeScript), Durable Object for rate limiting, `cloudflare:sockets` for SMTP
- Build: `scripts/prebuild.ts` (photos, CMS admin) → SvelteKit builds to `build/`
- Deploy: GitHub Actions "Deploy LP to Cloudflare Workers" → `wrangler deploy` (production, staging, stage2)
- Tests: vitest (`test/` content logic, `worker/test/` forms)

## Content sections (behind `SECCO_CONTENT=1`)

Projekty, Treści and Proces projektowy are prerendered from Markdown in
`content/`; `/admin/` is Sveltia CMS. Off on production, on for stage2. Read
`docs/content-and-cms.md` before changing any of it. The fields in
`cms/config.yml` must match `src/lib/content/model.ts`.

## Development Commands

```bash
npm install
npm run dev              # landing page + forms on :5173
SECCO_CONTENT=1 SECCO_DRAFTS=1 npm run dev    # with content sections and drafts
npm run build            # production build (content off)
npm run build:stage2     # stage2 build (content, drafts, noindex)
npm run dev:worker       # built site + Worker locally (wrangler dev)
npm test                 # unit tests (content logic + Worker)
npm run check && npm run check:worker
```

Hosting is Cloudflare Workers (`docs/hosting-cloudflare-workers.md`). The Go
backend and Dockerfile are the previous hosting, kept until the cluster is
decommissioned.

## Environment Variables

```
# SMTP (Gmail)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=secco.studio@gmail.com
SMTP_PASS=<gmail-app-password>
RECIPIENT_EMAIL=secco.studio@gmail.com

# Cloudflare Turnstile
PUBLIC_TURNSTILE_SITE_KEY=<public-sitekey>  # build-time only (baked into frontend)
TURNSTILE_SECRET_KEY=<secret-key>

# Build switches (see src/lib/flags.ts)
SECCO_CONTENT=1   # content sections, navigation bar, /admin/   (stage2)
SECCO_DRAFTS=1    # also build entries marked draft: true        (stage2)
SECCO_NOINDEX=1   # X-Robots-Tag: noindex on everything          (stage2)
```

On Cloudflare, `SMTP_PASS`, `TURNSTILE_SECRET_KEY` and (staging/stage2)
`RECIPIENT_EMAIL` are Worker secrets; the rest are `vars` in `wrangler.jsonc`.

## Changelog

Always update `CHANGELOG.md` when making user-facing changes. Follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format. Add entries under `## [Unreleased]` or a new version section as appropriate.

## Lighthouse Targets

All pages must score 100 on:
- Performance (prerendered HTML, optimized images, minimal JS)
- Accessibility (semantic HTML, ARIA labels, contrast ratios)
- Best Practices (HTTPS, no console errors, correct image aspect ratios)
- SEO (meta tags, structured data, canonical URLs, mobile viewport)
