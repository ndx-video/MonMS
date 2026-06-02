---
name: monms-doctree
description: >-
  MonMS markdown doctree — create and sync Git-canonical documents, manage schema
  bindings, browse folder trees, and wire docSection/docHeading into .gohtml templates.
  Use when adding markdown content, binding collections, documents sync, doctree MCP tools,
  or pulling markdown sections into site templates.
---

# MonMS Doctree

Markdown editorial rail: **Git `{siteDir}/documents/` is canonical**; PocketBase is a derived index for SSR, lookup, and `/_monms/documents`.

Foundation: [monms-architecture](../monms-architecture/SKILL.md). Operator reference: [markdown-content.md](../../../docs/operators/markdown-content.md).

## Which directory?

Resolve `{siteDir}` via `-s` / `MONMS_SITE` / default `./site` — same as [monms-site-shaping](../monms-site-shaping/SKILL.md).

## Dual rail reminder

| Rail | Edit | Promote |
|------|------|---------|
| PB-native | `/_monms/publish`, inline HTMX | Publish to live |
| **Markdown** | Git `documents/**/*.md` | Git tag + `monms site sync` |

Never use `monms_update_record` on markdown collections — use filesystem + sync (MCP doctree tools or CLI).

## Schema binding (new collection)

1. Write `{siteDir}/schema/{name}.json` with `editorial: true` and `monms` block:

```json
"monms": {
  "source": "markdown",
  "root": "documents/articles",
  "slugFrom": "path",
  "idFrom": "frontmatter.id",
  "fields": { "date": "published_at" }
}
```

2. Include PB fields: `id`, `title`, `slug`, `path`, `folder`, `body` (+ mapped frontmatter fields).
3. Restart or bootstrap so schema imports; sync runs automatically on serve.
4. Legacy bulk bind: `monms documents plan` → `monms documents bind --apply` — see [reference.md](reference.md).

## Document file contract

Path under `monms.root` → URL slug (default `slugFrom: path`):

| File | slug | PB id |
|------|------|-------|
| `documents/articles/hello.md` | `hello` | `articles--hello` |
| `documents/articles/guides/setup.md` | `guides/setup` | `articles--guides--setup` |

Frontmatter:

```yaml
---
id: articles--guides--setup
title: Setup Guide
date: 2024-03-15
---
# Setup

Body markdown.
```

- **`id`**: use `--` not `/` in IDs. Omit to auto-generate `{collection}--{path-with-dashes}`.
- **Body** → `body` field (or `monms.body` override in schema).

## Agent workflows

### Create or update a document (CLI)

```bash
# Write file, then sync
vi "$SITE/documents/articles/guides/new-page.md"
monms documents sync -s "$SITE"
monms documents diff -s "$SITE"   # optional: check orphans
```

### Create or update (MCP)

Requires `mcp.enabled` and `Authorization: Bearer monms_…` — see [mcp-and-api-keys.md](../../../docs/operators/mcp-and-api-keys.md).

| Tool | Use |
|------|-----|
| `monms_doctree_bindings` | List markdown collections + `monms.root` |
| `monms_doctree_forest` | Nested folder tree from PB index |
| `monms_doctree_list` | Files on disk for one collection |
| `monms_doctree_get` | Read by `slug` or `collection`+`path` |
| `monms_doctree_write` | Create/update `.md` + sync (default) |
| `monms_doctree_delete` | Remove `.md` + sync |
| `monms_doctree_sync` | Full or per-collection sync |
| `monms_doctree_diff` | Orphan PB records |
| `monms_doctree_sections` | Heading sections for template wiring |

Prefer **`monms_doctree_write`** over patching PB records directly.

### Wire sections into `.gohtml`

Public SSR exposes FuncMap helpers (any page template, not only doc fallback):

```gohtml
<h2>{{ docHeading "guides/setup" 2 0 }}</h2>
<div class="prose">{{ docSection "guides/setup" 2 0 }}</div>

{{ range docSections "guides/setup" }}
  <h{{ .Level }}>{{ .Title }}</h{{ .Level }}>
{{ end }}
```

Args: **slug**, heading **level** (1–6), **index** (0-based per level). Templates win over markdown slug routes.

On markdown doc pages, `{{ range .Doc.Sections }}` works in `templates/doc.gohtml`.

Use `monms_doctree_sections` to inspect headings before choosing level/index.

### Verify

```bash
monms validate -s "$SITE"
go test ./internal/documents/... -short   # engine changes only
```

Browser: `/_monms/documents` (authenticated) — folder tree + public links.

## Checklist — new markdown page

- [ ] Binding exists in `schema/{collection}.json` with `monms.source=markdown`
- [ ] File under `{siteDir}/{monms.root}/…/*.md` with frontmatter `id` + `title`
- [ ] `monms documents sync` or MCP `monms_doctree_sync`
- [ ] Public URL `/{slug}` renders (or custom `.gohtml` if template exists)
- [ ] If embedding sections elsewhere: `docSection` / `docHeading` in target template
- [ ] Commit `documents/` + `schema/` in site Git repo; promote via structure rail

## Engine touchpoints

| Concern | Package / path |
|---------|----------------|
| Sync, orphans, bind CLI | `internal/documents/` |
| Forest / dashboard tree | `internal/documents/tree.go`, `/_monms/documents` |
| Section parse | `internal/documents/sections.go` |
| SSR FuncMap | `internal/router/templatefuncs.go` |
| MCP doctree tools | `internal/mcp/doctree.go` |

## Related skills

| Task | Skill |
|------|-------|
| Schema dual-write, `.gohtml` pages | [monms-site-shaping](../monms-site-shaping/SKILL.md) |
| Dashboard documents page | [monms-dash](../monms-dash/SKILL.md) |
| MCP keys, listener config | [mcp-and-api-keys.md](../../../docs/operators/mcp-and-api-keys.md) |

Detailed MCP payloads and migration: [reference.md](reference.md).
