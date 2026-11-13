# Tasks: Behavioural classification and a specialised viewer

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: draft

Each task maps to requirement/acceptance IDs or a constitutional gate and names its evidence.
Mark done only with evidence. No task is started.

## Preparation

- [x] T001 — **DONE**: ADR-001, ADR-002 and ADR-003 recorded. Q1 agreed (neighbour is
- [x] T002 — **DONE**: schema confirmed and the ceiling measured rather than estimated. A

## Backend

- [x] T003 — **DONE**: `evidence_test.go`, 11 tests. Hour mask, Welford over 10k readings, mobility thresholds, write floor, day reset, unusable readings.
- [x] T004 — **DONE**: `evidence.go`. RSSI 0 counts as activity but not as signal evidence, because 0 dBm means unmeasured, not zero distance.
- [x] T005 — **DONE**: `classify_test.go`, 12 tests. Hysteresis is measured against real scores rather than asserted.
- [x] T006 — **DONE**: `classify.go`. Per-day scoring then recency weighting. Mobility counted in dB, not distance bands.
- [x] T006a — **DONE**: `provisional_test.go`, 5 tests.
- [x] T006b — **DONE**: `provisional.go`. `Provisional.asClassified` is the single bridge to the classifier's type.
- [ ] T007 — Red tests for the summaries store: write floor, retention, `Prune`, absence of a
  file when there are no observations, and that **transitions are pruned on the same 30-day
  boundary as summaries**. Paths: `internal/observations/summaries/*_test.go`.
  IDs: FR-07, FR-08, FR-09, NFR-01.
- [ ] T008 — Implement `internal/observations/summaries`, mirroring `spatialhistory`:
  repository interface, bbolt store, worker-serialised recorder, bounded queue that drops.
  IDs: FR-07, FR-08, NFR-01; depends: T007; done: green, `-race` clean.
- [ ] T009 — Wire the summariser into the service and add the additive snapshot fields.
  IDs: FR-01, FR-02, FR-04; depends: T004, T006, T008; done: snapshot carries class and
  evidence; transitions recorded on change only.
- [ ] T010 — Add the summary history route, protected, bounded. IDs: AC-10; depends: T008;
  done: route test with a bounded response.
- [ ] T011 — Triage override and class advisory precedence. IDs: FR-05, AC-09; done: snapshot
  shows both.

## Frontend

- [ ] T012 — Contracts and types for the new optional fields; mock fixtures covering each
  class plus learning. IDs: FE-02; done: malformed and unknown-class cases rejected.
- [ ] T013 — Tracked/Census split with the default filter; "N hidden" footer. IDs: FR-09,
  AC-15; done: default view hides infrastructure and neighbour.
- [ ] T014 — Class presentation: word, confidence, "why" phrase, transition note, and a
  distinct **provisional** treatment that states the class is not settled. IDs: AC-07, AC-10,
  AC-22; done: no bare badges anywhere and a provisional class is never shown as settled.
- [ ] T015 — Timeline pane, desktop grid and mobile horizontal scroll with a sticky identity
  column. IDs: AC-15; done: renders a day per identity.
- [ ] T016 — Radar: class colours, worded legend, labels for the tracked set. IDs: FE-04;
  done: colour is never the only cue.

## Handoff

- [ ] T017 — Browser verification at 1280 px and 390 px, keyboard traversal recorded.
  IDs: AC-16; gate: FE-04.
- [ ] T018 — Run all planned checks and record actual results. Gate: NH-05.
- [ ] T019 — Final rule audit against the full diff; report unresolved findings.
  Gates: NH-01, NH-04, NH-05.

Add discovered work here and update the spec/plan when scope changes.
