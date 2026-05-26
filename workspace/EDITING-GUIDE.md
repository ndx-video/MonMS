# Inline Editing Guide

This walkthrough verifies Phase 3 inline contextual editing for MonMS operators.

## Prerequisites

- MonMS server running with a initialized workspace
- `hero_content` collection present in `workspace/schema/hero_content.json`
- Homepage record seeded with title **Welcome to MonMS**

## 1. Sign in via Admin

1. Open `/_/` in your browser (PocketBase admin dashboard).
2. Sign in with your superuser credentials.
3. Navigate to `/` — the page auto-syncs your admin session and should show **Live Editor Active**.
4. If the editor badge is missing after admin login, hard-refresh `/` once.

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
3. Confirm there is **no** Live Editor badge and **no** `contenteditable` attributes in the page source (ICE-06).

Admin logout clears PocketBase's `pocketbase_auth` localStorage entry. When you next load `/`, MonMS detects the mismatch and clears its HttpOnly `monms_auth` cookie automatically. If the homepage was already open in another tab, it reloads as a guest. **Sign out** on the Live Editor badge clears both sessions at once.

## 5. Troubleshooting

| Symptom | Fix |
|---------|-----|
| Hero text replaced by JSON after blur | Ensure editable fields have `hx-swap="none"` — PocketBase returns JSON on PATCH; without it HTMX swaps the response into the element |
| 401 or 403 on blur-save | Re-authenticate at `/_/` — the Bearer token in the page may have expired. In **development** builds, a red banner at the bottom shows the HTTP status and error message |
| Badge missing after admin login | Hard-refresh `/` — the page syncs from PocketBase admin localStorage (`__pb_superusers__/*`) automatically |
| Hero shows fallback copy | Check server logs for seed warnings; ensure `hero_content` schema imported |

## Existing workspaces (manual merge)

If your workspace was created before Phase 3, `monms init` skips existing scaffold files. Manually merge or copy from a fresh init:

- `templates/layouts/base.gohtml` — editor badge + HTMX auth script
- `templates/index.gohtml` — hero binding + conditional inline edit
- `assets/main.css` — pulse dot and editable focus styles
- `schema/hero_content.json` — collection definition

See `internal/scaffold/embed/` in the MonMS repository for canonical copies.

## 6. Publish to live (publishers only)

Editors with the **publisher** role can push approved staging content to production. Editors without that role can inline-edit on staging but cannot publish (PUB-07).

### Prerequisites

1. Consultant configured staging once:
   - Copy `.monms/config.example.json` → `.monms/config.json`
   - Set `productionUrl` to the production MonMS instance
   - Add publisher emails to `publisherEmails`
   - Set matching `MONMS_PUBLISH_TOKEN` on staging and production
2. Sign in at `/_/` with an email on the publisher allowlist.

### Publish workflow

1. Edit content on **staging** via inline editing (sections 1–3 above).
2. Open **Publish to live**:
   - Click the link in the **Live Editor Active** badge (publishers only), or
   - Navigate directly to `/api/monms/publish`
3. Review the **diff preview** — changed collections and fields since last publish.
4. Click **Publish now** — staging exports editorial records and POSTs to production.
5. Confirm **Last published** timestamp updates; production site reflects new copy without a structure deploy.

### Roles

| Role | Inline edit (staging) | Publish to live |
|------|----------------------|-----------------|
| Editor | ✓ | ✗ |
| Publisher | ✓ | ✓ |

Publishers are configured in `workspace/.monms/config.json` — not in PocketBase collection rules. See [README.md](README.md) for environment setup and [MEDIA.md](MEDIA.md) for CDN URL guidance.
