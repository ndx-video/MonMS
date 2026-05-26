# Inline Editing Guide

This walkthrough verifies Phase 3 inline contextual editing for MonMS operators.

## Prerequisites

- MonMS server running with a initialized workspace
- `hero_content` collection present in `workspace/schema/hero_content.json`
- Homepage record seeded with title **Welcome to MonMS**

## 1. Sign in via Admin

1. Open `/_/` in your browser (PocketBase admin dashboard).
2. Sign in with your superuser credentials.

## 2. Confirm Live Editor badge

1. Navigate to `/` (the public homepage).
2. Verify the fixed badge in the top-right reads **Live Editor Active**.
3. Confirm the **Full Admin Dashboard** link points to `/_/`.

## 3. Edit hero content inline

1. Click the hero headline (**Welcome to MonMS**) and change the text.
2. Click away (blur) — the title saves via HTMX PUT to `hero_content/homepage`.
3. Repeat for the hero paragraph; blur to save the body field.
4. Refresh the page — your edits should persist.

## 4. Verify logged-out safety

1. Log out from the admin dashboard at `/_/`.
2. Reload `/`.
3. Confirm there is **no** Live Editor badge and **no** `contenteditable` attributes in the page source (ICE-06).

## 5. Troubleshooting

| Symptom | Fix |
|---------|-----|
| 401 or 403 on blur-save | Re-authenticate at `/_/` — the Bearer token in the page may have expired |
| Badge missing after login | Hard-refresh `/` after signing in at `/_/` |
| Hero shows fallback copy | Check server logs for seed warnings; ensure `hero_content` schema imported |

## Existing workspaces (manual merge)

If your workspace was created before Phase 3, `monms init` skips existing scaffold files. Manually merge or copy from a fresh init:

- `templates/layouts/base.gohtml` — editor badge + HTMX auth script
- `templates/index.gohtml` — hero binding + conditional inline edit
- `assets/main.css` — pulse dot and editable focus styles
- `schema/hero_content.json` — collection definition

See `internal/scaffold/embed/` in the MonMS repository for canonical copies.
