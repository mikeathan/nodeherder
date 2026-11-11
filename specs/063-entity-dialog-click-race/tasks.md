# Tasks: Entity dialog opens during a hub-state refresh

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete

## Investigation

- [x] T001 — Reproduce the failed leak test and inspect the CI/local trace; IDs: NH-01;
  done: screenshot showed no dialog, trace had no `/metrics/query`, instrumentation logged
  `device=false` in `EntityTile.openHistory`.

## Implementation

- [x] T002 — Remove the transient `!device.value` guard from `EntityTile.openHistory` and
  open by device/expose identity with a title fallback; IDs: FR-01, AC-01, AC-02; done:
  repeated leak test passes.
- [x] T003 — Wait for the dialog and history canvas independently with timeouts above the
  10 s history request budget; IDs: AC-03, NFR-02; done: spec file updated.
- [x] T004 — Confirm history, dialog close and leak assertions are unchanged; IDs: FR-02,
  AC-05; done: dialog desktop/mobile suites pass.

## Verification

- [x] T005 — Repeat the leak test 5×, run the exact failed desktop shard, the mobile dialog
  suite, unit tests, lint, build and e2e type-check; IDs: AC-04, SC-01…SC-03; done:
  5/5 repeats; 59 passed/9 skipped on shard 2/4; 39 mobile; 821 Jest; lint 0 errors.
- [x] T006 — Final diff review; gates: FE-01, FE-02; done: only `EntityTile.vue` and the
  leak test changed, plus the SDD artifacts.
