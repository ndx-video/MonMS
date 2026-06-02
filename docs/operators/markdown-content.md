# Markdown document rail

MonMS supports a **dual-rail** editorial model:

| Rail | Canonical store | Promote to production | Client editing |
|------|-------------------|----------------------|----------------|
| **PB-native** | `.pb_data/` records | `/_monms/publish` (JSON import) | Inline HTMX on live pages |
| **Markdown** | Git `documents/**/*.md` | Git tag + `monms site sync` | Git PR / consultant edit |

Collections declare their rail in `{siteDir}/schema/{name}.json` via the MonMS-only `monms` block (stripped by PocketBase on import, same as `editorial`).

## Schema binding

Example markdown collection:

```json
{
  "name": "articles",
  "type": "base",
  "editorial": true,
  "monms": {
    "source": "markdown",
    "root": "documents/articles",
    "slugFrom": "path",
    "idFrom": "frontmatter.id",
    "fields": { "date": "published_at" }
  },
  "fields": [
    { "name": "id", "type": "text", "primaryKey": true, "required": true,
      "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "folder", "type": "text" },
    { "name": "body", "type": "text" },
    { "name": "published_at", "type": "text" }
  ]
}
```

Omit `monms.source` (or set `"pocketbase"`) for collections on the JSON publish rail.

## Frontmatter contract

```yaml
---
id: articles--guides--setup
title: Setup Guide
status: published
date: 2024-03-15
---
# Setup

Body markdown here.
```

- **`id`** — stable PocketBase primary key. Use `--` instead of `/` (PocketBase disallows slashes in IDs). If omitted, the engine generates `{collection}--{path-with-dashes}`.
- **`slug`** — URL segment(s), derived from file path by default (e.g. `guides/setup` for `documents/articles/guides/setup.md`).
- **Body** — maps to the `body` field (or `monms.body` override in schema).

## Sync

Markdown files sync into PocketBase on:

1. **Serve bootstrap** (after schema import)
2. **`monms documents sync`**
3. **Production fsnotify** (`.md` changes under bound `monms.root`)

```bash
monms documents sync --site ./site
monms documents diff --site ./site   # orphan PB records without backing files
```

Markdown collections are **excluded** from `monms content export/publish` and grayed out in `/_monms/publish`.

## Rendering

When no `.gohtml` template matches a URL slug, MonMS looks up a markdown-backed record by `slug` and renders `templates/doc.gohtml` through the base layout. **Templates win** over markdown routes (same precedence as flat-file-over-index for templates).

### Section selection in `.gohtml`

Public templates can pull heading-bounded slices from any synced markdown document by **level** (1–6) and **per-level index** (0-based):

```gohtml
<h2>{{ docHeading "guides/setup" 2 0 }}</h2>
<div class="prose">{{ docSection "guides/setup" 2 0 }}</div>
```

| Func | Args | Returns |
|------|------|---------|
| `docSection` | slug, level, index | Rendered HTML for section body |
| `docHeading` | slug, level, index | Heading title string |
| `docSections` | slug | All sections (for `range`) |

On markdown doc pages, `.Doc.Sections` provides the same data (with pre-rendered `.HTML`) for use in `doc.gohtml`.

## Operator dashboard

Authenticated users can browse synced document trees at **`/_monms/documents`** — one panel per markdown collection, folder hierarchy, links to public slugs. Promote via Git, not Publish to live.

## Agent access (MCP)

When MCP is enabled, doctree tools read/write Git-canonical files and sync the PocketBase index — see [MCP and API keys](mcp-and-api-keys.md). Prefer `monms_doctree_write` over `monms_update_record` for markdown collections.

## Legacy migration

Four-step workflow:

```bash
# 1. Inventory
monms documents scan --source ./legacy-docs

# 2. Propose bindings (review/edit plan.yaml)
monms documents plan --source ./legacy-docs --out plan.yaml

# 3. Apply (writes schema, copies files, injects frontmatter)
monms documents bind --config plan.yaml --apply --site ./site

# 4. Verify sync
monms documents sync --site ./site
```

Use `--dry-run` on `bind` to preview without writing. Use `--force` to overwrite existing frontmatter `id` values.

### Plan YAML shape

```yaml
- collection: articles
  sourceRoot: ./legacy-docs/blog
  destRoot: documents/articles
  fieldMap:
    date: published_at
```

One collection per content **type**, not per sub-folder. Sub-folders become `path`, `folder`, and `slug` fields — not separate PocketBase collections.

## Promotion checklist (markdown)

1. Commit `documents/` and `schema/` changes in the site Git repo
2. Tag release (structure rail)
3. `monms site sync --ref vX.Y.Z` on staging and production
4. Restart or rely on bootstrap + fsnotify to refresh `.pb_data/` index

PB-native collections still use **Publish to live** separately.

## Related

- [Getting started](getting-started.md) — layers and rails overview
- [CLI reference](../reference/cli.md) — `monms documents` subcommands
