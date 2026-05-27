# MonMS Documentation

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1) (see [`go.mod`](../go.mod))  
> **PocketBase REST/admin API:** [official PocketBase docs](https://pocketbase.io/docs/) — not duplicated here  
> **Go templates:** [`html/template` package](https://pkg.go.dev/html/template)

MonMS is an agent-malleable, single-binary CMS. Consultants and agents **shape** structure in a Git-tracked `site/`; clients **stage** editorial copy and publish to production without rebuilding the engine.

**Published docs (forthcoming):** this tree is the source of truth in the engine repository. A GitHub Pages site will mirror it at a public URL (TBD).

---

## Documentation roadmap

### Content editors (user guide)

| Document | What you will learn |
|----------|---------------------|
| [Inline editing](user-guide/inline-editing.md) | Sign in at `/_/`, edit copy on the live page, logout safety |
| [Publish to live](user-guide/publish-to-live.md) | Publisher workflow at `/_monms/publish` |
| [Media URLs](user-guide/media-urls.md) | CDN URL policy — no blob copy between environments |

### Operators (shapers and administrators)

| Document | What you will learn |
|----------|---------------------|
| [Getting started](operators/getting-started.md) | Four layers, phases of work, staging vs production |
| [Architecture overview](operators/architecture-overview.md) | Product vision (verified against the codebase) |
| [Shaping and agents](operators/shaping-and-agents.md) | Schema dual-write, templates, validation, `agent:` commits |
| [Templates and routing](operators/templates-and-routing.md) | Mirror+index slug rules, fragments, inline edit attrs |
| [Security](operators/security.md) | SSH scope, tokens, git hygiene |
| [Docker deploy](operators/deploy-docker.md) | Optional L1 image + L2/L3 volumes |
| [Extensibility with Sentinel](operators/extensibility-with-sentinel.md) | Why MonMS has no plugins; Sentinel integration (forthcoming) |

### Reference

| Document | What you will learn |
|----------|---------------------|
| [MonMS HTTP API](reference/monms-api.md) | `/api/monms/*` and `/_monms/*` only |
| [CLI reference](reference/cli.md) | `monms init`, `validate`, `content`, `site sync`, `serve` |
| [External dependencies](reference/external-dependencies.md) | PocketBase version, PocketBase API, Go templates |

---

## Quick links

| Audience | Start here |
|----------|------------|
| Client editor | [Inline editing](user-guide/inline-editing.md) |
| Publisher | [Publish to live](user-guide/publish-to-live.md) |
| Consultant / agent | [Shaping and agents](operators/shaping-and-agents.md) |
| Integrator | [MonMS HTTP API](reference/monms-api.md) |

Legacy product specs in [`specs/`](../specs/) are deprecated — prefer this tree and the live codebase.
