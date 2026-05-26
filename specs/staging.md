# SPEC: MonMS Staging, Environments & Content Publish

**Status:** Accepted (2026-05-23) · **Implemented** (Phase 4, 2026-05-26)  
**Supersedes:** Implicit single-environment assumptions in v1 planning  
**Related:** `specs/monms-prd.md`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`

---

## 1. Summary

MonMS operates across **four layers** with **two promotion rails**. Structure (templates, schema, assets) promotes via Git tags. Editorial content (PocketBase records) promotes via JSON upsert — initiated by **clients** from the **Publish to live** console at `/api/monms/publish`, without consultant involvement for routine updates.

Media referenced by public CDN URLs does **not** move between environments; only URL strings in content records are synced.

---

## 2. Four Layers

| Layer | Name | Who | Artifact | Change frequency |
|-------|------|-----|----------|------------------|
| **L1** | Engine | MonMS developers | `monms` binary (this repo) | Rare — semver releases |
| **L2** | Site structure | Consultants, AI agents, advanced clients | `workspace/` Git repo — `templates/`, `assets/`, `schema/` | Per feature / page |
| **L3** | Content | Business clients (editors/publishers) | `workspace/.pb_data/` (runtime) + optional `workspace/content/` (export) | Daily / weekly |
| **L4** | Audience | End users | Production URL (read-only) | Continuous consumption |

**L1 is frozen at deploy.** L2–L4 run on the same binary version. v1 implemented L1–L3 on a single host; this spec defines how L2–L4 split across staging and production environments.

---

## 3. Environments

Typical deployment uses **two MonMS instances**:

| Environment | Purpose | Layers active |
|-------------|---------|---------------|
| **Staging** | Structure iteration + client content editing | L2 (branch), L3 (staging `.pb_data/`) |
| **Production** | Public audience | L2 (tagged release), L3 (production `.pb_data/`), L4 |

Both instances run the same pinned `monms` binary version unless an engine upgrade is intentional.

### 3.1 Structure promotion (L2) — consultant-driven, infrequent

1. Consultant or agent mutates `workspace/` on staging (templates, schema JSON, assets).
2. Changes committed to workspace Git with validation (`monms validate`, pre-commit hook).
3. Review → **Git tag** on workspace repo (e.g. `v1.2.0`).
4. Production deploy checks out that tag.
5. fsnotify/cache picks up template changes; bootstrap imports `schema/*.json`.

**Git is the transport for structure.** This rail is unchanged from v1 planning.

### 3.2 Content promotion (L3→L4) — client-driven, frequent

1. Client edits content on **staging** via inline HTMX editing (existing v1 feature).
2. Client opens **Publish to live** at `/api/monms/publish` (or via the editor badge link).
3. Staging exports editorial collections → JSON payload.
4. Staging POSTs to production **content import API** (scoped token).
5. Production **upserts records by fixed ID** — idempotent, no binary restart.
6. Staging records last-publish checksum for “unpublished changes” indicator.

**Consultants are not in the loop for routine content pushes.** One-time setup: production URL, publish token, editorial collection marking.

---

## 4. Two-Rail Model (decided)

```
Structure rail:  workspace Git tag  →  production checkout
Content rail:    editorial JSON upsert  →  production .pb_data/ records
Media rail:      shared public CDN URLs  →  no blob copy (URL strings only)
```

These rails are **independent**. Structure must reach production before content that depends on new collections or fields.

---

## 5. Content Sync Design

### 5.1 `workspace/content/` convention

Parallel to `workspace/schema/`:

```
workspace/
├── schema/           # L2 — collection shape (existing)
│   └── hero_content.json
├── content/          # L3 — editorial records (new)
│   └── hero_content.json
└── .pb_data/         # Runtime DB (never committed)
```

**Schema JSON** defines fields and API rules. **Content JSON** holds editorial records for export/import/publish.

Example `workspace/content/hero_content.json`:

```json
{
  "collection": "hero_content",
  "records": [
    {
      "id": "homepage",
      "title": "Welcome to Acme Corp",
      "body": "We build things."
    }
  ]
}
```

### 5.2 Editorial collection marking

Collections participating in content sync declare `"editorial": true` in their schema JSON. System collections (`_superusers`, auth tables) are always excluded.

Fixed record IDs (e.g. `homepage`) are required for reliable upsert — same pattern as v1 hero seed.

### 5.3 CLI commands (engine)

| Command | Purpose |
|---------|---------|
| `monms content export` | Snapshot editorial collections → `workspace/content/*.json` |
| `monms content import` | Upsert from `workspace/content/*.json` → local `.pb_data/` |
| `monms content diff` | Show records/fields that differ from last export or target |
| `monms content publish --to URL` | Export from running instance + POST to production (operator/CI fallback; requires `MONMS_PUBLISH_TOKEN`) |

Primary UX is the **Publish to live** console at `/api/monms/publish`; CLI remains for CI and consultant emergencies.

### 5.4 Production import API

New authenticated endpoint on production:

```
POST /api/monms/content/import
Authorization: Bearer {publish_token}
Body: { "collections": [ { "name": "...", "records": [...] } ] }
```

- Token scoped to **content import only** — not full superuser.
- Upsert by record ID; skip or warn on unknown fields if structure lagged.
- Never imports auth users, admin accounts, or non-editorial collections.

### 5.5 Publish console — Publish to live

**Location:** `/api/monms/publish` (standalone MonMS route — **not** `/_/publish`, which PocketBase's admin SPA would catch).

| Element | Behavior |
|---------|----------|
| Diff preview | Lists collections/records changed since last publish |
| Publish now | Triggers export + POST to production import API |
| Last published | Timestamp + checksum stored in `.monms/publish-state.json` |
| Permissions | Superuser session + **publisher** email allowlist in `.monms/config.json` |

Clients use this console; consultants configure `productionUrl`, `publisherEmails`, and `MONMS_PUBLISH_TOKEN` once at site setup.

---

## 6. Media & Blobs

### 6.1 Decided policy

| Storage | Promotion |
|---------|-----------|
| `workspace/assets/` (Git-tracked) | Structure rail — deploys with workspace tag |
| Public CDN bucket URLs in text/HTML fields | **No blob copy** — URL string upserted in content rail |
| PocketBase file fields → local `.pb_data/storage/` | **Avoid for publishable media** — environment-local unless shared S3 backend |

Inline content and record fields store **canonical public CDN URLs**. Staging and production reference the same objects. Content publish moves text/JSON only.

### 6.2 Consultant setup (once per site)

- Configure one public bucket + CDN prefix for client uploads.
- Prefer text URL fields or PocketBase storage backed by shared S3 — not staging-local file storage for publishable assets.

---

## 7. Roles

| Role | Staging | Production | Structure Git | Content publish |
|------|---------|------------|---------------|-----------------|
| Engine developer | — | — | monms repo | — |
| Consultant / agent | ✓ mutate | deploy tags | ✓ commit/tag | optional |
| Editor | ✓ inline edit | — | — | — |
| Publisher | ✓ inline edit | — | — | ✓ Publish button |
| Audience | — | ✓ view | — | — |

---

## 8. Requirements (v2 — merged into GSD, complete)

### Environment & Lifecycle

- **ENV-01**: Documentation and tooling distinguish four layers (engine, structure, content, audience).
- **ENV-02**: Structure promotion uses workspace Git tags; content promotion is a separate rail.
- **ENV-03**: Staging and production are separate MonMS instances with separate `.pb_data/` directories.

### Content Publish

- **PUB-01**: Editorial collections marked `"editorial": true` in schema JSON.
- **PUB-02**: `workspace/content/*.json` holds exported editorial records with stable IDs.
- **PUB-03**: `monms content export` writes editorial snapshots to `workspace/content/`.
- **PUB-04**: `monms content import` upserts records idempotently by ID.
- **PUB-05**: Production exposes `POST /api/monms/content/import` with scoped publish token.
- **PUB-06**: Staging admin UI includes **Publish to live** with diff preview.
- **PUB-07**: Publisher role gates the publish action; editors may edit without publishing.
- **PUB-08**: Staging tracks last-published checksum for unpublished-changes indicator.
- **PUB-09**: `monms content diff` shows pending changes before publish.

### Media

- **MED-01**: Publishable media uses public CDN URLs stored in content fields; blobs are not copied between environments.
- **MED-02**: Documentation warns against PocketBase-local file storage for publishable assets.

### Non-goals (this spec)

- Full `.pb_data/` backup/restore as primary publish mechanism
- Consultant involvement in every content push
- Blob replication between staging and production

---

## 9. Phase 4 delivery (complete)

**Phase 4: Staging Environments & Client Content Publish** — executed 2026-05-26.

| Deliverable | Status |
|-------------|--------|
| `workspace/content/` + `"editorial": true` in schema JSON | ✓ |
| `internal/content/` — export, import, diff, upsert by ID | ✓ |
| `monms content` CLI subcommands | ✓ |
| `POST /api/monms/content/import` + `MONMS_PUBLISH_TOKEN` auth | ✓ |
| Publish console at `/api/monms/publish` with diff + Publish now | ✓ |
| Publisher allowlist in `.monms/config.json` | ✓ |
| Docs: lifecycle, roles, media (`MEDIA.md`, README updates) | ✓ |

**Requirements covered:** ENV-01–03, PUB-01–09, MED-01–02 (see `.planning/REQUIREMENTS.md`).

---

## 10. Relationship to v1

v1 delivered L1 engine, L2 workspace/Git mutation, and L3 inline editing on a single instance. This spec **extends** v1 without contradicting it:

- Schema dual-write (`schema/*.json`) — unchanged
- HTMX inline editing — unchanged; staging instance is where editing happens
- Pre-commit template validation — unchanged; applies to structure rail
- `.pb_data/` gitignored — unchanged

Phase 4 / v2 milestone scope is **implemented**. Human UAT may continue via `.planning/phases/04-staging-environments-client-content-publish/04-UAT.md`.

---

## 11. Implementation reference (as built)

| Surface | Path / command |
|---------|----------------|
| Publish console | `GET/POST /api/monms/publish`, `GET /api/monms/publish/diff` |
| Production import | `POST /api/monms/content/import` + `Authorization: Bearer $MONMS_PUBLISH_TOKEN` |
| Staging config | `workspace/.monms/config.json` (gitignored; copy from `config.example.json`) |
| Publish state | `workspace/.monms/publish-state.json` (checksum + last publish time) |
| Editorial allowlist | `"editorial": true` in `workspace/schema/*.json` (parsed by `internal/schema/editorial.go`) |
| CLI | `monms content export\|import\|diff\|publish --to URL` |

**Publish behavior:** upsert by record ID only — deletions on staging do not remove production records. Diff reports deleted records for operator awareness.

---

## 12. GSD integration (historical)

Reconciliation completed 2026-05-23 via `gsd-ingest-docs`; Phase 4 plans executed 2026-05-26.

---

*Accepted: 2026-05-23 · Implemented: 2026-05-26*
