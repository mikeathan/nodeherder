# Frontend Constitution

Version: 1.0.0 | Status: Proposed for adoption | Last Amended: 2026-10-01
Parent: [NodeHerder v1.0.0](../.specify/memory/constitution.md)
Scope: `frontend/` and browser-facing contracts/behavior.

Parent governance applies; these additional rules do not certify existing compliance.

## FE-01 — State/service ownership

Components MUST own presentation/interactions; shared state MUST use established Vuex
modules and reusable behavior MUST use services/composables. New components MUST NOT
duplicate shared transport logic or create competing device truth. Replacing Vuex,
routing, or the component system requires an ADR.

## FE-02 — Browser contracts

Changed API/WS handlers MUST match backend payloads/errors and use `src/contracts/`
and `src/types/` for data shapes. External data MUST be checked at ingress where it
can corrupt state/actions; TypeScript alone is not runtime validation. Contract changes
MUST verify producers/consumers, including mocks used by affected development/test flows.

## FE-03 — Observable network state

Controls MUST distinguish requested actions from confirmed state. Changed flows MUST
specify applicable loading/empty/error/offline/reconnect behavior. Retries MUST NOT
silently duplicate device commands. Charts MUST preserve units/timezones and distinguish
missing samples from measured zero.

## FE-04 — Accessible, consistent interaction

Changed controls MUST have accessible names, keyboard operation, and visible focus;
status/errors MUST NOT rely on color alone. Use suitable shared controls/theme tokens.
Layout changes MUST verify mobile/desktop widths specified in the plan. Destructive
interactions MUST explain scope and provide NH-06 confirmation/recovery.

## FE-05 — Sensitive operations

Provider/service secrets MUST NOT enter `VITE_*` or bundles. Changed authenticated
calls MUST use established API/auth helpers. Assistant text MUST remain untrusted:
no arbitrary HTML/scripts or direct conversion to device commands. Household data
MUST NOT enter browser persistence/logs without a specified need and review.

## Verification

Toolchains/dependencies: [`.nvmrc`](../.nvmrc), [package.json](package.json), lockfile.
From `frontend/`: `npm ci` when needed; `npm test -- --runInBand`; relevant source:
`npm run lint`; implementation: `npm run build` (includes TypeScript checks).
`npm run type-check` watches; do not use it as a terminating review check.
UI changes require recorded keyboard/failure/viewport verification. Jest transforms
TypeScript; Vue mounting tests are not assumed configured. Record actual results in the plan.
