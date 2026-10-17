# Plan: Panel navigation reliability

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-17
Constitutions: parent/frontend v1.0.0 proposed; backend N/A.

## Design and compatibility

Add a pure normalizeRoomIndex helper to domain/panel.ts, used before indexing groups.
Keep the existing saved room key and modulo behavior for valid integer values.
PanelMode defines dock sizing variables used by both dock CSS and the notice bottom
clearance/max-height. StatusBanners consumes optional inherited CSS variables, retaining
its existing defaults elsewhere. A wrapper keeps its fragment roots in one grid row.
Notices stay fixed above the dock, preserving the established no-content-shift contract.
No measured height state, observers or independently hardcoded clearance values.
The initial in-flow proposal failed the existing no-layout-shift browser regression;
revised the design rather than weaken that assertion.

Scope room-tab keyboard events to the tablist, prevent handled defaults, then focus
and reveal the selected tab after Vue updates. Global shortcuts ignore already-handled,
modified and interactive/modal events. Native focus and ARIA associations preserve
selection clarity. Responsive clock sizing/nowrap replaces accidental wrapping.
No storage/wire changes, dependencies or new timers/listeners. Existing key listener
cleanup retained. No significant boundary/dependency decision requires ADR.

## Constitution Check

| Rule | Design evidence | Before implementation | Final |
| --- | --- | --- | --- |
| NH-01 | Reproductions and AC-01–05 first | PASS | PASS |
| NH-02 | Presentation/navigation only; no commands | PASS | PASS |
| NH-03 | Validate stored index before indexing | PASS | PASS |
| NH-04 | Domain helper, panel state, shared notice presentation | PASS | PASS |
| NH-05 | Companion units and real mock browser regressions | PASS | PASS |
| NH-06 | Safe fallback, bounded notices, no new workers | PASS | PASS |
| FE-01 | Existing state/composables remain authoritative | PASS | PASS |
| FE-02 | No external contracts changed | PASS | PASS |
| FE-03 | Offline notice retry/dismiss preserved | PASS | PASS |
| FE-04 | Roving tabs, modal isolation, 320/390/1280px | PASS | PASS |
| FE-05 | Same browser-local preference; no secrets/HTML | PASS | PASS |

## Verification and completion

AC-02 unit regressions in domain/panel.test.ts; AC-01/03/04 in panel.spec.ts. Preserve
all existing persistence/layout checks. Run all frontend units/lint/build plus panel,
StatusBanners and layout browser checks in an isolated mock copy (4200/4210), then
review before publishing. No household devices. Results (2026-10-17):

- `npm test -- --runInBand --silent`: PASS 32 suites / 254 tests.
- `npm run build`: PASS; existing duplicate DeviceCard naming and chunk-size warnings.
- `npm run lint`: PASS; seven existing script warnings.
- Full `npm run test:e2e`: 90/90 PASS on desktop/mobile, including unchanged connection
  notice no-layout-shift tests, new notice/dock separation, 320px clock, corrupt room
  preference recovery, keyboard focus, modal isolation, pin persistence and navigation.
- Browser checks ran against an isolated copy on ports 4200/4210 with mock-only
  origins/fixture endpoints. Production source was copied unchanged; local Vite allowed
  shared node_modules font assets. User development services were left running.
- Inspected compact dock light/dark and disconnected mobile screenshots; notice ends
  above dock, clock stays on one line, and safe-area spacing is retained. Tested narrow
  320/390px and desktop 1280px plus short 844x390 landscape.
- Before fixes, isolated diagnostics reproduced all three reported bugs. Initial
  in-flow notice implementation passed 32/34 affected browser checks but failed the
  existing desktop/mobile no-layout-shift assertions. Replaced that approach with
  shared CSS geometry; those original assertions are unchanged and now pass.
- `git diff --check`: PASS. Final self-review checked computed notice/dock geometry,
  validation before array access, event/modifier/modal guards, ARIA tab associations,
  scoped watcher/key-listener cleanup and preserved shared notice defaults.
  No measurement observer, dependency or new timer/listener was added.

Added short panel navigation guidance to docs/setup.md. Backend tests were not rerun:
no backend source/contracts changed. Complete 021/022 changes are published together
on fix/compact-panel-toolbar with the user-requested commit and descriptive PR.
