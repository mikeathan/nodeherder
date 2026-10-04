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

`.agents/rules/` loads through these links; `.agents/skills/` holds discoverable
workflows. Read only relevant files; constitutions retain policy authority.

## Planning requests

For new, updated, or reviewed implementation plans, use
[$sdd-plan](.agents/skills/sdd-plan/SKILL.md). Reuse the relevant active spec or create
a unique feature directory; follow the plan template. Planning-only requests stop at
the requested artifacts; do not start implementation. If skills are unavailable,
read the linked skill as workflow guidance.

## Commands

- Backend (`backend/`): `go test ./...`; `go build ./...` for implementation changes.
- Frontend (`frontend/`): `npm ci` when needed; `npm test -- --runInBand`;
  `npm run lint` and `npm run build` for source changes; `npm run dev` for local UI.
- Toolchains: `backend/go.mod`, `.nvmrc`. Additional checks: component constitutions.

## Done

Behavior changes need spec → plan → tasks before implementation; templates:
`.specify/templates/`. Documentation/mechanical changes use NH-01's lightweight path.
Verify acceptance/rules; report actual checks, failures, and unrun checks.
Never run routine tests against household devices. Verify contracts in code; README
examples may lag. Keep guidance concise; link to existing rules rather than repeat them.
