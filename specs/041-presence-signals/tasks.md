# Tasks: Presence signals for automations

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete (review pending)

## Preparation

- [x] T001 — Confirm ADR-001 decisions and spec questions; IDs: NH-01; done: questions either
  answered or marked non-blocking.
- [x] T002 — Review processor seam and settings accessor against the plan; IDs: NH-04, BE-01;
  done: no cycle and no new wire contract.

## Coverage and implementation

- [x] T003 — Pure rules in `backend/internal/spatial/rules.go` with tests in `rules_test.go`;
  IDs: FR-02, FR-03, SC-01; done: hysteresis, dedupe, dwell, empty-time tests pass.
- [x] T004 — Derived device/entity builder in `entities.go` with `entities_test.go`; IDs:
  FR-01, FR-04, AC-01, AC-09; done: stable names, read-only metadata, zone removal.
- [x] T005 — Engine and lifecycle in `presence_engine.go` with `presence_engine_test.go`; IDs: FR-07, AC-10,
  NFR-01–NFR-02; done: bounded queue, 250 ms burst guard, 10 s counter flush, Close joins.
- [x] T006 — Settings `SpatialWorkspace` accessor with `cache_test.go` coverage; IDs: FR-01,
  BE-04; done: owned copy, zones only.
- [x] T007 — Wire frames, config updates, device add/remove in `hub-controller.go`; IDs:
  FR-01, AC-01, AC-06; done: derived device appears, syncs and disappears correctly.
- [x] T008 — Hub automation scenario on derived occupancy without a browser; IDs: AC-05,
  SC-02; done: scenario passes.
- [x] T009 — Frontend resolution of derived zone signals and panel labels; IDs: FR-06,
  AC-08, NFR-05; done: domain tests and browser check.
- [x] T010 — Update help/plan evidence; IDs: FR-06; done: copy states the source.
- [x] T011 — Run planned checks and record actual results; gate: NH-05; done: evidence in plan.

## Handoff

- [x] T012 — Review the diff against rules, audit GoDoc/JSDoc and dead code; gates: NH-01/04/05,
  BE-02/03, FE-01–05; done: findings listed and resolved or raised.
- [x] T013 — Evidence recorded in [plan.md](plan.md#traceability-and-verification); suite results
  and the one contention flake noted.
