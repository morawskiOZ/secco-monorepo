# Changelog

All notable changes to this project will be documented in this file.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)

## [Unreleased]

### Planned
- Navbar component with mobile hamburger menu
- Blog section (Treści) — listing and article pages
- Portfolio section (Projekty) — grid and project detail pages
- Design process page (Proces projektowy)
- SQLite content storage in Go backend
- Content management API (public read + protected CRUD)
- Cloudflare Images integration for image upload and optimization
- SEO meta tag injection for dynamic content pages
- Separate CMS app (secco_cms) with markdown editor and preview
- Responsive image delivery via Cloudflare flexible variants
- Image lightbox for project galleries
- Content preview system for draft content

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
