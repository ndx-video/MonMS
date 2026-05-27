# SPEC: MonMS Staging, Environments & Content Publish

**Status:** Accepted (2026-05-23) · **Implemented** (Phase 4, 2026-05-26)  
**Supersedes:** Implicit single-environment assumptions in v1 planning  
**Related:** `specs/monms-prd.md`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`

---

## 1. Summary

MonMS operates across **four layers** (artifacts) and **four phases of work** (who does what). Structure (templates, schema, assets) promotes via Git tags. Editorial content (PocketBase records) promotes via JSON upsert **outside Git** — initiated by **clients** from the **Publish to live** console at `/_monms/publish`, without consultant involvement for routine updates.

Media referenced by public CDN URLs does **not** move between environments; only URL strings in content records are synced.

---

## 2. Phases of work (terminology)

MonMS documentation uses four **phases of work**. These are distinct from the four **layers** in §3 (what artifacts exist).

| Phase | Who | Focus | Primary artifact |
|-------|-----|-------|------------------|
| **Development** | MonMS engine developers | Building this open-source Go project — router, content sync, validation | `monms` binary (L1) |
| **Shaping** | Consultant, operator, or AI agent | Site design — templates, CSS, collection schema | `site/` Git repo (L2) |
| **Staging** | Client editors and publishers | Preparing and editing editorial copy before go-live | Staging instance `.pb_data/` (L3) |
| **Production** | End audience (read); clients publish into | Live public site | Production instance `.pb_data/` + URL (L3/L4) |

**Note on "development":** In product documentation, *development* means contributing to the MonMS **engine** repository. That is separate from the compile-time **`buildMode=development`** flag (no template cache) used when running a locally built binary.

### Shaping → Git tag → both environments

During **shaping**, the consultant or agent checks out the workspace repository and mutates templates, schema JSON, and assets. When a shape release is ready, they commit, validate (`monms validate`, pre-commit hook), and apply a **Git tag** on the workspace repo (e.g. `v1.2.0`).

A **policy-based deploy mechanism** (outside the scope of this project) then pulls that tag into **both** the staging and production workspace checkouts. Example implementations operators may choose:

- GitHub Actions (or similar CI) triggered on tag push — SSH or webhook to each host
- Cron job or systemd timer calling `monms site sync --ref v1.2.0`
- Optional startup sync via `shapeSync.enabled` in `site/.monms/config.json` (fetch + checkout ref, not `git pull`)

MonMS provides the built-in `monms site sync` helper and optional config hook; operators choose whether to invoke them manually, on a schedule, or at serve startup.

### Content stays out of Git

Editorial **content is not tracked in Git**. It lives in each environment's `.pb_data/` SQLite database at runtime. Clients move approved copy from staging to production via **JSON export/import outside Git** — ephemeral `content/*.json` files or an HTTP payload from the Publish console, never commits on the structure rail.

---

## 3. Four Layers

| Layer | Name | Who | Artifact | Change frequency |
|-------|------|-----|----------|------------------|
| **L1** | Engine | MonMS developers | `monms` binary (this repo) | Rare — semver releases |
| **L2** | Site structure | Consultants, AI agents, advanced clients | `site/` Git repo — `templates/`, `assets/`, `schema/` | Per feature / page |
| **L3** | Content | Business clients (editors/publishers) | `site/.pb_data/` (runtime) + optional `site/content/` (export) | Daily / weekly |
| **L4** | Audience | End users | Production URL (read-only) | Continuous consumption |

**L1 is frozen at deploy.** L2–L4 run on the same binary version. v1 implemented L1–L3 on a single host; this spec defines how L2–L4 split across staging and production environments.

---

## 4. Environments

Typical deployment uses **two MonMS instances** — one for **staging** (client content work) and one for **production** (live audience):

| Environment | Phase | Purpose | Layers active |
|-------------|-------|---------|---------------|
| **Staging** | Staging | Client content editing and preview | L2 (tagged shape), L3 (staging `.pb_data/`) |
| **Production** | Production | Public audience | L2 (tagged shape), L3 (production `.pb_data/`), L4 |

**Shaping** happens on a workspace Git checkout (consultant laptop, CI sandbox, or dedicated clone) — not as routine client work on the staging instance. Both staging and production instances receive the same tagged shape via the operator's deploy policy (§2).

Both instances run the same pinned `monms` binary version unless an engine upgrade is intentional.

### 4.1 Structure promotion (L2) — shaping, infrequent

1. Consultant or agent **shapes** the site on a workspace checkout (templates, schema JSON, assets).
2. Changes committed to workspace Git with validation (`monms validate`, pre-commit hook).
3. Review → **Git tag** on workspace repo (e.g. `v1.2.0`).
4. Operator policy pulls that tag into **both** staging and production workspace checkouts (§2 — e.g. GitHub Actions, cron, or `monms site sync`).
5. fsnotify/cache picks up template changes; bootstrap imports `schema/*.json`.

**Git is the transport for structure only.** Tags do not carry editorial content from `.pb_data/`.

### 4.2 Content promotion (L3→L4) — client-driven, frequent

1. Client edits content on **staging** via inline HTMX editing (existing v1 feature).
2. Client opens **Publish to live** at `/_monms/publish` (or via the editor badge link).
3. Staging exports editorial collections → JSON payload.
4. Staging POSTs to production **content import API** (scoped token).
5. Production **upserts records by fixed ID** — idempotent, no binary restart.
6. Staging records last-publish checksum for “unpublished changes” indicator.

**Consultants are not in the loop for routine content pushes.** One-time setup: production URL, publish token, editorial collection marking.

---

## 5. Two-Rail Model (decided)

```
Structure rail:  workspace Git tag  →  staging + production checkouts (operator policy)
Content rail:    editorial JSON upsert (outside Git)  →  production .pb_data/ records
Media rail:      shared public CDN URLs  →  no blob copy (URL strings only)
```

These rails are **independent**. Structure must reach production before content that depends on new collections or fields.

---

## 6. Content Sync Design

### 6.1 `site/content/` convention

Parallel to `site/schema/`:

```
site/
├── schema/           # L2 — collection shape (existing)
│   └── hero_content.json
├── content/          # L3 — editorial records (new)
│   └── hero_content.json
└── .pb_data/         # Runtime DB (never committed)
```

**Schema JSON** defines fields and API rules. **Content JSON** holds editorial records for export/import/publish only — ephemeral payloads **outside Git**, not committed editorial history. Runtime truth is each environment's `.pb_data/` database.

Example `site/content/hero_content.json`:

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

### 6.2 Editorial collection marking

Collections participating in content sync declare `"editorial": true` in their schema JSON. System collections (`_superusers`, auth tables) are always excluded.

Fixed record IDs (e.g. `homepage`) are required for reliable upsert — same pattern as v1 hero seed.

### 6.3 CLI commands (engine)

| Command | Purpose |
|---------|---------|
| `monms content export` | Snapshot editorial collections → `site/content/*.json` |
| `monms content import` | Upsert from `site/content/*.json` → local `.pb_data/` |
| `monms content diff` | Show records/fields that differ from last export or target |
| `monms content publish --to URL` | Export from running instance + POST to production (operator/CI fallback; requires `MONMS_PUBLISH_TOKEN`) |

Primary UX is the **Publish to live** console at `/_monms/publish`; CLI remains for CI and consultant emergencies.

### 6.4 Production import API

New authenticated endpoint on production:

```
POST /api/monms/content/import
Authorization: Bearer {publish_token}
Body: { "collections": [ { "name": "...", "records": [...] } ] }
```

- Token scoped to **content import only** — not full superuser.
- Upsert by record ID; skip or warn on unknown fields if structure lagged.
- Never imports auth users, admin accounts, or non-editorial collections.

### 6.5 Publish console — Publish to live

**Location:** `/_monms/publish` (standalone MonMS route — **not** `/_/publish`, which PocketBase's admin SPA would catch).

| Element | Behavior |
|---------|----------|
| Diff preview | Lists collections/records changed since last publish |
| Publish now | Triggers export + POST to production import API |
| Last published | Timestamp + checksum stored in `.monms/publish-state.json` |
| Permissions | Superuser session + **publisher** email allowlist in `.monms/config.json` |

Clients use this console; consultants configure `productionUrl`, `publisherEmails`, and `MONMS_PUBLISH_TOKEN` once at site setup.

---

## 7. Media & Blobs

### 7.1 Decided policy

| Storage | Promotion |
|---------|-----------|
| `site/assets/` (Git-tracked) | Structure rail — deploys with workspace tag |
| Public CDN bucket URLs in text/HTML fields | **No blob copy** — URL string upserted in content rail |
| PocketBase file fields → local `.pb_data/storage/` | **Avoid for publishable media** — environment-local unless shared S3 backend |

Inline content and record fields store **canonical public CDN URLs**. Staging and production reference the same objects. Content publish moves text/JSON only.

### 7.2 Consultant setup (once per site)

- Configure one public bucket + CDN prefix for client uploads.
- Prefer text URL fields or PocketBase storage backed by shared S3 — not staging-local file storage for publishable assets.

---

## 8. Roles

| Role | Shaping | Staging instance | Production instance | Structure Git | Content publish |
|------|---------|------------------|---------------------|---------------|-----------------|
| Engine developer | — | — | — | monms repo (L1) | — |
| Consultant / agent | ✓ shape | — | — | ✓ commit/tag | optional |
| Editor | — | ✓ inline edit | — | — | — |
| Publisher | — | ✓ inline edit | — | — | ✓ Publish button |
| Audience | — | — | ✓ view | — | — |

---

## 9. Requirements (v2 — merged into GSD, complete)

### Environment & Lifecycle

- **ENV-01**: Documentation and tooling distinguish four layers (engine, structure, content, audience).
- **ENV-02**: Structure promotion uses workspace Git tags; content promotion is a separate rail.
- **ENV-03**: Staging and production are separate MonMS instances with separate `.pb_data/` directories.

### Content Publish

- **PUB-01**: Editorial collections marked `"editorial": true` in schema JSON.
- **PUB-02**: `site/content/*.json` holds exported editorial records with stable IDs.
- **PUB-03**: `monms content export` writes editorial snapshots to `site/content/`.
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

## 10. Phase 4 delivery (complete)

**Phase 4: Staging Environments & Client Content Publish** — executed 2026-05-26.

| Deliverable | Status |
|-------------|--------|
| `site/content/` + `"editorial": true` in schema JSON | ✓ |
| `internal/content/` — export, import, diff, upsert by ID | ✓ |
| `monms content` CLI subcommands | ✓ |
| `POST /api/monms/content/import` + `MONMS_PUBLISH_TOKEN` auth | ✓ |
| Publish console at `/_monms/publish` with diff + Publish now | ✓ |
| Publisher allowlist in `.monms/config.json` | ✓ |
| Docs: lifecycle, roles, media (`MEDIA.md`, README updates) | ✓ |

**Requirements covered:** ENV-01–03, PUB-01–09, MED-01–02 (see `.planning/REQUIREMENTS.md`).

---

## 11. Relationship to v1

v1 delivered L1 engine, L2 site/Git mutation, and L3 inline editing on a single instance. This spec **extends** v1 without contradicting it:

- Schema dual-write (`schema/*.json`) — unchanged
- HTMX inline editing — unchanged; staging instance is where editing happens
- Pre-commit template validation — unchanged; applies to structure rail
- `.pb_data/` gitignored — unchanged

Phase 4 / v2 milestone scope is **implemented**. Human UAT may continue via `.planning/phases/04-staging-environments-client-content-publish/04-UAT.md`.

---

## 12. Implementation reference (as built)

| Surface | Path / command |
|---------|----------------|
| Publish console | `GET/POST /_monms/publish`, `GET /_monms/publish/diff` |
| Production import | `POST /api/monms/content/import` + `Authorization: Bearer $MONMS_PUBLISH_TOKEN` |
| Staging config | `site/.monms/config.json` (gitignored; copy from `config.example.json`) |
| Publish state | `site/.monms/publish-state.json` (checksum + last publish time) |
| Editorial allowlist | `"editorial": true` in `site/schema/*.json` (parsed by `internal/schema/editorial.go`) |
| CLI | `monms content export\|import\|diff\|publish --to URL` |
| Shape sync CLI | `monms site sync --ref TAG [--remote origin] [--force]` |
| Shape sync config | `shapeSync` in `site/.monms/config.json` (optional startup fetch + checkout) |

**Publish behavior:** upsert by record ID only — deletions on staging do not remove production records. Diff reports deleted records for operator awareness.

---

## 13. GSD integration (historical)

Reconciliation completed 2026-05-23 via `gsd-ingest-docs`; Phase 4 plans executed 2026-05-26.

---

*Accepted: 2026-05-23 · Implemented: 2026-05-26*
