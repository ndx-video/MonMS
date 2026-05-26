# Media & Publishable Assets

MonMS content publish moves **text and JSON only** between staging and production. Blobs do not copy across environments (D-55, MED-02).

Full policy: [../specs/staging.md](../specs/staging.md) §6

## What promotes where

| Storage | Layer | Promotion |
|---------|-------|-----------|
| `assets/` in this Git repo | L2 Structure | Deploys with workspace **Git tag** |
| Public CDN URLs in text/HTML fields | L3 Content | URL string upserted via **Publish to live** |
| PocketBase `file` fields → `.pb_data/storage/` | L3 runtime | **Environment-local** — not synced |

## Publishable media = CDN URLs

For assets clients reference in editorial content (hero images, PDFs, logos):

1. Store the **full public CDN URL** in a text or HTML field (e.g. `image_url`, rich text with `<img src="https://cdn.example.com/...">`).
2. Staging and production both reference the **same CDN object** — only the URL string travels in the content rail.
3. Do **not** use PocketBase file upload fields for cross-environment publishable media unless storage is backed by shared S3 (advanced).

## Export behavior

`monms content export` and the Publish console skip PocketBase **file-type** columns and log a warning. Exportable editorial fields are text, number, bool, select, and similar scalar types.

If a collection needs a publishable image, add a text field for the CDN URL in schema JSON (structure rail) before clients paste URLs (content rail).

## Consultant setup (once per site)

1. Provision one public bucket + CDN prefix (e.g. Cloudflare R2, S3 + CloudFront, Bunny).
2. Document the CDN base URL for clients (e.g. `https://cdn.client.com/uploads/`).
3. Prefer text URL fields over PocketBase-local file storage for anything that must appear on production after publish.

Site CSS and fonts in `assets/` remain on the structure rail — they deploy with Git tags, not the Publish button.

## Related docs

| Doc | Topic |
|-----|-------|
| [README.md](README.md) | Four layers, dual rails, staging vs production |
| [EDITING-GUIDE.md](EDITING-GUIDE.md) | Inline editing and Publish to live |
| [../specs/staging.md](../specs/staging.md) | Authoritative media policy |
