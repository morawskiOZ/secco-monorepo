# Secco Studio — Project Root

Website of Secco Studio (interior design, Wrocław): landing page, two forms,
and — behind a build switch — a portfolio, articles and a design-process page
with a git-based CMS.

```
secco/
├── secco_LP/                    # the whole site (SvelteKit static build + Cloudflare Worker)
│   ├── content/                 # projects, articles, pages, photos (Markdown + images)
│   ├── cms/                     # Sveltia CMS admin page and config (served at /admin/)
│   ├── src/                     # pages and components
│   ├── worker/                  # forms API, redirects (Cloudflare Worker)
│   └── docs/                    # plans and runbooks — read before changing hosting or content
└── .github/workflows/deploy-lp.yml
```

## Environments

| Environment | URL | What it is |
|---|---|---|
| production | www.seccostudio.com | Landing page and forms. Content sections off |
| staging | secco-studio-staging.secco-studio.workers.dev | Same build as production, for testing releases and forms |
| stage2 | secco-studio-stage2.secco-studio.workers.dev | Content sections, CMS and drafts. Redeployed on every content commit to `main` |

All three are Cloudflare Workers on the Free plan, deployed by the
"Deploy LP to Cloudflare Workers" workflow. Production is only deployed by hand.

## Key decisions

| Decision | Choice | Reason |
|---|---|---|
| Content at runtime | None — every page is prerendered | No database, Lighthouse 100, free hosting |
| Content storage | Markdown files in `secco_LP/content` | History and rollback from git; no server |
| Editor | Sveltia CMS (static page, commits through the GitHub API) | No backend or login system of our own |
| Photos | Originals in git, resized to WebP at build time | No quota, no extra Cloudflare product |
| Launch control | `SECCO_CONTENT` / `SECCO_DRAFTS` build switches | Collect content on stage2 before production shows any of it |

Details: `secco_LP/docs/content-and-cms.md` (content and CMS),
`secco_LP/docs/hosting-cloudflare-workers.md` (hosting, forms, secrets),
`secco_LP/docs/lighthouse-100.md`.

## Design tokens

```
Primary background: #EBD5C9 (warm beige)
Accent:            #2E3192 (deep navy)
Secondary accent:  #7F7F80 (medium gray)
Page background:   #fafafa (off-white)
Text:              #454545 (charcoal)
Navbar height:     64px desktop / 56px mobile
Fonts:             "Open Sans", sans-serif
```

All UI text is Polish.
