# Spec-driven development

Read the project constitution and affected component rules first. They are proposals
for adoption; use this workflow to prepare reviewable changes without claiming approval.

## Hierarchy

| Authority | Document | Purpose |
| --- | --- | --- |
| 1 | [Project constitution](../../.specify/memory/constitution.md) | Principles/governance |
| 2 | [Backend](../../backend/CONSTITUTION.md), [frontend](../../frontend/CONSTITUTION.md) | Scoped additional rules |
| 3 | `specs/<id>-<name>/spec.md` | Outcomes/acceptance |
| 4 | `plan.md`, optional `contracts/`, ADRs | Design/compatibility/compliance |
| 5 | `tasks.md` | Implementation/verification sequence |
| Context | [Map](repository-map.md), setup/architecture/agent guides | Navigation; no policy overrides |

Scope follows affected behavior, not only edited paths. ADRs record decisions; they
do not authorize violations. Parent governance defines conflicts and exceptions.

## Workflow

1. Inspect affected code/tests with the map; record current behavior/gaps. README
   examples may lag actual contracts.
2. Create `specs/<unique-id>-<name>/spec.md` using the
   [spec template](../../.specify/templates/spec-template.md): stable IDs, prioritized
   scenarios, resolved questions before dependent work.
3. Write `plan.md` using the [plan template](../../.specify/templates/plan-template.md):
   design/interfaces, traceability, constitutional checks; use an
   [ADR](../../.specify/templates/adr-template.md) for significant decisions.
   Review spec/plan before implementation; record only actual reviewer approval.
4. Write `tasks.md` using the [tasks template](../../.specify/templates/tasks-template.md):
   prerequisites, coverage, implementation, verification. Independent work requires
   resolved requirements/dependencies.
5. Implement; demonstrate regressions before fixes where feasible. Update artifacts
   for scope/design changes and rerun affected checks.
6. Review acceptance/rules against the final diff; record results/risks. Unapproved
   violations block merge under parent governance.

Scale artifact length to the change. NH-01 permits lightweight documentation/mechanical
work; small security/persistence/protocol/device-control diffs still need relevant analysis.

## Completion and history

Complete means acceptance evidence, verified compatibility/recovery, accurate docs,
and truthful checks. Failed/unrun checks need cause and consequence. [CI](../ci.md)
verifies tests/lint/builds/races/deployment tooling/images; SDD acceptance and
constitutional review are not automatically enforced. Required status checks need
GitHub branch protection/rulesets.
Keep completed specs as history; follow-ups use new linked specs. Amend active specs
openly, preserving IDs. Keep unspecified backlog in issues/`TODO.md`.

## Tooling and document maintenance

Agent guidance lives under `.agents/`: `rules/` for coding conventions routed by
root `AGENTS.md`, `skills/` for workflows. Rules do not load automatically.

For implementation plans, use `$sdd-plan` or its
[instructions](../../.agents/skills/sdd-plan/SKILL.md). It handles active/missing specs
and distinguishes planning, review, and implementation. `.agents/skills/` enables
Codex discovery; templates remain authoritative.

Structure/location follow [GitHub Spec Kit](https://github.github.com/spec-kit/) and its
[constitution template](https://github.com/github/spec-kit/blob/main/templates/constitution-template.md);
component hierarchy/IDs/gates are project policy. These are project-owned templates,
not an installed `/speckit.*` runtime. Copy/fill manually; later tooling must preserve them.

Keep one authoritative home per rule; reference IDs instead of repeating policy.
Load the parent, affected components, and task-specific artifacts; use the map as needed.
Preserve obligations, exceptions, evidence, and known gaps when shortening text.
Use concise prompts and remove unused placeholders before review.
AGENTS.md follows official [guidance](https://learn.chatgpt.com/docs/agent-configuration/agents-md)
and [best practices](https://learn.chatgpt.com/guides/best-practices): concrete commands,
constraints/completion, and links to detailed task guidance.
