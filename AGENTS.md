# NodeHerder agent guide

Go smart-home backend (`backend/`); Vue/TypeScript UI (`frontend/`); MQTT/Zigbee2MQTT,
metrics, automations, and assistant/MCP integrations.

## Before editing

Read the [project constitution](.specify/memory/constitution.md) and applicable
[backend](backend/CONSTITUTION.md)/[frontend](frontend/CONSTITUTION.md) rules; both for
cross-stack changes. Use the [SDD workflow](docs/sdd/README.md) for behavior changes
and its [map](docs/sdd/repository-map.md) to locate code/tests. Constitutions are
proposed; do not invent ratification or exception approval.

Classify implementation/refactoring/review before loading rules:

- Backend → [Go](.agents/rules/go-engineer.md).
- Frontend → [Vue/TypeScript](.agents/rules/vue-engineer.md).
- Cross-stack → both plus [integration](.agents/rules/fullstack-engineer.md).

These links do not automatically load rules: read the relevant files when needed.
Reuse instructions already in context; reread if they may have changed or were lost
after compaction. Constitutions retain policy authority. See the
[agent configuration guide](docs/sdd/agent-configuration.md) for discovery and setup.

For code implementation, refactoring, or review, activate `clean-code` (Claude Code:
`/clean-code`; Codex: `$clean-code`) or read its
[instructions](.agents/skills/clean-code/SKILL.md) if skill discovery is unavailable.
This complements, and does not replace, the applicable engineering rules.

## Planning requests

For new, updated, or reviewed implementation plans, activate `sdd-plan` (Claude
Code: `/sdd-plan`; Codex: `$sdd-plan`) or read its
[instructions](.agents/skills/sdd-plan/SKILL.md) if skill discovery is unavailable. Follow the
[branch scope rule](docs/sdd/README.md#branch-scope): update the current branch's one
spec/plan/tasks set for follow-up changes; new work after completed earlier work gets
a new spec. Planning-only requests stop at the requested artifacts; do not start
implementation.

## Reusable workflows

When work reveals a demonstrated, recurring repository workflow that benefits from
reusable instructions, capture it in an existing skill or create a focused project
skill. Check for an existing owner first; use ordinary documentation for facts or
one-off findings. Respect the requested scope, including review-only and planning-only
work. Follow the [agent configuration guide](docs/sdd/agent-configuration.md) for
skill discovery and verification.

## Commands

- Backend (`backend/`): `go test ./...`; `go build ./...` for implementation changes.
- Frontend (`frontend/`): `npm ci` when needed; `npm test -- --runInBand`;
  `npm run lint` and `npm run build` for source changes; `npm run dev` for local UI.
- Toolchains: `backend/go.mod`, `.nvmrc`. Additional checks: component constitutions.

## Test file ownership

Test files MUST belong to an identifiable production source file and use that source
file's name. Add regressions to its existing companion test file; do not create separate
feature-, issue-, or regression-named test files (for example, `session_reliability_test.go`
or `auth-session.test.ts`).

- Go: `auth.go` → `auth_test.go` in the same package directory.
- Frontend unit tests: mirror the source path under `src/__tests__/`, for example
  `src/store/modules/auth/index.ts` → `src/__tests__/store/modules/auth/index.test.ts`.

Before adding a test file, identify its source owner and check for an existing companion.

## Done

Behavior changes need spec → plan → tasks before implementation; templates:
`.specify/templates/`. Documentation/mechanical changes use NH-01's lightweight path.
Verify acceptance/rules; report actual checks, failures, and unrun checks.
For code changes and reviews, inspect the relevant diff and check it against the
applicable engineering rules before finishing; report unresolved findings.
Never run routine tests against household devices. Verify contracts in code; README
examples may lag. Keep guidance concise; link to existing rules rather than repeat them.
