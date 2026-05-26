---
phase: 03-inline-contextual-editing-demonstration-content
plan: "02"
subsystem: auth
tags: [pocketbase, ssr, cookie, httponly]

requires:
  - phase: 03-inline-contextual-editing-demonstration-content
    provides: hero_content collection for Hero map load
provides:
  - HttpOnly monms_auth cookie bridge on auth
  - enrichSSRData with AuthToken and Hero fields
affects: [03-03, 03-04]

tech-stack:
  added: []
  patterns: [LoadAuthFromCookie before SSR, NewAuthToken for HTMX Bearer injection]

key-files:
  created:
    - internal/router/auth.go
    - internal/router/auth_test.go
  modified:
    - internal/router/ssr.go
    - main.go

key-decisions:
  - "Cookie name monms_auth (not pb_auth) keeps SEC-04 HttpOnly separate from admin localStorage"
  - "Hero map only on index slug; fallback copy on missing record"

patterns-established:
  - "withAuthCookie wrapper on SSR and fragment routes"

requirements-completed: [ICE-02, ICE-05, SEC-04, DEMO-02]

duration: 15min
completed: 2026-05-23
---

# Phase 3 Plan 02 Summary

**Browser SSR auth bridge via HttpOnly cookie and index-only Hero/AuthToken enrichment for inline editing.**

## Accomplishments
- RegisterAuthHooks sets monms_auth cookie on OnRecordAuthRequest
- LoadAuthFromCookie populates e.Auth before SSR handlers run
- enrichSSRData supplies AuthToken via NewAuthToken and Hero from hero_content/homepage

## Self-Check: PASSED
- go test ./internal/router/ -run 'TestLoadAuth|TestEnrichSSR' -count=1
