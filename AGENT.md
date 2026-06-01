# AGENT.md: The Master Orchestrator

You are an autonomous development agent. This file is your root operating directive. Do not load massive context files blindly. Your primary operating principle is **Lazy Loading via Modular Skills**. 

You must follow the Discovery, Evaluation, and Extension loop below for every task.

## 1. The Skill Discovery Phase ("Sniffing")
Before writing any code or proposing an architecture, you must determine the required tech stack for the task (e.g., Go, Astro, Deno Fresh, Docker, Kubernetes). 

1. **Scan Frontmatter:** Look inside `.cursor/skills/`. Read **only** the YAML frontmatter or the first 10 lines of each `SKILL.md` to understand scope.
2. **Load Selectively:** Only read the full contents of the skill files that directly apply to the current task.

### Project skills (MonMS)

| Skill | Load when |
|-------|-----------|
| `monms-architecture` | Any MonMS task; cold start; terminology (site directory vs `./site` default, staging vs buildMode) |
| `monms-site-shaping` | Editing templates, schema, or assets in the configured site directory; agent L2 mutations |
| `monms-engine-development` | Editing `internal/*`, `main.go`, CLI, engine tests |
| `monms-dash` | `/_monms/` operator console, notifications, new dashboard pages |
| `monms-operators-deploy` | Docker, `monms site sync`, `.monms/config.json`, logging, multi-instance setup |
| `monms-extensibility` | Plugins, in-process extensions, Sentinel integration requests |

Start with `monms-architecture`, then load the task-specific skill. Skills live at `.cursor/skills/<name>/SKILL.md`.

## 2. The Gap Analysis & Decision Matrix
If a required skill is missing or incomplete, do not guess. You must act autonomously to bridge the gap.

**Scenario A: The Skill is Missing Entirely**
If the task requires a specific technology (e.g., building a new LXC template) and no skill file exists:
* **Evaluate:** Is this a one-off script, or a recurring pattern for this codebase?
* **Action:** If it is a recurring pattern, draft a new modular skill file (e.g., `.cursor/skills/proxmox-lxc/SKILL.md`) with a clear YAML frontmatter description. Ask the user for quick validation, then save it.

**Scenario B: The Skill Exists, but Lacks Detail**
If a skill file exists (e.g., `monms-engine-development`) but lacks the specific architectural constraints needed for the current task:
1. **Search Local Context:** Search `./docs/*.md` or relevant project documentation to find the established pattern.
2. **Search the Web:** If local docs are insufficient, run a web search for the official documentation of the library/framework in question to find current best practices.
3. **Ask for Clarification:** If architectural ambiguity remains, ask the user a targeted, multiple-choice question to clarify the preferred approach.
4. **Extend:** Once you have the missing details, permanently update the existing skill file with the new constraints so you do not have to ask again in the future. 

## 3. Universal Hard Boundaries
*These rules apply globally, regardless of the active skill files.*
* **Sovereignty & Self-Hosting:** Always default to configurations that prioritize self-hosted, sovereign infrastructure over managed cloud services.
* **Terminal Operations:** When providing command-line instructions or editing configuration files manually in the terminal, **always use `vi`**. Never suggest or use `nano`.
* **Security:** Never hardcode secrets. Always assume a reverse proxy (like Tailscale/Headscale) is handling external access.

## 4. Skill File Template
*When generating a new skill file, use this structure:*

---
name: [technology-domain]
description: [1-2 sentences summarizing when to load this file. e.g., "Rules for the Sentinel Go control plane and identity brokering."]
---
# [Domain] Standards
[Detailed architectural rules, preferred libraries, and anti-patterns...]

If you're at the begining of a context window (cold start) and the user assumes you know more about the project than you do, then go read ./PROJECT.md
