# Changelog

All notable changes to this project will be documented in this file.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)

## [Unreleased]

### Added
- Content sections behind a build switch (`SECCO_CONTENT=1`, off on production): Projekty (portfolio with photo viewer), Treści (articles with topic pages), Proces projektowy, and a navigation bar. Content lives as Markdown in `content/`; plan and runbook in `docs/content-and-cms.md`
- CMS at `/admin/` (Sveltia CMS, git-based) with Polish field labels; saving commits to the repository
- "Sign In with GitHub" for the CMS through the `secco-cms-auth` Worker (`cms-auth/`, vendored sveltia-cms-auth) and a GitHub App limited to this repository
- Drafts (`draft: true`): built only where `SECCO_DRAFTS=1` is set
- Photos in `content/media` are resized at build time to WebP in up to four widths
- `stage2` environment (`secco-studio-stage2` Worker): content sections, CMS and drafts, hidden from search engines; redeployed on every content commit to `main`
- Cloudflare Workers hosting: `worker/` (TypeScript port of the Go backend — forms, Turnstile, 5/hour rate limit, Gmail SMTP), `wrangler.jsonc`, and the "Deploy LP to Cloudflare Workers" workflow. Plan and runbook in `docs/hosting-cloudflare-workers.md`. The Go backend stays until cutover is complete.
- Staging Worker for testing forms end to end: mails a tester address (secret) with a `[TEST]` subject prefix instead of the studio inbox
- Tester allowlist (`TEST_EMAILS` secret): a form submitted with a listed email address is mailed back to that address with a `[TEST]` subject instead of the studio inbox, so the production form can be tested end to end
- One-year cache header for hashed files under `/_app/immutable/`

### Changed
- The Worker no longer runs for files under `/_app/`, `/media/`, `/images/` and `/fonts/`: those requests do not count against the Workers Free daily limit, and missing files there return 404
- `/polityka-prywatnosci` ships no JavaScript
- `sitemap.xml` is generated at build time (identical output while no content is published)
- Deploy workflow no longer downloads a content snapshot from R2

### Removed
- `secco_cms` (Go + SQLite admin app), `secco_renderer` (shared npm package) and the R2 snapshot pipeline, replaced by Markdown in git + Sveltia CMS. Never deployed; the code remains in git history

### Changed
- Lighthouse 100 work (plan and measurements in `docs/lighthouse-100.md`): stylesheets are inlined, the landing page ships no JavaScript, the hero image is served in three sizes, and the form's welcome image reserves its space so the text beside it no longer jumps
- Landing page and form pages have a `<main>` landmark
- Missing files under `/.well-known/` return 404 instead of the app shell
- Footer keyword line is a darker gray (`#5e5e5f`) to meet WCAG AA contrast
- Links in the privacy policy are underlined

### Added
- `llms.txt`

## [1.2.0] - 2026-04-15

### Added
- Preference survey form at `/form/preferences` — multi-question style survey for existing clients to share their design preferences before a project starts
- Backend handler `POST /api/submit-preferences` with Turnstile verification, rate limiting, and email notification
- Dedicated email template for preference survey submissions
- `sitemap.xml` with main page (priority 1.0) and privacy policy (priority 0.3)

### Changed
- `FormEngine` is now reusable — accepts optional `config` and `submitUrl` props, defaulting to the valuation form config

### Fixed
- `robots.txt` now declares `Sitemap:` directive pointing to `sitemap.xml`

## [1.1.2] - 2026-03-21

### Fixed
- Privacy policy link on form's last step now opens in a new tab so form state is not lost

## [1.1.1] - 2026-03-21

### Fixed
- Keydown events no longer trigger unintended actions (keyboard navigation/selection fixes)
- Input key event handling corrected for text fields in the form

## [1.1.0] - 2026-03-20

### Changed
- Footer redesign — removed phone, email, and Issuu link; Instagram moved to left with SVG icon (40px); keywords centered; border-top removed so footer blends into page background
- Blue decorative dot shrinks and shifts off-screen below 500px to avoid obscuring title

### Added
- "Inne" (Other) option on Q1 ("Jak trafiłeś na tą ankietę?") with free-text input, matching Q10 UX
- Uploaded file names listed in email body under "Przesłane pliki:" section

### Fixed
- Keyboard shortcuts (A/B/C…) no longer trigger option selection while typing in "Inne" text input (Q1 and Q10)

## [1.0.0] - 2026-03-19

### Added
- Landing page with hero section, studio description, and CTA
- 14-question valuation form (Typeform-style, one question per screen)
- Privacy policy page (GDPR Art. 13 compliant)
- Go backend with embedded static files
- SMTP email delivery with attachments
- Cloudflare Turnstile bot protection
- Rate limiting (5 submissions/hour/IP)
- Self-hosted Open Sans fonts (woff2)
- WebP hero images with JPG fallback
- SEO meta tags, OG tags, JSON-LD Organization schema
- Docker multi-stage build (scratch base, ~15MB image)
- GitHub Actions CI/CD (build + push to GHCR)
- MicroK8s deployment manifests
