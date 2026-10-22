# Tasks: Restore automation toggle commands

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Inspect regression, read applicable rules/skill, create spec/plan/tasks
  before production edits; NH-01/04. Evidence: current branch starts at b60d2bdb;
  editor drops backend values.toggle; specification above.
- [x] T002 — Add failing companion domain regressions for native options, boundary
  metadata, strict readings, action summaries; FR-01/02, AC-02/04; depends T001.
- [x] T003 — Implement domain helpers and action-only editor opt-in; FR-01,
  NFR-01; depends T002. Done: passing focused Jest and typed source checks.
- [x] T004 — Reuse hubharness/production routes in opt-in synthetic browser hub;
  add real-engine UI-create/save/reload/two-event E2E in automations.spec.ts and
  managed Playwright/CI setup; AC-01/02/03, NFR-02; depends T003.
- [x] T005 — Run full frontend/backend checks, affected race checks and desktop/
  mobile browser suite; AC-04/05, NH-05; depends T004. Record exact outcomes.
- [x] T007 — Correct the zero-delay save contract in `contracts/automations.ts`
  and the shared Vuex save action, with companion tests; AC-06; discovered in T004;
  required before T005 completes.
- [x] T008 — Synchronize the callback registry in `backend/store/store.go` with
  copy-on-registration and unlocked notification; companion regression tests;
  AC-07, BE-03; discovered during T005; rerun full/race/backend and browser checks.
- [x] T006 — Review diff, fixture cleanup, failures, rules and compatibility;
  update plan evidence; report branch/checkout handoff in the final response; NH-01–06;
  depends T005. Do not claim unrun checks or live-device verification.
- [x] T009 — Preserve empty/replaced trigger lists atomically in
  `backend/internal/automations/trigger.go`; source-owned regression tests and
  UI empty-save/reload coverage; AC-08; rerun affected full/race/browser checks.

Completion evidence: [plan verification](plan.md#traceability-and-verification).
T002/T008/T009 reproduced failures before fixes; T003/T007 have passing companion
and real-engine save tests; T004/T005 passed full component/browser checks and final
race-instrumented repeated browser cases; T006 includes final diff/rule self-review.
