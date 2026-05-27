# Templates and routing

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)  
> **Go templates:** [`html/template` reference](https://pkg.go.dev/html/template)

## Page template structure

Every route template defines a `body` block — the base layout wraps it:

```html
{{define "body"}}
<section>
  <h1>My Page</h1>
</section>
{{end}}
```

Do **not** re-declare `<!DOCTYPE html>`, `<html>`, `<head>`, or `<body>` in page templates. The shell lives in `templates/layouts/base.gohtml`.

## Slug → template path (mirror + index)

| URL | Template path |
|-----|---------------|
| `/` | `templates/index.gohtml` |
| `/press` | `templates/press/index.gohtml` or `templates/press.gohtml` |
| `/press/2024` | `templates/press/2024.gohtml` |
| `/about` | `templates/about/index.gohtml` or `templates/about.gohtml` |

404 after adding a page usually means the template path does not mirror the URL slug.

## Fragments

Files in `templates/fragments/` are served at `/fragments/{name}` **without** the base layout. Do not use `{{define "body"}}` in fragments.

## Inline editing attributes

Gate `contenteditable` and HTMX save attrs on `{{if .IsLoggedIn}}`. The base layout injects the PocketBase Bearer token into HTMX via `htmx:configRequest` — not from `document.cookie` (HttpOnly session).

Example pattern:

```html
<h1
  {{if .IsLoggedIn}}
  contenteditable="true"
  hx-patch="/api/collections/hero_content/records/homepage"
  hx-trigger="blur"
  hx-ext="json-enc"
  hx-vals='js:{"title": event.target.innerText}'
  hx-swap="none"
  {{end}}>
  {{.Hero.Title}}
</h1>
```

## Validation

`monms validate` uses the same `ParseFiles(layout, page)` paths as production SSR. Run before commit; the pre-commit hook runs it on staged `*.gohtml` files.

## Related

- [Shaping and agents](shaping-and-agents.md)
- [Inline editing](../user-guide/inline-editing.md)
