# MonMS doctree — reference

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

## Binding schema template

Minimal markdown collection:

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
