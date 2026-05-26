---
phase: 03-inline-contextual-editing-demonstration-content
plan: "01"
subsystem: database
tags: [pocketbase, schema, bootstrap, seed]

requires:
  - phase: 02-agent-mutation-engine-safety-guardrails
    provides: schema sync bootstrap hook pattern
provides:
  - hero_content collection with SEC-02 API rules
  - idempotent homepage record seed on bootstrap
  - monms init ships hero_content.json
affects: [03-02, 03-03, 03-04]

tech-stack:
  added: []
  patterns: [bootstrap seed after schema import, custom id field for fixed homepage record id]

key-files:
  created:
    - workspace/schema/hero_content.json
    - internal/scaffold/embed/hero_content.json
    - internal/schema/seed.go
    - internal/schema/seed_test.go
  modified:
    - internal/schema/sync.go
    - internal/scaffold/init.go
    - internal/scaffold/init_test.go

key-decisions:
  - "hero_content id field allows 1-50 char lowercase slugs so fixed id homepage validates"
  - "Seed failures log warn and do not fail bootstrap (D-60)"

patterns-established:
  - "seedHeroHomepage runs after ImportCollections even when no new schema files"

requirements-completed: [DEMO-01, SEC-02]

duration: 15min
completed: 2026-05-23
---

# Phase 3 Plan 01 Summary

**hero_content collection with public read / authenticated write rules and idempotent homepage seed on every bootstrap.**

## Accomplishments
- Added hero_content schema JSON with listRule/viewRule public and mutation rules requiring auth
- Implemented seedHeroHomepage with fixed id homepage and UI-SPEC seed copy
- Extended monms init to copy hero_content.json into workspace/schema/

## Self-Check: PASSED
- go test ./internal/schema/ -run TestSeedHeroHomepage -count=1
