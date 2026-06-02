# MonMS doctree — reference

## Path-driven `dt_*` collection names

| Leaf folder (under stub `guide`) | `monms.root` | Collection name |
|----------------------------------|--------------|-----------------|
| `.md` directly in `guide/` | `guide` | `dt_guide` |
| `guide/tutorials/` | `guide/tutorials` | `dt_guide_tutorials` |
| `guide/tutorials/advanced/` | `guide/tutorials/advanced` | `dt_guide_tutorials_advanced` |

Names are never singularized. After folder rename/move: `/_monms/doctrees` alignment panel → re-scan → confirm → retire old PB collection manually.

## Doctree frontmatter (confirm / sync)

```yaml
---
id: dt_guide_tutorials--intro
title: Intro
description:
ts_create: "2026-06-02T10:00:00Z"
ts_mod: "2026-06-02T12:00:00Z"
---
# Intro

Body.
```

- **`title`**: on first confirm, set from first `# ` heading in body, else humanized basename; not overwritten on re-confirm if already set.
- **`description`**: nullable; not auto-filled in v1.
- **`ts_mod`**: content LWW vs file mtime for `dt_*` sync.

## MCP doctree tools

Auth: `Authorization: Bearer <monms_api_key>` on `http://<mcp.host>:<mcp.port>/mcp`.

### monms_doctree_write

```json
{
  "collection": "articles",
  "path": "guides/setup",
  "body": "## Intro\n\nUpdated content.",
  "meta": {
    "id": "articles--guides--setup",
    "title": "Setup Guide"
  },
  "sync": true
}
```

Response includes `written` path and `synced` record count.

### monms_doctree_get

By slug (searches all markdown collections):

```json
{ "slug": "guides/setup" }
```

By filesystem path:

```json
{ "collection": "articles", "path": "guides/setup" }
```

Returns `meta`, `body`, `filePath`, `id`, `slug`, `source` (`filesystem` or `pocketbase`).

### monms_doctree_sections

All sections:

```json
{ "slug": "guides/setup" }
```

Single section (level 2, first `##`):

```json
{ "slug": "guides/setup", "level": 2, "index": 0 }
```

Section shape: `{ level, index, title, anchor, source }` — `source` is raw markdown for that section.

### monms_doctree_forest

Returns `collections[]` with nested `folders`, `orphans`, and top-level `orphanCount` from `DiffOrphans`.

## CLI parity

| MCP | CLI |
|-----|-----|
| `monms_doctree_sync` | `monms documents sync` |
| `monms_doctree_diff` | `monms documents diff` |
| `monms_doctree_list` | (filesystem; no direct CLI — use `documents scan` on external trees) |
| bindings / schema | `monms documents bind --apply` after `plan` |

## Legacy migration plan YAML

```yaml
- collection: articles
  sourceRoot: ./legacy-docs/blog
  destRoot: documents/articles
  fieldMap:
    date: published_at
```

One collection per content **type**; subfolders become `path` / `folder` / `slug`, not separate collections.

## Template patterns

**Hero pulling doc intro:**

```gohtml
{{ define "body" }}
<section class="hero">
  <h1>{{ docHeading "guides/setup" 1 0 }}</h1>
  <div class="prose">{{ docSection "guides/setup" 2 0 }}</div>
</section>
{{ end }}
```

**Doc layout with section nav:**

```gohtml
{{ define "body" }}
<nav>
  {{ range .Doc.Sections }}
    <a href="#{{ .Anchor }}">{{ .Title }}</a>
  {{ end }}
</nav>
<article>{{ .Doc.HTMLBody }}</article>
{{ end }}
```

## Orphan handling

Orphans = PB records whose `path` has no backing `.md`. After deleting files, run sync; orphans remain until manually deleted in PocketBase admin or re-created on disk. Use `monms_doctree_diff` before releases.

## Binding schema templates

**Doctree** (`dt_*`, created by migration confirm — use `DefaultDoctreeCollectionSchema` in engine):

```json
{
  "name": "dt_guide_tutorials",
  "type": "base",
  "editorial": true,
  "monms": {
    "source": "markdown",
    "root": "guide/tutorials",
    "doctree": "guide",
    "slugFrom": "path",
    "idFrom": "frontmatter.id",
    "fields": {
      "title": "title",
      "description": "description",
      "ts_create": "ts_create",
      "ts_mod": "ts_mod",
      "date": "published_at"
    }
  },
  "fields": [
    { "name": "id", "type": "text", "primaryKey": true, "required": true },
    { "name": "title", "type": "text" },
    { "name": "description", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "folder", "type": "text" },
    { "name": "body", "type": "text" },
    { "name": "doctree_id", "type": "text" },
    { "name": "leaf_path", "type": "text" },
    { "name": "ts_create", "type": "text" },
    { "name": "ts_mod", "type": "text" }
  ]
}
```

**Legacy** markdown collection (`documents/articles`):

```json
{
  "name": "articles",
  "type": "base",
  "editorial": true,
  "monms": {
    "source": "markdown",
    "root": "documents/articles"
  },
  "fields": [
    { "name": "id", "type": "text", "primaryKey": true, "required": true,
      "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "folder", "type": "text" },
    { "name": "body", "type": "text" }
  ]
}
```

PocketBase strips `monms` on import — keep the block in Git schema files only.
