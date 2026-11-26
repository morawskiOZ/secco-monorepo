// Build-time switches, set from environment variables in vite.config.ts.
//
// SECCO_CONTENT=1  builds the content sections (Projekty, Treści, Proces),
//                  the navigation bar and the CMS admin page. Without it the
//                  site is the landing page and forms only.
// SECCO_DRAFTS=1   also builds entries marked `draft: true` (stage2 only).

declare const __SECCO_CONTENT__: boolean;
declare const __SECCO_DRAFTS__: boolean;

export const CONTENT_ENABLED: boolean = __SECCO_CONTENT__;
export const SHOW_DRAFTS: boolean = __SECCO_DRAFTS__;
