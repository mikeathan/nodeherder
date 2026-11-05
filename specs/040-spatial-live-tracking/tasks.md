# Tasks: Spatial live tracking

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete

Each task maps to requirement/acceptance IDs or a constitutional gate and records its
evidence. Tasks are ordered so the contract exists before consumers depend on it.

Evidence is recorded in [plan.md](plan.md#traceability-and-verification): backend and
frontend unit suites (716 tests), lint, build, `type-check:e2e`, and the desktop/mobile
Playwright suites all pass; real screen-reader behaviour was inspected, not tested.
T021 includes the independent review round: two blockers (inert smoothing control, source
flip on a sampled coordinate) and the emitter ordering hole were found and fixed, with
regression tests; the plan records each correction. A later layout review replaced the
always-on presence card with a device tab plus header chip and removed the now-unused widget
and viewport helper; its browser check now passes on desktop and mobile, including an
assertion that the control row shares one baseline.

## Preparation

- [x] T001 — Record decisions for Q-01–Q-05 in `spec.md`; IDs: NH-01; done: questions
  answered or explicitly deferred with owner.
- [x] T002 — Confirm contract and rule check before code; IDs: NH-04, BE-01, FE-02;
  depends: T001; done: contract reviewed against plan; no open safety/contract question.

## Coverage and implementation

- [x] T003 — Frame model and pure builder tests in `backend/models/devices/spatial_test.go`,
  `backend/internal/services/spatial_frame_test.go`; IDs: FR-01, AC-01, AC-06; depends: T002;
  done: failing tests for incomplete rows, unknown units, inactive slots, unit passthrough.
- [x] T004 — Implement `backend/models/devices/spatial.go` and
  `backend/internal/services/spatial_frame.go`; IDs: FR-01; depends: T003; done: T003 passes.
- [x] T005 — Emit change-driven, rate-capped frames with trailing flush, seed and
  config-update emission, offline/shutdown stop in `backend/internal/services/device_lifetime.go`;
  IDs: FR-01, NFR-01, SC-02; depends: T004; done: `device_lifetime_test.go` passes, automation
  delta unchanged.
- [x] T006 — Add `spatialTargets` event and wire producer in
  `backend/internal/ws/eventhub.go`, `backend/internal/controllers/hub-controller.go`,
  `backend/models/devices/requests.go`; IDs: FR-01, NH-04; depends: T005; done: event constant
  covered and old callback removed.
- [x] T007 — Add `SpatialMappings` read to `backend/models/settings/cache.go`; IDs: FR-01,
  BE-04; depends: T004; done: `cache_test.go` proves a copied mapping under the read lock.
- [x] T008 — Add frame types, runtime validation and event route in
  `frontend/src/types/spatial.ts`, `frontend/src/store/modules/ws/index.ts`,
  `frontend/src/store/modules/hub-state/index.ts`; IDs: FR-02, FE-02; depends: T006; done:
  store tests pass for valid, malformed and unknown-device frames.
- [x] T009 — Add frame/source preference in `frontend/src/domain/spatial/frames.ts`;
  IDs: FR-02, AC-06; depends: T008; done: unit conversion and newest-source tests pass.
- [x] T010 — Add adaptive filter in `frontend/src/domain/spatial/filter.ts`; IDs: FR-03,
  AC-02–AC-05, SC-01; depends: T009; done: jitter, lag, teleport, gap and hold tests pass.
- [x] T011 — Add track association in `frontend/src/domain/spatial/tracks.ts`; IDs: FR-04,
  AC-07, AC-08; depends: T010; done: slot-swap, merge/reverse and retention tests pass.
- [x] T012 — Add zone hysteresis in `frontend/src/domain/spatial/occupancy.ts`; IDs: FR-05,
  AC-09; depends: T011; done: boundary chatter and exclusion-mask tests pass.
- [x] T013 — Add status summary and source labels in `frontend/src/domain/spatial/summary.ts`,
  `view.ts`; IDs: FR-06, AC-10–AC-12; depends: T012; done: precedence and preview-source tests pass.
  The auto-fit viewport helper was removed with the always-on card (see the layout follow-up).
- [x] T014 — Add `frontend/src/composables/spatial/useSpatialTracking.ts`; IDs: FR-02–FR-05;
  depends: T013; done: reset-on-change and source-preference tests pass.
- [x] T015 — Render states and heading in `SpatialCanvas.vue`; IDs: FR-03, AC-05, AC-13;
  depends: T014; done: no raw-slot rendering remains; styles use theme tokens.
- [x] T016 — Add the Presence device tab and header chip in `domain/devices/page.ts`,
  `useDeviceTabs.ts`, `SpatialPresenceChip.vue`; IDs: FR-06, AC-10–AC-12, NFR-05;
  depends: T013; done: tab gating and fallback, headline rules, fallback path tested.
- [x] T017 — Integrate the workspace layout in `SpatialPanel.vue` (sticky status/save,
  view switcher, one control row, canvas hero, mapping last); IDs: FR-06, FR-07, AC-12,
  AC-13; depends: T014–T016; done: panel renders; browser check passes after the alignment fix.
- [x] T018 — Update `SpatialHelp.vue` for tracks, states and non-claims; IDs: FR-08, AC-14;
  depends: T017; done: help text reviewed against AC-14.

## Handoff

- [x] T019 — Update Playwright coverage in `frontend/e2e/devices.spec.ts`; IDs: AC-06,
  AC-10–AC-13, NFR-05; depends: T017; done: desktop and mobile suites pass, including the
  control-row alignment assertion.
- [x] T020 — Run planned checks and record actual results in `plan.md`; gate: NH-05;
  depends: all; done: commands, results, failures and unrun checks recorded.
- [x] T021 — Review the full diff against rules, audit GoDoc/JSDoc, remove dead code; gates:
  NH-01/04/05, BE-02/03, FE-01–05; depends: T020; done: findings listed and resolved or raised.
