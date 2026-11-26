# Tasks: Automation run history and duplication

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: draft
All tasks are future work; this branch contains planning artifacts only.

## Decisions

- [ ] T001 — Decide history durability (in-memory vs persisted) in plan Q1; IDs: AC-07/NFR-01;
  done: decision recorded in plan with rationale.
- [ ] T002 — Approve WS request/event contract in plan Q2; IDs: FR-03/AC-08; done: contract examples.
- [ ] T003 — Decide duplicate identity model (keep one-per-device vs add `deviceId` routing) with an
  ADR; IDs: FR-05/AC-10/11; done: ADR or explicit decision to drop duplicate.

## Backend (after T001/T002)

- [ ] T004 — Add `RunRecord` + bounded `RunRecorder` in `internal/automations/run_history.go`;
  IDs: FR-01/02/NFR-01; depends: T001; done: cap/clear/race tests.
- [ ] T005 — Record success/failure/manual at `Device.Evaluate` and `EvaluateTrigger`; IDs:
  AC-01–03/NFR-02; depends: T004; done: tests for all origins and error truncation.
- [ ] T006 — Wire WS `loadAutomationHistory`, `automationHistory`, `automationRun`; IDs: FR-03/
  AC-08; depends: T002/T005; done: event-hub contract tests.
- [ ] T007 — Clear-history action and confirmation contract; IDs: AC-06; depends: T006; done:
  authenticated clear test.

## Frontend (after T006)

- [ ] T008 — `useAutomationHistory` service/composable with bounded live subscription; IDs:
  FR-04/AC-08/09; done: unit test + unmount cleanup.
- [ ] T009 — `AutomationHistoryDialog.vue` + row History action; IDs: AC-01–06; done: desktop/mobile
  browser spec, empty/error/clear states.
- [ ] T010 — Confirm no Duplicate action renders until T003 approves a model; IDs: AC-11; done:
  browser assertion.

## Verification

- [ ] T011 — Backend unit/race tests for recorder and WS; gate: NH-05.
- [ ] T012 — Browser specs desktop/mobile for history dialog and live append.
- [ ] T013 — Update spec/plan/tasks with actual results; record unrun checks.

## If duplicate is approved (T003 option 2)

- [ ] T014 — ADR for automation identity/routing; review spec 068 guard impact.
- [ ] T015 — Add `deviceId` to model + storage migration/rollback; IDs: AC-10.
- [ ] T016 — Route `HandleDevice` by device index; update manual/scheduler lookups.
- [ ] T017 — Editor draft identity and Duplicate action with bounded copy naming.
