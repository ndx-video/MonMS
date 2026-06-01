# Site Shaping Reference — press_releases walkthrough

End-to-end example from [docs/operators/shaping-and-agents.md](../../../docs/operators/shaping-and-agents.md).

Set the site directory for this instance (default name in demos is `./site`; staging might be `./site-stage`):

```bash
export SITE=./site   # or ./site-stage, ./site-prod, etc.
```

All file paths below are under `$SITE`.

## 1. Obtain admin token

```bash
export MONMS_URL="http://127.0.0.1:8090"
curl -s -X POST "$MONMS_URL/api/collections/_superusers/auth-with-password" \
  -H "Content-Type: application/json" \
  -d '{"identity":"admin@example.com","password":"your-admin-password"}' \
  | jq -r '.token'
export POCKETBASE_ADMIN_TOKEN="<token>"
```

## 2. POST collection via PocketBase API

```bash
curl -s -X POST "$MONMS_URL/api/collections" \
  -H "Authorization: Bearer $POCKETBASE_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "press_releases",
    "type": "base",
    "fields": [
      {"name": "title", "type": "text"},
      {"name": "body",  "type": "text"}
    ]
  }'
```

## 3. Write schema audit file

Create `$SITE/schema/press_releases.json`:

```json
{
  "name": "press_releases",
  "type": "base",
  "editorial": true,
  "listRule": "",
  "viewRule": "",
  "updateRule": "@request.auth.id != ''",
  "createRule": "@request.auth.id != ''",
  "deleteRule": "@request.auth.id != ''",
  "fields": [
    {"name": "title", "type": "text"},
    {"name": "body", "type": "text"}
  ]
}
```

Add `"editorial": true` only if clients should publish this collection to production.

## 4. Create page template

Create `$SITE/templates/press/index.gohtml`:

```html
{{define "body"}}
<section>
  <h1>Press Releases</h1>
</section>
{{end}}
```

## 5. Validate

```bash
cd "$SITE"
monms validate --site .
# or from anywhere:
monms validate -s "$SITE"
```

Binary resolution order: `MONMS_BIN` → `monms` on PATH → path relative to the site directory (see shaping skill).

## 6. Commit

```bash
git add schema/press_releases.json templates/press/index.gohtml
git commit -m "agent: add press_releases collection and press index template"
```

## 7. Verify

Open `/press` — renders without server restart (dev reads from disk; production invalidates cache via watcher).

If 404: verify template path is `templates/press/index.gohtml` or `templates/press.gohtml`.
