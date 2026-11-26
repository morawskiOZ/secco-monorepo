# CMS sign-in Worker (`secco-cms-auth`)

Makes the **Sign In with GitHub** button work in the CMS at `/admin/`.
The code in `src/index.js` is [sveltia-cms-auth](https://github.com/sveltia/sveltia-cms-auth),
vendored unchanged at a fixed commit. It stores nothing: it redirects the
editor to GitHub, swaps the code GitHub returns for a token, and passes the
token to the CMS page that opened it (only if that page is on a host listed in
`ALLOWED_DOMAINS`).

URL: `https://secco-cms-auth.secco-studio.workers.dev` (Workers Free plan; a
sign-in is two requests).

## One-time setup (owner)

1. **Register a GitHub App** — open this pre-filled form:
   <https://github.com/settings/apps/new?name=Secco+Studio+CMS&url=https://www.seccostudio.com&webhook_active=false&contents=write&pull_requests=write>
   and set:
   - **Callback URL**: `https://secco-cms-auth.secco-studio.workers.dev/callback`
   - Expire user authorization tokens: checked
   - Request user authorization (OAuth) during installation: unchecked
   - Webhook → Active: unchecked
   - Repository permissions: Contents *Read and write*, Pull requests *Read and write*, nothing else
   - Where can this GitHub App be installed: *Only on this account*

   Create it, then copy the **Client ID** and generate a **client secret**.
2. **Install the app** on the account, *Only select repositories* →
   `secco-monorepo`.
3. **Store the two values as repository secrets** (each command prompts for the value):
   ```bash
   gh secret set CMS_AUTH_CLIENT_ID -R morawskiOZ/secco-monorepo
   gh secret set CMS_AUTH_CLIENT_SECRET -R morawskiOZ/secco-monorepo
   ```
4. **Deploy**: Actions → "Deploy CMS sign-in Worker" → Run workflow. It deploys
   the Worker and copies the two secrets to it.

## Adding an editor

The editor creates a GitHub account (GitHub's own sign-in page offers
"Continue with Google") and is added to the repository as a collaborator with
write access. On first sign-in they approve the app once. Tokens issued through
the app can only read and write file contents and pull requests of this
repository — not its workflows, and nothing in the editor's other repositories.

Removing access: remove the collaborator. Turning sign-in off for everyone:
delete the client secret in the GitHub App settings.

## Changing things

- Another site host (e.g. a new staging URL): add it to `ALLOWED_DOMAINS` in
  `wrangler.jsonc` and run the workflow.
- New client secret: `gh secret set CMS_AUTH_CLIENT_SECRET`, run the workflow.
- The CMS finds this Worker through `backend.base_url` in `secco_LP/cms/config.yml`.
