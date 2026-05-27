# Publish to live

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

Editors with the **publisher** role can push approved staging content to production. Editors without that role can inline-edit on staging but cannot publish.

Complete inline editing steps first: [Inline editing](inline-editing.md).

## Prerequisites

1. Consultant configured staging once:
   - Copy `.monms/config.example.json` → `.monms/config.json`
   - Set `productionUrl` to the production MonMS instance
   - Add publisher emails to `publisherEmails`
   - Set matching `MONMS_PUBLISH_TOKEN` on staging and production
2. Sign in at `/_/` with an email on the publisher allowlist.

## Publish workflow

1. Edit content on **staging** via inline editing.
2. Open **Publish to live**:
   - Click the link in the **Live Editor Active** badge (publishers only), or
   - Navigate directly to `/_monms/publish`
3. Review the **diff preview** — changed collections and fields since last publish.
4. Click **Publish now** — staging exports editorial records and POSTs to production.
5. Confirm **Last published** timestamp updates; production site reflects new copy without a structure deploy.

## Publish limitations

- **Upsert only:** Publish sends the current editorial snapshot to production and upserts records by ID. Records you delete on staging remain on production until removed manually in production admin (`/_/`).
- **Diff includes deletions:** The pending-changes diff reports records removed from staging, but publish does not propagate those deletions.
- **Partial import:** If a publish or CLI import fails mid-batch, production may already contain some updated collections. Imports are idempotent — retry after fixing the error and verify with diff or the publish console.

## Roles

| Role | Inline edit (staging) | Publish to live |
|------|----------------------|-----------------|
| Editor | Yes | No |
| Publisher | Yes | Yes |

Publishers are configured in `site/.monms/config.json` — not in PocketBase collection rules.

For media: [Media URLs](media-urls.md). For operator setup: [Getting started](../operators/getting-started.md).
