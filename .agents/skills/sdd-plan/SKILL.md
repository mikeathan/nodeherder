---
name: sdd-plan
description: Create, update, or review NodeHerder SDD implementation plans for features, fixes, or refactors. Use for requests to plan repository changes; excludes non-software plans and implementation-only requests.
---

# SDD planning

Produce a repository-grounded plan within the requested scope. Planning-only requests
authorize planning artifacts, not application edits or implementation.

## Workflow

1. Identify the goal, operation (new/update/review), and affected components. Inspect
   code/tests and existing `specs/`. Reuse an unambiguous active spec; for new work
   choose an unused `specs/<id>-<name>/`. Preserve completed history and unrelated plans.
   Ask only when selecting the target would materially change scope.
2. Read the [parent](../../../.specify/memory/constitution.md), applicable
   [backend](../../../backend/CONSTITUTION.md)/[frontend](../../../frontend/CONSTITUTION.md)
   rules, and [workflow](../../../docs/sdd/README.md). Use the
   [map](../../../docs/sdd/repository-map.md) to locate code as needed.
3. Read the feature spec. If absent, draft one from the request using the
   [spec template](../../../.specify/templates/spec-template.md). Label assumptions
   and unresolved questions; never invent requirements or approval. Continue independent
   planning while flagging questions blocking dependent design/implementation.
4. Create/update `plan.md` using the
   [plan template](../../../.specify/templates/plan-template.md): actual paths,
   ownership, alternatives, contracts, compatibility/recovery, risks, requirement
   traceability, and rule checks. Distinguish supported design from pending execution
   evidence. Add ADRs only for significant decisions; create `tasks.md` when requested
   or proceeding to authorized implementation.
5. Verify links, placeholders, requirement coverage, every NH rule, and applicable
   BE/FE rules. Read check commands from component guidance. Planning alone requires
   no application test suites; proposed checks are not passed checks.

## Handoff

Link artifacts; summarize design, assumptions, blockers, and proposed verification.
Review-only requests return findings without edits. Stop after requested planning;
existing authorization to implement remains valid.
