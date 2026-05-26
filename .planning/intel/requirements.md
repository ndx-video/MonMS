# Synthesized Requirements — ingest specs/staging.md

## Environment & Lifecycle (v2)

- **ENV-01**: Documentation and tooling distinguish four layers (engine, structure, content, audience).
- **ENV-02**: Structure promotion uses workspace Git tags; content promotion is a separate rail.
- **ENV-03**: Staging and production are separate MonMS instances with separate `.pb_data/` directories.

## Content Publish (v2)

- **PUB-01**: Editorial collections marked `"editorial": true` in schema JSON.
- **PUB-02**: `workspace/content/*.json` holds exported editorial records with stable IDs.
- **PUB-03**: `monms content export` writes editorial snapshots to `workspace/content/`.
- **PUB-04**: `monms content import` upserts records idempotently by ID.
- **PUB-05**: Production exposes `POST /api/monms/content/import` with scoped publish token.
- **PUB-06**: Staging admin UI includes **Publish to live** with diff preview.
- **PUB-07**: Publisher role gates the publish action; editors may edit without publishing.
- **PUB-08**: Staging tracks last-published checksum for unpublished-changes indicator.
- **PUB-09**: `monms content diff` shows pending changes before publish.

## Media (v2)

- **MED-01**: Publishable media uses public CDN URLs stored in content fields; blobs are not copied between environments.
- **MED-02**: Documentation warns against PocketBase-local file storage for publishable assets.
