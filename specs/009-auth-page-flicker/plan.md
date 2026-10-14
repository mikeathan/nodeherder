# Plan: Stable authentication page handoff

Spec: [spec.md](spec.md) | Status: Implemented | Author: Codex | Updated: 2026-10-14
Constitutions: parent/frontend 1.0.0 (proposed)

## Context and design

App.vue owns the standalone login versus shell decision. Require authenticated
state and current route meta.requiresAuth before mounting AppShell. Keep the same
LoginPage instance while navigation loads its destination. Avoid timers, animations,
or session-service changes: those cannot correct the ownership race.
Wait for router.isReady() before mounting App in src/main.ts so the first login
check sees the resolved route and OAuth URL cleanup is not overwritten by initial
navigation. The new shell gate exposed that pre-existing startup race: the old
extra login mount accidentally repeated the OAuth cleanup.
Keep LoginPage in its status presentation after successful validation while
navigation is pending, so sign-in options do not flash for an authenticated user.
Existing router metadata covers all shell routes including panel mode. No dependencies
or significant architecture decisions; no ADR needed. Spec 008 remains historical.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 009 reproduction/acceptance and tasks | PASS | PASS |
| NH-02 | Presentation only; no device/event changes | N/A | N/A |
| NH-03 | No auth authorization/exposure or data changes | PASS | PASS |
| NH-04 | Existing root/Vuex/router ownership retained | PASS | PASS |
| NH-05 | Delayed-route browser regression; Jest/lint/build | PASS (design) | PASS |
| NH-06 | No persistence/resource lifecycle changes | N/A | N/A |
| FE-01 | Root presentation derives existing shared state | PASS | PASS |
| FE-02 | No API/WS handler or shape changes | N/A | N/A |
| FE-03 | Existing checking/error/offline recovery retained | PASS | PASS |
| FE-04 | No control/style changes; verify both viewports and keyboard auth | PASS (design) | PASS |
| FE-05 | No changed authenticated calls/secrets | N/A | N/A |

## Contracts, data, rollout

No changed API/WS/storage/configuration. Frontend-only rollout and rollback; no migration.

## Traceability and verification

| IDs | Paths | Test/procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01/02, NFR-01 | src/App.vue; src/main.ts; src/components/auth/LoginPage.vue; e2e/auth.spec.ts | Hold lazy destination module; record login DOM mounts and mode requests; assert standalone login then destination | PASS: OAuth/cookie, both viewports; one mount/mode check and no sign-in options |
| FR-02, AC-03 | e2e/auth.spec.ts | Existing auth browser tests, both viewports incl. keyboard local sign-in and failure recovery | PASS: 22 auth cases |
| NH-05 | frontend | npm test -- --runInBand; npm run lint; npm run build | PASS: 23 suites / 152 tests, lint zero errors, build succeeds |

Tests use localhost mock hub only, never household devices. Browser cases belong in
existing auth.spec.ts; no new test files. No mounting-tool migration.

## Debt, risks, exceptions

Cached-session startup still follows spec 008's retained identity policy; no change
in when backend rejects an expired cookie. A lazy-route failure remains router-owned.
No exceptions or maintainer approvals claimed.

## Completion

Implemented on `fix/auth-page-flicker`; prepared for commit and pull-request review. Final diff
review confirms presentation-only scope, unchanged contracts, and all AC coverage.
No maintainer review/approval is claimed.

Actual verification (frontend/, Node v24.21.0 via
`PATH=/opt/homebrew/opt/node@24/bin:$PATH`, npm 11.12.1):

- Before the fix, `npm run test:e2e -- --grep 'session handoff' --project desktop`
  failed: navigation appeared while the login route destination was held back.
- The shell-only fix exposed initial-navigation overwrite of OAuth URL cleanup;
  waiting for router readiness resolved it. A subsequent no-sign-in-options assertion
  failed before LoginPage's authenticated handoff status was added.
- `npm run test:e2e -- e2e/auth.spec.ts e2e/layout.spec.ts e2e/panel.spec.ts`:
  26 passed, desktop/mobile. Includes auth failure/recovery, keyboard local sign-in,
  direct main-screen loads with retained auth, responsive layouts and panel exit.
  Node 24's first sandbox attempt could not bind port 4110 (EPERM); the approved
  rerun completed using localhost mock servers only.
- `npm test -- --runInBand`: 23 suites / 152 tests passed. Existing ts-jest
  deprecation and delayed-exit/open-handles warning remain; process exited zero.
- `npm run lint`: passed, seven existing script warnings, zero errors.
- `npm run build`: passed including TypeScript; existing duplicate DeviceCard and
  large-chunk warnings remain.
- `git diff --check`: passed. Inspected pending-handoff screenshots at both widths:
  centered standalone status card, no nested shell or sign-in button.

Initial Node 26 checks also passed; final evidence above uses `.nvmrc`'s Node 24.
Backend checks and remaining device/automation browser suites were not run: no backend
or device behavior changed. No real household devices were contacted.

## Requested cleanup — design before refactoring

Use the sdd-plan skill and Vue rules. Extract page session coordination into
`src/composables/useLoginSession.ts`, retaining Vuex/service ownership. Replace
`started` plus `checking` coordination with one in-flight promise, publish async
results only while mounted, use finally for cleanup, and keep recovery listeners
symmetric. Named computed presentation state replaces inline status conditions.
Companion Jest tests mock services/session/lifecycle and verify success, rejection,
concurrent requests, recovery, and disposal. Simplify browser mount observation by
tracking unique login elements in a Set rather than nested mutation loops/global casts.
No new runtime dependency or network/storage contract; parent/component rule design
checks above still apply. Execution evidence: [quality.md](quality.md); 160 unit tests and 26 browser checks passed.

Measure scoped cyclomatic complexity with the installed TypeScript AST (separate
nested callables), and statement coverage from Jest's Istanbul report. Record
CRAP = C²(1 - statement coverage fraction)³ + C, per callable; target C <= 10,
CRAP <= 30. Vue script baseline has no Jest instrumentation; use zero coverage as
an explicit conservative baseline, not inferred browser coverage. This is a scoped
measurement, not a repository-wide quality gate. No frontend CRAP runner exists.

Cleanup final review: NH-01/04/05 and FE-01/03 remain satisfied by the documented
plan, existing state/service ownership, readonly presentation refs, single-flight
checks, symmetric listener cleanup, and companion regression coverage. Unexpected
session-check rejection now produces an actionable error and settles loading.
No exceptions/approval claimed; checks and metric limitations are in quality.md.
