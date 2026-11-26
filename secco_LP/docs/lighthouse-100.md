# Lighthouse 100 — plan and runbook

Brings every public page of www.seccostudio.com to a full Lighthouse score
without changing what the site does or how it looks. Status is at the bottom.

## Goal

100 in every Lighthouse category on `/`, `/form`, `/form/preferences` and
`/polityka-prywatnosci`, mobile and desktop, with the same behaviour and the
same pixels as before — except the items listed under "Visible changes", which
need the owner's approval on staging.

## What ran before (production, measured 2026-11-26, Lighthouse 13.6.0)

| Page | Device | Perf | A11y | Best pr. | SEO | Agentic |
|---|---|---|---|---|---|---|
| `/` | mobile | 99 | 92 | 100 | 100 | 50 |
| `/` | desktop | 100 | 92 | 100 | 100 | 50 |
| `/form`, `/form/preferences` | mobile | 100 | 97 | 96 | 63 | 50 |
| `/form`, `/form/preferences` | desktop | 91 | 98 | 100 | 66 | 41 |
| `/polityka-prywatnosci` | mobile | 99 | 94 | 100 | 100 | 50 |
| `/polityka-prywatnosci` | desktop | 99 | 94 | 100 | 100 | 50 |

"Agentic" is the agentic-browsing category Lighthouse 13 added (llms.txt, ARD
manifest, accessibility tree, layout stability).

## Causes and fixes

| Failing audit | Cause | Fix | Visible? |
|---|---|---|---|
| LCP 2.1 s on `/` mobile; render-blocking CSS | Two CSS requests before first paint; 34 kB of hydration JS competing with the hero image on a page with nothing interactive; 800 px image sent to phones | Stylesheets inlined (`inlineStyleThreshold`); `csr = false` on `/`; hero `srcset` with 480/640/800 px files | No |
| CLS 0.186 on the form pages (desktop) | Welcome image had `width: auto` and wrong `width`/`height` attributes (500×1202 for an 800×1202 file), so it took no space until loaded | Attributes corrected; CSS width set to the value `auto` resolves to | No |
| Image aspect ratio (form, mobile) | Same wrong attributes | Same | No |
| No main landmark (`/`, forms) | No `<main>` | `<main>` around the hero; form container is a `<main>` | No |
| llms.txt invalid | No file: the SPA fallback answered with HTML | `static/llms.txt` | No |
| ARD manifest invalid | No manifest: `/.well-known/ard.json` got the SPA shell with 200 | Missing files under `/.well-known/` return 404 (audit then does not apply) | No |
| Colour contrast (`/`) | Footer keyword line `#7F7F80` on `#EBD5C9` = 2.8:1 | `#5e5e5f` = 4.6:1, the lightest gray that passes | **Yes** |
| Links rely on colour (privacy policy) | Navy links inside gray text, 2.0:1 against it, no underline | Links underlined | **Yes** |
| Page blocked from indexing (forms) | `<meta name="robots" content="noindex">`, set on purpose | **Not changed** — see open decision | — |

### Options rejected

- **Prerendering the form pages** (server-rendered welcome screen): would make
  the form's first paint faster, but it changes how a 1,500-line interactive
  flow starts up (hydration, Turnstile) for no score gain. The pages reach 100
  on performance without it.
- **A real ARD manifest**: the site offers no agent-usable resources (no API,
  no MCP server), so a manifest would describe nothing. A 404 is the honest
  answer and the audit accepts it.
- **Dropping the font preloads** to speed up the hero image: would make text
  flash from the fallback font on slow connections.
- **Darkening the shared `--color-accent-secondary` token**: it also colours
  the gray circle in the hero. Only the footer text colour changed.

## Differences from before

Behaviour:

- `/` ships no JavaScript. The "Wypełnij formularz wyceny" button is a normal
  link, so the form opens with a page load instead of a client-side transition.
- Missing files under `/.well-known/` return 404 instead of the SPA shell.
  Every other unknown path still returns the shell with 200.
- `/llms.txt` exists.
- Phones and standard-density desktops download a smaller hero file
  (`hero-480.webp` 18 kB or `hero-640.webp` 26 kB instead of 53 kB). Layout and
  crop are identical; the picture is resampled from a smaller source that still
  has at least as many pixels as the screen shows.

Visible changes (need approval, each is one commit and reverts on its own):

1. Footer keyword line on `/`: gray `#7F7F80` → `#5e5e5f`.
2. Privacy policy: the three contact links are underlined.

Proof of "nothing else moved": screenshots of production and of the new build
at 390, 768, 1350 and 1920 px wide (plus five extra sizes for the form) were
compared pixel by pixel. The form pages are identical. `/` differs only inside
the hero photo and in the footer line; the privacy policy only at the links.

## Open decision (owner)

The form pages keep `noindex`, so their SEO score stays at 63 (mobile) / 66
(desktop): Lighthouse fails "page is blocked from indexing" and nothing else.
Reaching 100 means letting search engines index them.

- `/form` is the public valuation form linked from the home page. Indexing it
  is low-risk and may bring visitors straight to the form.
- `/form/preferences` is meant for existing clients who get the link from the
  studio. Indexing it would expose it in search results.

Recommendation: index `/form` (and add it to `sitemap.xml`), keep
`/form/preferences` hidden and accept 63/66 there.

## Not fixable from this repo

Production injects Cloudflare Web Analytics (`beacon.min.js`, a zone setting).
Lighthouse lists it under "Legacy JavaScript" and "Use efficient cache
lifetimes". These are unscored insights and do not lower any category score;
the only way to clear them is to turn Web Analytics off in the Cloudflare
dashboard. Staging (`workers.dev`) does not have the beacon.

## Runbook

```bash
cd secco_LP
npm ci && npm test && npm run check:worker && npm run check
```

1. Deploy the branch to staging: Actions → "Deploy LP to Cloudflare Workers" →
   branch `perf/lighthouse-100`, environment `staging`.
2. Verify staging: routes and headers match the saved baseline, Lighthouse on
   all four pages (mobile and desktop), one form submission end to end (staging
   mails the tester address, never the studio).
3. Owner looks at staging and approves the two visible changes and decides the
   `noindex` question.
4. Merge the PR, then run the same workflow with environment `production`.
5. Verify production: same checks as step 2 (form test through the tester
   allowlist, see `hosting-cloudflare-workers.md`).

**Rollback:** `npx wrangler rollback --env=""` returns production to the
previous deployment in one step (or re-run the workflow from the previous
`lp-v*` tag). No DNS, secrets or data are touched by this change.

## Status

- [x] Baseline saved (scores, headers, HTML, redirects) 2026-11-26
- [x] Fixes implemented; 33 Worker tests, `svelte-check` clean
- [x] Local build: 100 in every category on `/` and the privacy policy; forms
      100 except SEO (noindex)
- [ ] Staging deployed and verified
- [ ] Owner approval of visible changes; `noindex` decision
- [ ] Production deploy and verification
