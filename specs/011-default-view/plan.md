# Plan: Default view on this browser

Spec: [spec.md](spec.md) | Status: Complete | Author: Codex | Updated: 2026-10-16
Constitutions: parent 1.0.0; frontend 1.0.0 (both proposed)

## Context and design

`router/index.ts` redirects authenticated root visits to Home; `router/navigation.ts`
separately handles post-login defaults. Share startup resolution through a small pure
`domain/default-view.ts` policy with allowlisted route names. A shared composable owns
reactive selection and safeStorage writes, following appearance preference ownership.
Reuse one `DefaultViewButton.vue` beside AppHeader's title and in PanelMode's dock.
Text on desktop, compact title control on mobile; Panel dock retains short labels.
Extract nested post-login branching into tested policy to reduce complexity and avoid
inconsistent defaults. No unrelated refactoring, dependencies or significant boundary
replacement; no ADR needed. Automatic last-page tracking rejected: incidental navigation
would change the preferred startup page. Account storage would prevent per-device defaults.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec acceptance and tasks precede code | PASS | PASS |
| NH-02 | Presentation only, no commands/events | N/A | N/A |
| NH-03 | Only allowlisted route names persisted, auth guards retained | PASS | PASS |
| NH-04 | Existing router/composable/storage boundaries, no API change | PASS | PASS |
| NH-05 | Pure policy tests, mock browser flows, lint/build | PASS (design) | PASS |
| NH-06 | Invalid storage falls back; remove key to recover | PASS | PASS |
| FE-01 | Shared composable like existing appearance preferences | PASS | PASS |
| FE-02 | No API/WS changes | N/A | N/A |
| FE-03 | No network/device operations; storage failure remains navigable | PASS | PASS |
| FE-04 | Shared button styling, pressed state, keyboard, two widths | PASS (design) | PASS |
| FE-05 | Presentation route enum only, no household data | PASS | PASS |

## Contracts, data, rollout

New optional `nodeherder_default_view` string. Read validates membership; absent or
invalid values mean Home. Existing preferences untouched. Old clients ignore the key;
rollback leaves an unused preference. No server migration/deployment ordering.

## Traceability and verification

| IDs | Paths | Verification | Result |
| --- | --- | --- | --- |
| FR-01, AC-01–04 | domain/default-view.ts, router/index.ts, router/navigation.ts, composables/useDefaultView.ts | Companion policy tests and mock Panel pin/revisit/auth flows | PASS |
| FR-02, AC-05/06, NFR-01 | components/shell/AppHeader.vue, components/panel/PanelMode.vue, components/ui/DefaultViewButton.vue | Mock browser keyboard, mobile/desktop overflow, screenshots | PASS |
| SC-01 | frontend | Jest, lint, build; targeted Playwright | PASS |

## Debt, risks, exceptions

Keep cleanup in touched navigation; replace nested auth resolution with a tested pure
function. No frontend CRAP tooling found: report complexity/coverage evidence without
claiming a measured CRAP reduction. Existing backend metrics are outside this change.
Storage can be blocked or cleared: tab-memory fallback, Home on next visit.
No exceptions or maintainer approvals claimed.

## Completion

Implementation and verification complete. All automated UI checks use the
mock hub, with no household-device traffic. No backend source changes: Go checks N/A.

Verification (2026-10-16):
- Baseline policy regression test failed because the requested module did not exist.
- `npm test -- --runInBand --coverage --coverageDirectory=/private/tmp/nodeherder-default-view-coverage --collectCoverageFrom=src/domain/default-view.ts --collectCoverageFrom=src/composables/useDefaultView.ts`:
  29 suites / 219 tests passed; new domain and composable 100% statements, branches,
  functions and lines. An intermediate companion-test compile error was corrected by
  using dynamic import instead of untyped CommonJS require.
- `npm run lint`: passed, seven existing warnings in version/start scripts.
- `npm run build`: passed; existing duplicate DeviceCard name and large chunk warnings.
- `git diff --check`: passed.
- Browser suite initially passed 30 tests. Screenshot review found dock hover contrast
  and prompted a fix. The new contrast assertion initially sampled the CSS transition;
  it now polls for its final color. Final rerun: 30/30 passed at 390×844 and 1280×800.
- Screenshot review: title pin fits desktop/mobile; Panel dock has visible focus and
  pressed state. Mobile dock spacing tightened for four labeled controls.
- Execution environment: installed frontend dependencies, Node 24 for Playwright, local
  mock hub on 4110 and Vite on 4100. Sandbox EPERM on server binding required an
  approved rerun outside the sandbox; no real services used.
- Full unrelated browser suite not run: targeted auth, layout and Panel coverage covers
  changed routes and controls. No numeric CRAP score claimed (frontend tooling absent).

Rule review: route enum only in persistence, allowlist checked on read and mutation;
authentication guard retained; deep links and post-login redirects precede defaults;
no API/device semantics changes. No exceptions or ratification claimed.

Final browser command: `PATH=/opt/homebrew/opt/node@24/bin:$PATH npm run test:e2e -- e2e/panel.spec.ts e2e/layout.spec.ts e2e/auth.spec.ts` (34.7s). Screenshots under `frontend/test-results/panel-browser-default-view-28457-aces-and-clears-the-default-{desktop,mobile}/` reviewed after the final styling fix.
