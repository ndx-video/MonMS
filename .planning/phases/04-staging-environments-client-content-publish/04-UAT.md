---
status: testing
phase: 04-staging-environments-client-content-publish
source: 04-01-SUMMARY.md, 04-02-SUMMARY.md, 04-03-SUMMARY.md, 04-04-SUMMARY.md, 04-05-SUMMARY.md, 04-06-SUMMARY.md
started: 2026-05-26T13:50:42Z
updated: 2026-05-26T14:10:00Z
---

## Current Test

number: 4
name: Publisher Badge Link
expected: |
  Sign in at `/_/` with an email on the publisher allowlist in `.monms/config.json`. Visit `/`.
  Live Editor badge includes a **Publish to live** link pointing to `/_monms/publish`.
  Clicking the link opens the publish console HTML page (not a JSON 401).
awaiting: user response

## Tests

### 1. Cold Start Smoke Test
expected: Kill running server, start monms fresh — boots on 8090, `/` loads hero content, `/_/` admin reachable
result: pass

### 2. Content Export CLI
expected: Run `./monms content export -s site-stage`. Creates or updates `site-stage/content/hero_content.json` with editorial records (collection + records array).
result: pass

### 3. Content Diff After Inline Edit
expected: Edit hero title or body via inline editing on staging, then run `./monms content diff -s site-stage`. Terminal shows field-level pending changes (e.g. title old → new). Exit code is non-zero.
result: pass

### 4. Publisher Badge Link
expected: Sign in at `/_/` with an email on the publisher allowlist in `.monms/config.json`. Visit `/`. Live Editor badge includes a **Publish to live** link pointing to `/_monms/publish`. Clicking opens publish console HTML.
result: pass
fix: "Commit 269afc1 — authBind (bindLoadAuth) added to all three publish routes before RequireSuperuserAuth"

### 5. Non-Publisher Badge Hidden
expected: Sign in at `/_/` with a superuser email NOT on the publisher allowlist. Visit `/`. Live Editor badge shows **Full Admin Dashboard** but no **Publish to live** link.
result: [pending]

### 6. Publish Console Page
expected: As a publisher, open `/_monms/publish`. Page shows title "Publish to live", last published timestamp, status indicator (Up to date or Unpublished changes), pending changes diff section, and **Publish now** button.
result: [pending]

### 7. Publish Console Setup Mode
expected: With `productionUrl` unset or empty in `.monms/config.json`, open `/_monms/publish` as publisher. Page shows setup instructions referencing `config.example.json` — not a crash or blank page.
result: [pending]

### 8. Publish Now Workflow
expected: With staging configured (productionUrl + matching MONMS_PUBLISH_TOKEN on both instances), edit hero content on staging, open publish console, click **Publish now**. Success message appears, last published timestamp updates, and production homepage shows the new copy.
result: pass

### 9. Publish Failure Preserves State
expected: Trigger a failed publish (wrong token, production down, or bad URL). Error message displays on the publish page. Last published timestamp and checksum do NOT advance — retry remains possible after fixing the issue.
result: [pending]

### 10. Production Import Auth
expected: POST to `/api/monms/content/import` without a valid Bearer token returns HTTP 401. With correct `MONMS_PUBLISH_TOKEN` Bearer header and valid editorial payload, returns success JSON with upserted counts.
result: pass (verified via monms content publish CLI)

### 11. Client Documentation
expected: `workspace/EDITING-GUIDE.md` section 6 documents the publish workflow (prerequisites, badge link, diff review, Publish now). `workspace/MEDIA.md` explains CDN URL policy for publishable assets.
result: [pending]

## Summary

total: 11
passed: 6
issues: 0
pending: 5
skipped: 0
blocked: 0

## Gaps

- truth: "Clicking Publish to live from the editor badge opens the publish console for authenticated publishers"
  status: failed
  reason: "User reported: When I hit Publish to live, I get {\"data\":{},\"message\":\"The request requires valid record authorization token.\",\"status\":401}"
  severity: blocker
  test: 4
  root_cause: "Publish routes used apis.RequireSuperuserAuth() without loading monms_auth HttpOnly cookie first; SSR pages call LoadAuthFromCookie via withAuthCookie but /_monms/publish did not"
  artifacts:
    - path: "internal/content/publish_handlers.go"
      issue: "registerPublishRoutes missing cookie auth bind before RequireSuperuserAuth"
    - path: "internal/router/ssr.go"
      issue: "withAuthCookie loads cookie for SSR only"
  missing:
    - "Bind LoadAuthFromCookie on publish routes before RequireSuperuserAuth"
  debug_session: ""
