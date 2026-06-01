# Inline editing

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)  
> **PocketBase admin:** [official docs](https://pocketbase.io/docs/)

Walkthrough for MonMS **inline contextual editing** on the **staging** instance — preparing editorial copy before it goes to production.

*Staging* here means the client content-editing environment, not MonMS engine development or consultant **shaping** of templates/schema. See [Getting started](../operators/getting-started.md).

## Prerequisites

- MonMS server running with an initialized site
- `hero_content` collection present in `site/schema/hero_content.json`
- Homepage record seeded with title **Welcome to MonMS**

## 1. Sign in via Admin

1. Open `/_/` in your browser (PocketBase admin dashboard).
2. Sign in with your superuser credentials.
3. Navigate to `/` — the page auto-syncs your admin session and should show **Live Editor Active**.
4. If the editor badge is missing after admin login, hard-refresh `/` once.

Use **MonMS Console** and **View site** in the PocketBase admin header (MonMS UI extension) to open the operator dashboard or public homepage.

## 2. Confirm Live Editor badge

1. Navigate to `/` (the public homepage).
2. Verify the fixed badge in the top-right reads **Live Editor Active**.
3. Confirm the **Full Admin Dashboard** link points to `/_/`.

## 3. Edit hero content inline

1. Click the hero headline (**Welcome to MonMS**) and change the text.
2. Click away (blur) — the title saves via HTMX PATCH to `hero_content/homepage`.
3. Repeat for the hero paragraph; blur to save the body field.
4. Refresh the page — your edits should persist.

## 4. Verify logged-out safety

1. Log out from the admin dashboard at `/_/` (or use **Sign out** on the Live Editor badge).
2. Reload `/`.
3. Confirm there is **no** Live Editor badge and **no** `contenteditable` attributes in the page source.

Admin logout clears PocketBase's `pocketbase_auth` localStorage entry. When you next load `/`, MonMS detects the mismatch and clears its HttpOnly `monms_auth` cookie automatically. **Sign out** on the Live Editor badge clears both sessions at once.

## 5. Troubleshooting

| Symptom | Fix |
|---------|-----|
| Hero text replaced by JSON after blur | Ensure editable fields have `hx-swap="none"` — PocketBase returns JSON on PATCH; without it HTMX swaps the response into the element |
| 401 or 403 on blur-save | Re-authenticate at `/_/` — the Bearer token in the page may have expired. In **development** builds, a red banner at the bottom shows the HTTP status and error message |
| Badge missing after admin login | Hard-refresh `/` — the page syncs from PocketBase admin localStorage automatically |
| Hero shows fallback copy | Check server logs for seed warnings; ensure `hero_content` schema imported |

## Existing sites (manual merge)

If your site was created before inline editing shipped, `monms init` skips existing scaffold files. Manually merge or copy from a fresh init:

- `templates/layouts/base.gohtml` — editor badge + HTMX auth script
- `templates/index.gohtml` — hero binding + conditional inline edit
- `assets/main.css` — pulse dot and editable focus styles
- `schema/hero_content.json` — collection definition

See `internal/scaffold/embed/` in the MonMS engine repository for canonical copies.

## Next step

Publishers: [Publish to live](publish-to-live.md)
