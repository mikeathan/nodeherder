# Agent configuration

The root [AGENTS.md](../../AGENTS.md) is the shared entry point and routes work to
the proposed constitutions, scoped engineering rules, and SDD workflow. Its linked
files are authoritative for their stated scopes; links are not automatic imports.
This guide explains discovery and maintenance, not additional project policy.

## Discovery and routing

| Agent | Project instructions | Planning skill |
| --- | --- | --- |
| Codex CLI | Root `AGENTS.md` is loaded when starting at the root. Codex combines applicable files from the Git root to the working directory, subject to its configured size limit and overrides. | `.agents/skills/sdd-plan/SKILL.md`; invoke `$sdd-plan` or let Codex select it. |
| Claude Code | Root `CLAUDE.md` imports `AGENTS.md`. `.claude/rules/` adds path-scoped routing reminders when matching files are accessed; detailed rules still require a read. | `.claude/commands/sdd-plan.md` routes `/sdd-plan` to the shared skill instructions. |
| OpenCode | Root `AGENTS.md` is the project instruction file; OpenCode also loads applicable user-level instructions. It does not expand ordinary Markdown links. | `.agents/skills/sdd-plan/SKILL.md`; OpenCode discovers the skill and can load it on demand. |

The Claude import also supports versions and sessions without native `AGENTS.md`
support. Native support in Claude Code 2.1.277 and later is conditional on its
project-instruction setting and other instruction files. User/managed exclusions,
disabled skills, trust settings, and custom configuration can change discovery.
The Claude command is a small router to the canonical `.agents/skills/sdd-plan/`
instructions. It avoids a duplicate skill body. A shared-directory symlink was tested,
but OpenCode 1.18.23 warned about duplicate names when scanning both skill paths.
The command avoids that collision; keep its description aligned with the shared skill.

The shared `clean-code` skill follows the same pattern: Codex and OpenCode discover
`.agents/skills/clean-code/SKILL.md`, and Claude's `/clean-code` command routes to it.
For implementation, refactoring, and code review, the root instructions request this
skill alongside the applicable rules; its checklist does not replace those rules.

Claude's scoped reminders are triggered by matching file accesses, not a substitute
for classifying the task at the start. Cross-stack work may affect contracts outside
the paths being edited. Codex and OpenCode do not use `.claude/rules/` as general
project rules. All agents should read the relevant detailed files as directed by
`AGENTS.md`, reuse instructions already in context, and reread when stale or lost.
Keep long-running decisions, unresolved requirements, and verification in the branch's
SDD artifacts; after compaction recover the task state from those artifacts as needed.

## Research and design decisions

Official documentation, checked 2026-10-26:

- [Codex AGENTS.md discovery](https://learn.chatgpt.com/docs/agent-configuration/agents-md)
  describes global/project precedence, per-directory overrides, and the size cap;
  [Codex skills](https://learn.chatgpt.com/docs/build-skills) describes metadata-first
  discovery and explicit or implicit activation. Description lists have a context budget.
- [Claude Code memory](https://code.claude.com/docs/en/memory) documents imports,
  conditional path rules, `AGENTS.md` fallback and exclusions. Its
  [skills documentation](https://code.claude.com/docs/en/skills) documents on-demand
  skills and legacy project commands with the same invocation and frontmatter model.
- [OpenCode rules](https://docs.opencode.ai/docs/rules/) documents `AGENTS.md`,
  `CLAUDE.md` fallback and user-level instructions; [OpenCode skills](https://docs.opencode.ai/docs/skills/)
  documents both `.agents/skills/` and `.claude/skills/` discovery and on-demand loading.
- The [Agent Skills specification](https://agentskills.io/specification) defines
  portable `SKILL.md` frontmatter and naming constraints. The planning skill already
  has a specific task trigger and a matching directory name.

Engineering recommendations, rather than guarantees: [OpenAI's skill guidance](https://developers.openai.com/blog/rethinking-skills-and-prompts-for-gpt-6-astra)
recommends narrow descriptions and progressive disclosure. [Anthropic's context guidance](https://www.anthropic.com/research/effective-context-engineering-for-ai-agents)
discusses context relevance, compaction, and durable notes. We expect a shared root,
small path reminders, and one skill body to reduce duplication and make routing more
reliable; those outcomes require agent evaluations and are not established by file size.

We did not add nested `AGENTS.md` files: OpenCode documents first-match project
instruction discovery, so a nested file could displace the shared root there. Claude's
scoped reminders instead supplement the shared entry point when files are accessed.

No hook is required to read instructions, and a prompt is not deterministic policy
enforcement. Prefer existing CI tests, lint, and review for checkable rules; introduce
hooks only for a specific demonstrated need, accounting for each agent's trust and
execution model. Do not install automatic device checks or rely on hooks as the sole gate.

## Maintenance and verification

- When a demonstrated recurring workflow merits a skill, check existing skills and
  documentation first. Keep the canonical skill in `.agents/skills/<name>/SKILL.md`;
  use a specific name and description that say when it applies, and keep the body
  concise. Put substantial conditional detail in linked references. If a skill-creation
  workflow is available, follow it. Preserve existing policy in its authoritative home.
- For a new shared skill, add a small `.claude/commands/<name>.md` router with a matching
  description and a path to the canonical skill. Check its metadata, links and discovery
  in fresh agent sessions; report any unverified behavior. Do not duplicate skill bodies.
- Preserve the proposed status and authority of the constitutions. Edit a rule at its
  canonical location and keep routing reminders concise; do not duplicate policy bodies.
- Use realistic backend, frontend, cross-stack, planning, and documentation prompts
  to check loading and adherence after changes. A discovery check alone does not prove
  that an agent followed an instruction. Test in fresh sessions and record agent version,
  configuration, prompt, observed behavior, and any disabled integrations.
- Confirm local links and the Claude command's canonical skill path resolve. OpenCode's
  `opencode debug skill` lists skills; check for exactly one `sdd-plan` and no duplicate
  warning, and repeat for `clean-code`. Codex exposes
  skill metadata in its session; Claude offers `/skills` and `/memory` in an interactive
  session. Restart agents when configuration changes are not reflected in a running session.
- Compare `wc -c AGENTS.md` and skill metadata when managing instruction growth. Bytes
  are only a size proxy, not token counts or measured usage. Measure actual provider
  token usage and repeated-read behavior before claiming a saving.
- Review user-level instructions separately. For example, OpenCode can load global
  `instructions` in addition to project guidance; Claude user/managed settings and
  `CLAUDE.local.md` can affect behavior. Do not overwrite personal settings as part of
  repository maintenance.

Initial audit: the repository had one root `AGENTS.md`, one shared planning skill,
three manually routed engineering-rule files, and no project Claude or OpenCode
configuration. The parent and component constitutions are proposed, not ratified.
The root instruction file grew from 2,898 to 3,099 bytes in this migration. This change adds
the Claude import, scoped reminders, and planning command router without duplicating policy.
At audit time, the user's OpenCode configuration explicitly included
two global instruction files; Claude also had a user-level `CLAUDE.md`. Those personal
instructions may affect behavior independently of repository guidance and were not edited.
